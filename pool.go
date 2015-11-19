package pool

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/satori/go.uuid"
)

type GResult struct {
	Id  string
	Res interface{}
}

type GTask struct {
	Id   string
	Args []interface{}
	Task func(args ...interface{}) interface{}
}

type GPool struct {
	gNum        int
	totalTask   int32
	taskQueue   chan GTask
	resultQueue chan GResult
	wg          sync.WaitGroup
}

const (
	TASK_QUEUE_MAX_SIZE   = 2
	RESULT_QUEUE_MAX_SIZE = 2
)

func NewGPool(num int) *GPool {
	taskQueue := make(chan GTask, TASK_QUEUE_MAX_SIZE)
	resultQueue := make(chan GResult, RESULT_QUEUE_MAX_SIZE)
	return &GPool{gNum: num, taskQueue: taskQueue, resultQueue: resultQueue}
}

func (pool *GPool) AddTask(task func(...interface{}) interface{}, args ...interface{}) {
	id := uuid.NewV4().String()
	fmt.Printf("add %v start\n", id)
	pool.taskQueue <- GTask{Id: id, Task: task, Args: args}
	fmt.Printf("add %v stop\n", id)
}

func (pool *GPool) Start() {
	for i := 0; i < pool.gNum; i++ {
		pool.wg.Add(1)
		go func() {
			defer func() {
				fmt.Println("closed")
				pool.wg.Done()
			}()

			for task := range pool.taskQueue {
				fmt.Printf("take out %v\n", task.Id)
				res := task.Task(task.Args...)
				pool.resultQueue <- GResult{Id: task.Id, Res: res}
			}
		}()
	}
}

func (pool *GPool) Stop() {
	close(pool.taskQueue)
	pool.wg.Wait()
	close(pool.resultQueue)
}

func (pool *GPool) GetResult() {
	go func() {
		for result := range pool.resultQueue {
			fmt.Printf("TaskId: %v\tTaskResult: %v\n", result.Id, result.Res)
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
