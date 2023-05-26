package timeWheel

import "time"

// task 任务
type task struct {
	//属性：延时时间、key、job函数、属于的圈
	delay  time.Duration
	key    string
	job    func()
	circle int
}
