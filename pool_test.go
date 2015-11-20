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

	for i := 0; i < 100000; i++ {
		pool.AddTask(func(args ...interface{}) interface{} {
			//fmt.Println("Add TaskId")
			//fmt.Println(reflect.TypeOf(args))
			//fmt.Printf("%#v\n", args)
			time.Sleep(100 * time.Millisecond)
			return args
		}, i)
	}
	pool.Stop()
	time.Sleep(time.Second)
	fmt.Printf("Processed total task number: %v\n", pool.totalTask)
	if pool.totalTask != 100000 {
		t.Error("Task Lost!")
	}
}
