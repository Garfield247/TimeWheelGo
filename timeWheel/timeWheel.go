package timeWheel

import (
	"container/list"
	"fmt"
	"log"
	"time"
)

// TimeWheel 时间轮
type TimeWheel struct {
	//属性：时间间隔、时间轮槽位数、当前所在槽位
	interval time.Duration
	slotNum  int
	currPos  int

	//属性：列表数组（当做时间轮） []*list.List
	slots []*list.List
	//属性：任务对应的节点信息 map[key]location
	m map[string]location

	//属性：添加任务chan、停止chan、移除任务chan
	addTaskChan    chan task
	stopChan       chan bool
	removeTaskChan chan string
	//属性：定时器
	timer *time.Ticker
}

// location 任务所处时间轮上的位置（方便移除任务）
type location struct {
	pos  int           //任务所在数组上的槽位
	elem *list.Element //任务所在list的节点
}

// NewTimeWheel 初始化
func NewTimeWheel(interval time.Duration, slotNum int) *TimeWheel {
	tw := &TimeWheel{
		interval:       interval,
		slotNum:        slotNum,
		currPos:        0,
		slots:          make([]*list.List, slotNum), //还需初始化*list.List
		m:              make(map[string]location),
		addTaskChan:    make(chan task),
		stopChan:       make(chan bool),
		removeTaskChan: make(chan string),
	}
	tw.initList()
	return tw
}

// 初始化槽位
func (tw *TimeWheel) initList() {
	for i, _ := range tw.slots {
		tw.slots[i] = list.New()
	}
}

// Start 运行
func (tw *TimeWheel) Start() {
	//初始化定时器
	tw.timer = time.NewTicker(tw.interval)
	go tw.start()
}

func (tw *TimeWheel) start() {
	//执行循环
	for {
		select {
		case <-tw.timer.C: //定时任务
			tw.tickHandle()
		case task := <-tw.addTaskChan: //新任务队列
			tw.addRealJob(&task)
		case key := <-tw.removeTaskChan: //移除任务
			tw.removeRealJob(key)
		case <-tw.stopChan: //停止
			tw.Stop()
			return //直接退出
		}
	}
}

// 执行任务
func (tw *TimeWheel) tickHandle() {
	//获取当前槽位队列
	list := tw.slots[tw.currPos]
	//槽位自增
	tw.currPos++
	if tw.currPos == tw.slotNum-1 {
		tw.currPos = 0
	}
	//起协程执行job
	go tw.scanAndRunJob(list)
}

func (tw *TimeWheel) scanAndRunJob(l *list.List) {
	//寻找可执行的任务
	//不能用该方法遍历list.List，因为list.Remove(elem)时，elem前驱、后缀指针会置空
	//所以elem = elem.Next()会导致，elem = nil
	//for elem := l.Front(); elem != nil; elem = elem.Next() {
	for elem := l.Front(); elem != nil; {
		task := elem.Value.(*task)
		if task.key == "k2" {
			fmt.Println(task)
		}
		//对多圈的任务，进行圈数-1
		if task.circle > 0 {
			task.circle--
			elem = elem.Next()
			continue
		}
		//剩下的为当前需要执行的任务（起协程执行任务）
		go func() {
			defer func() { //用于捕获 job的panic
				if err := recover(); err != nil {
					log.Fatalf("job err: %v", err)
					return
				}
			}()
			//执行任务
			task.job()
		}()
		//保存下一个elem
		next := elem.Next()
		//将当前的移除出队列、map
		l.Remove(elem)
		if task.key != "" {
			delete(tw.m, task.key)
		}
		elem = next
	}

}

// AddJob 添加新任务（将任务添加进队列）
func (tw *TimeWheel) AddJob(delay time.Duration, key string, job func()) {
	if delay < 0 { //传入的是延时多长时间
		return
	}
	tw.addTaskChan <- task{delay: delay, key: key, job: job}
}

// 添加任务（将任务添加进时间轮上）
func (tw *TimeWheel) addRealJob(task *task) {
	pos, circle := tw.getTaskPosAndCircle(task.delay)
	task.circle = circle
	//判断是否存在，若存在，则需先删除槽位上的任务，再加入
	if task.key != "" {
		//if _, ok := tw.m[task.key]; ok {
		//	tw.removeRealJob(task.key)
		//}
		tw.removeRealJob(task.key)
		//槽位
		elem := tw.slots[pos].PushBack(task)
		//map
		tw.m[task.key] = location{
			pos:  pos,
			elem: elem,
		}
	}
	fmt.Println("addRealJob end..., key:", task.key)
}

// RemoveJob 移除(将要移除的key放入chan) -- 解耦
func (tw *TimeWheel) RemoveJob(key string) {
	if key == "" {
		return
	}
	tw.removeTaskChan <- key
}

// 移除任务（真实的从时间轮上移除）
func (tw *TimeWheel) removeRealJob(key string) {
	loc, ok := tw.m[key]
	if !ok {
		return
	}
	tw.slots[loc.pos].Remove(loc.elem) //移除槽位队列
	delete(tw.m, key)                  //移除map
}

func (tw *TimeWheel) Stop() {
	tw.timer.Stop()
}

// 获取task所处时间轮的位置和圈数
func (tw *TimeWheel) getTaskPosAndCircle(delay time.Duration) (pos int, circle int) {
	//时间轮间隔
	//需要转的圈数
	//所在时间轮槽位
	delaySecond := int(delay.Seconds())
	circleSecond := int(tw.interval.Seconds())
	circle = int(delaySecond / circleSecond / tw.slotNum)
	pos = int(tw.currPos+delaySecond/circleSecond) % tw.slotNum
	return
}
