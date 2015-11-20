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
	pool.GetResult(func(r GResult) {
		fmt.Printf("Result: %v\n", r)
	})
	//pool.Heatbeat()

	for i := 0; i < 10000; i++ {
		pool.AddTask(func(args ...interface{}) interface{} {
			//fmt.Println("Task execute!")
			//fmt.Println(reflect.TypeOf(args))
			//fmt.Printf("%#v\n", args)
			time.Sleep(100 * time.Millisecond)
			return args
		}, i)
	}
	fmt.Printf("1!!!!!!!!!!!!!!!")
	//time.Sleep(30 * time.Second)
	//fmt.Printf("2!!!!!!!!!!!!!!!")
	//for i := 0; i < 1000; i++ {
	//pool.AddTask(func(args ...interface{}) interface{} {
	//fmt.Println("Task execute!")
	//fmt.Println(reflect.TypeOf(args))
	//fmt.Printf("%#v\n", args)
	//time.Sleep(100 * time.Millisecond)
	//return args
	//}, i)
	//}
	//time.Sleep(15 * time.Second)
	pool.Stop()
	fmt.Printf("Processed total task number: %v\n", pool.totalTask)
	if pool.totalTask != 10000 {
		t.Error("Task Lost!")
	}
}
