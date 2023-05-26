package main

import (
	"fmt"
	"time"

	"github.com/Garfield247/TimeWheelGo.git/timeWheel"
)

var tw = timeWheel.NewTimeWheel(time.Second, 3600)

var num = 0

func init() {
	tw.Start() //启动时间轮
}

// AddJob 添加任务
func AddJob(t time.Time, key string, job func()) {
	tw.AddJob(t.Sub(time.Now()), key, job)
}

// CancelJob 取消任务
func CancelJob(key string) {
	tw.RemoveJob(key)
}

func func1() {

	fmt.Println("func1 job after:", time.Now().Format("2006-01-02 15:04:05"), "num", num)
	num += 1
	if num < 100 {
		AddJob(time.Now().Add(time.Second*5), "func1", func1)
	}
}

func main() {
	fmt.Println("k1 job before:", time.Now().Format("2006-01-02 15:04:05"))
	AddJob(time.Now().Add(time.Second*5), "k1", func1)

	fmt.Println("k2 job before:", time.Now().Format("2006-01-02 15:04:05"))
	AddJob(time.Now().Add(time.Second*10), "k2", func() {
		fmt.Println("k2 job after:", time.Now().Format("2006-01-02 15:04:05"))
	})

	fmt.Println("k3 job before:", time.Now().Format("2006-01-02 15:04:05"))
	AddJob(time.Now().Add(time.Second*10), "k3", func() {
		fmt.Println("k3 job after:", time.Now().Format("2006-01-02 15:04:05"))
	})

	CancelJob("k2")
	select {}
	//time.Sleep(time.Second * 20)
}
