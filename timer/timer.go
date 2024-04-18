package timer

import (
	"sync"

	"github.com/robfig/cron/v3"
)

type Timer interface {
	FindTaskList() map[_ModuleName]*manager
	AddTask(module, spec, name string, fn func(), options ...cron.Option) error
	StartModuleTask(name string)
	StopModuleTask(name string)
	DelModuleTask(module string)
	DelTaskById(module string, id int)
	Close()
}

type _ModuleName string //模块名称

type timer struct {
	sync.Mutex
	allTask map[_ModuleName]*manager
}

type task struct {
	EntryID  cron.EntryID //定时任务ID
	Spec     string       //规则
	TaskName string       //定时任务名称
}

type manager struct {
	instance *cron.Cron
	tasks    map[cron.EntryID]*task
}

func NewTimer() Timer {
	return &timer{
		allTask: make(map[_ModuleName]*manager),
	}
}

// FindTaskList 定时任务列表
func (e *timer) FindTaskList() map[_ModuleName]*manager {
	e.Lock()
	defer e.Unlock()
	return e.allTask
}

/*
AddTask 添加定时任务,参数说明:

	1、模块名,将定时任务分组
	2、定时规则
	3、定时任务名称
	4、处理逻辑
	5、可选,实例对象配置信息
*/
func (e *timer) AddTask(module, spec, name string, fn func(), options ...cron.Option) error {
	e.Lock()
	defer e.Unlock()
	//说明该模块任务不存在
	if _, ok := e.allTask[_ModuleName(module)]; !ok {
		e.allTask[_ModuleName(module)] = &manager{
			instance: cron.New(options...),
			tasks:    make(map[cron.EntryID]*task),
		}
	}
	//添加
	id, err := e.allTask[_ModuleName(module)].instance.AddFunc(spec, fn)
	if err != nil {
		return err
	}
	//启动
	e.allTask[_ModuleName(module)].instance.Start()
	//完善数据结构
	e.allTask[_ModuleName(module)].tasks[id] = &task{
		EntryID:  id,
		Spec:     spec,
		TaskName: name,
	}
	return nil
}

// StartModuleTask 启动该模块定时任务
func (e *timer) StartModuleTask(module string) {
	e.Lock()
	defer e.Unlock()
	if v, ok := e.allTask[_ModuleName(module)]; ok {
		v.instance.Start()
	}
}

// StopModuleTask 停止该模块定时任务
func (e *timer) StopModuleTask(module string) {
	e.Lock()
	defer e.Unlock()
	if v, ok := e.allTask[_ModuleName(module)]; ok {
		v.instance.Stop()
	}
}

// DelModuleTask 删除该模块任务
func (e *timer) DelModuleTask(module string) {
	e.Lock()
	defer e.Unlock()
	if v, ok := e.allTask[_ModuleName(module)]; ok {
		v.instance.Stop()
		delete(e.allTask, _ModuleName(module))
	}
}

// DelTaskById 删除模块内某个任务
func (e *timer) DelTaskById(module string, id int) {
	e.Lock()
	defer e.Unlock()
	if v, ok := e.allTask[_ModuleName(module)]; ok {
		v.instance.Remove(cron.EntryID(id))
		delete(v.tasks, cron.EntryID(id))
	}
}

// Close 停止所有定时任务
func (e *timer) Close() {
	e.Lock()
	defer e.Unlock()
	for _, v := range e.allTask {
		v.instance.Stop()
	}
}
