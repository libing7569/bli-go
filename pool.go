package pool

import (
	"fmt"
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
	gNum        int
	scaleStep   int
	totalTask   int32
	taskQueue   chan GTask
	resultQueue chan GResult
	wg          sync.WaitGroup
}

const (
	TASK_QUEUE_MAX_SIZE   = 100000
	RESULT_QUEUE_MAX_SIZE = 2
	MAX_GOROUTINE_NUM     = 60
	MIN_GOROUTINE_NUM     = 5
	MONITOR_INTERVAL      = 1
	SCALE_CHECK_MAX       = 5
)

func NewGPool(gNum int) *GPool {
	taskQueue := make(chan GTask, TASK_QUEUE_MAX_SIZE)
	resultQueue := make(chan GResult, RESULT_QUEUE_MAX_SIZE)
	if gNum < MIN_GOROUTINE_NUM {
		gNum = MIN_GOROUTINE_NUM
	}
	return &GPool{gNum: gNum, taskQueue: taskQueue, resultQueue: resultQueue}
}

func (pool *GPool) AddTask(task func(...interface{}) interface{}, args ...interface{}) {
	id := uuid.NewV4().String()
	//fmt.Printf("add %v start\n", id)
	pool.taskQueue <- GTask{id: id, task: task, args: args}
	//fmt.Printf("add %v stop\n", id)
}

func (pool *GPool) scale() {
	timer := time.NewTicker(MONITOR_INTERVAL * time.Second)
	var scale int16

	for {
		select {
		case <-timer.C:
			fmt.Printf("total tasks: %v\n", len(pool.taskQueue))
			switch {
			case len(pool.taskQueue) >= 2*pool.gNum:
				scale++
				if pool.gNum*2 <= MAX_GOROUTINE_NUM && scale > SCALE_CHECK_MAX {
					pool.gNum *= 2
					pool.incr(pool.gNum)
					scale = 0
				}

			case len(pool.taskQueue) < pool.gNum/2:
				scale--
				if pool.gNum/2 >= MIN_GOROUTINE_NUM && scale < -1*SCALE_CHECK_MAX {
					pool.gNum /= 2
					pool.incr(pool.gNum)
					scale = 0
				}

			default:
				fmt.Println("no scale!")
			}
		}
		fmt.Printf("pool Goroutine num: %v\tscale: %v\n", pool.gNum, scale)
	}
}

func (pool *GPool) incr(num int) {
	for i := 0; i < num; i++ {
		pool.wg.Add(1)
		go func() {
			defer func() {
				fmt.Println("closed")
				pool.wg.Done()
			}()

			for task := range pool.taskQueue {
				//fmt.Printf("take out %v\n", task.id)
				res := task.task(task.args...)
				pool.resultQueue <- GResult{id: task.id, res: res}
			}
		}()
	}
}

func (pool *GPool) Start() {
	pool.incr(pool.gNum)
	go pool.scale()
}

func (pool *GPool) Stop() {
	close(pool.taskQueue)
	pool.wg.Wait()
	close(pool.resultQueue)
}

func (pool *GPool) GetResult() {
	go func() {
		for _ = range pool.resultQueue {
			//fmt.Printf("Taskid: %v\tTaskresult: %v\n", result.id, result.res)
			atomic.AddInt32(&pool.totalTask, 1)
		}
	}()
}

//func (pool *GPool) Heatbeat() {
//go func() {
//for {
//time.Sleep(1 * time.Second)
//fmt.Printf("taskQueue length: %v, resultQueue length: %v\n", len(pool.taskQueue), len(pool.resultQueue))
//}
//}()
//}
