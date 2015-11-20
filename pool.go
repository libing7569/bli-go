package pool

import (
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/satori/go.uuid"
)

type GResult struct {
	id  string
	res interface{}
}

type GTask struct {
	id   string
	args []interface{}
	task func(args ...interface{}) interface{}
}

type GPool struct {
	graceful    chan byte
	gSugNum     int32
	gCurNum     int32
	totalTask   int32
	taskQueue   chan GTask
	resultQueue chan GResult
	gchans      []chan byte
	wg          sync.WaitGroup
}

const (
	TASK_QUEUE_MAX_SIZE   = 100000
	RESULT_QUEUE_MAX_SIZE = 2
	MAX_GOROUTINE_NUM     = 1024
	MIN_GOROUTINE_NUM     = 5
	MONITOR_INTERVAL      = 1
	SCALE_CHECK_MAX       = 5
)

var logger = log.New(os.Stdout, "debug: ", log.LstdFlags|log.Lshortfile)

func NewGPool(gSugNum int32) *GPool {
	taskQueue := make(chan GTask, TASK_QUEUE_MAX_SIZE)
	resultQueue := make(chan GResult, RESULT_QUEUE_MAX_SIZE)
	gracechan := make(chan byte)
	if gSugNum < MIN_GOROUTINE_NUM {
		gSugNum = MIN_GOROUTINE_NUM
	}
	return &GPool{gSugNum: gSugNum, taskQueue: taskQueue, resultQueue: resultQueue, graceful: gracechan}
}

func (pool *GPool) AddTask(task func(...interface{}) interface{}, args ...interface{}) {
	id := uuid.NewV4().String()
	//logger.Printf("add %v start\n", id)
	pool.taskQueue <- GTask{id: id, task: task, args: args}
	//logger.Printf("add %v stop\n", id)
}

func (pool *GPool) scale() {
	timer := time.NewTicker(MONITOR_INTERVAL * time.Second)
	var scale int16

	for {
		select {
		case <-timer.C:
			logger.Printf("gCurNum: %v\ttotal tasks: %v\n", pool.gCurNum, len(pool.taskQueue))
			switch {
			case len(pool.taskQueue) >= int(2*pool.gCurNum):
				if pool.gCurNum*2 <= MAX_GOROUTINE_NUM && scale >= SCALE_CHECK_MAX {
					//pool.gCurNum *= 2
					pool.incr(pool.gCurNum)
					scale = 0
				} else if scale < SCALE_CHECK_MAX {
					scale++
				}

			case len(pool.taskQueue) <= int(pool.gCurNum/2):
				if pool.gCurNum/2 >= MIN_GOROUTINE_NUM && scale <= -1*SCALE_CHECK_MAX {
					logger.Printf("desc action - gNum: %v\n", pool.gCurNum)
					pool.desc(pool.gCurNum / 2)
					scale = 0
				} else if scale > -1*SCALE_CHECK_MAX {
					scale--
				}

			default:
				logger.Println("no scale!")
			}
		}
		logger.Printf("pool Goroutine num: %v\tscale: %v\n", pool.gCurNum, scale)
	}
}

func (pool *GPool) desc(num int32) {
	logger.Printf("desc info - num: %v\t gchansLen: %v\n", num, len(pool.gchans))
	if int(num) > len(pool.gchans) {
		return
	}
	for _, c := range pool.gchans[:len(pool.gchans)-int(num)] {
		c <- 1
	}
	pool.gchans = pool.gchans[len(pool.gchans)-int(num):]
}

func (pool *GPool) incr(num int32) {
	for i := 0; i < int(num); i++ {
		//logger.Printf("!add %v goroutine\n", i)
		ch := make(chan byte)
		pool.gchans = append(pool.gchans, ch)
		//id := uuid.NewV4().String()
		pool.wg.Add(1)
		go func(mych <-chan byte) {
			atomic.AddInt32(&(pool.gCurNum), 1)
			//logger.Printf("the goroutine %v\n", mych)
			defer func() {
				logger.Println("closed")
				atomic.AddInt32(&pool.gCurNum, -1)
				pool.wg.Done()
			}()

		l:
			for {
				select {
				case task, ok := <-pool.taskQueue:
					if ok {
						res := task.task(task.args...)
						pool.resultQueue <- GResult{id: task.id, res: res}
					} else {
						break l
					}
				case cmd := <-mych:
					switch cmd {
					case 1:
						logger.Println("receive stop cmd!")
						break l
					}
				}
			}
			logger.Println("goroutine end!")
		}(ch)
	}
}

func (pool *GPool) Start() {
	pool.incr(pool.gSugNum)
	go pool.scale()
}

func (pool *GPool) Stop() {
	close(pool.taskQueue)
	logger.Println("close taskQueue")
	pool.wg.Wait()
	close(pool.resultQueue)
	logger.Println("close resultQueue")
	<-pool.graceful
}

func (pool *GPool) GetResult(f func(res GResult)) {
	go func() {
		for r := range pool.resultQueue {
			f(r)
			atomic.AddInt32(&pool.totalTask, 1)
		}
		pool.graceful <- 1
	}()
}

func (pool *GPool) Heatbeat() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			logger.Printf("gCurNum: %v, taskQueue length: %v, resultQueue length: %v\n", pool.gCurNum, len(pool.taskQueue), len(pool.resultQueue))
		}
	}()
}
