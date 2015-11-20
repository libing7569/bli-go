package pool

import (
	"testing"
	"time"
)

func TestGPool(t *testing.T) {
	pool := NewGPool(2)
	pool.Start()
	pool.GetResult(func(r GResult) {
		t.Logf("Result: %v\n", r)
	})
	//pool.Heatbeat()

	for i := 0; i < 1000000; i++ {
		pool.AddTask(func(args ...interface{}) interface{} {
			//fmt.Println("Task execute!")
			//fmt.Println(reflect.TypeOf(args))
			//fmt.Printf("%#v\n", args)
			time.Sleep(10 * time.Millisecond)
			return args
		}, i)
	}
	//time.Sleep(30 * time.Second)
	//for i := 0; i < 1000; i++ {
	//pool.AddTask(func(args ...interface{}) interface{} {
	//fmt.Println("Task execute!")
	//fmt.Println(reflect.TypeOf(args))
	//fmt.Printf("%#v\n", args)
	//time.Sleep(100 * time.Millisecond)
	//return args
	//}, i)
	//}
	time.Sleep(90 * time.Second)
	//pool.Stop()

	t.Logf("Processed total task number: %v\n", pool.totalTask)
	if pool.totalTask != 1000000 {
		t.Error("Task Lost!")
	}
}
