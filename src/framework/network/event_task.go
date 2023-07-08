package network

import "sync"

type EventFunc func(interface{}) error

type EventTask struct {
	eventFunc EventFunc
	param     interface{}
}

func NewEventTask(f EventFunc, param interface{}) *EventTask {
	return &EventTask{
		eventFunc: f,
		param:     param,
	}
}

func (t *EventTask) Execute() error {
	return t.eventFunc(t.param)
}

type EventTaskQueue struct {
	mutex sync.Mutex
	tasks []*EventTask
}

func (q *EventTaskQueue) Enqueue(t *EventTask) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.tasks = append(q.tasks, t)
}

func (q *EventTaskQueue) Dequeue() *EventTask {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if len(q.tasks) <= 0 {
		return nil
	}
	t := q.tasks[len(q.tasks)-1]
	q.tasks = q.tasks[0 : len(q.tasks)-1]
	return t
}
