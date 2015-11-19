package pool

import (
	"fmt"
	"testing"
	"time"
)

func TestGPool(t *testing.T) {
	t.Log("Hello")

	pool := NewGPool(2)
	pool.Start()
	pool.GetResult()
	//pool.Heatbeat()

	for i := 0; i < 1000; i++ {
		pool.AddTask(func(args ...interface{}) interface{} {
			fmt.Println("Add TaskId")
			//fmt.Println(reflect.TypeOf(args))
			fmt.Printf("%#v\n", args)
			return args
		}, i)
	}
	pool.Stop()
	time.Sleep(time.Second)
	if pool.totalTask != 1000 {
		t.Error("Task Lost!")
	}
}
