package main

import (
	"container/list"
	"sync"
)

//Queue -
type Queue struct {
	Container         *list.List
	Lock              *sync.RWMutex
	nonEmpty, nonFull *sync.Cond
	Limit             int
}

//NewQueue -
func NewQueue(limit int) *Queue {
	lock := &sync.RWMutex{}
	return &Queue{
		Container: list.New(),
		Lock:      lock,
		nonFull:   sync.NewCond(lock),
		nonEmpty:  sync.NewCond(lock.RLocker()),
		Limit:     limit,
	}
}

//Enqueue -
func (q *Queue) Enqueue(item interface{}) *list.Element {
	q.Lock.Lock()
	for q.isFull() {
		q.nonFull.Wait()
	}
	result := q.Container.PushFront(item)
	q.Lock.Unlock()
	q.nonEmpty.Broadcast()
	return result
}

//Remove -
func (q *Queue) Remove(el *list.Element) {
	if el != nil {
		q.Container.Remove(el)
		q.nonFull.Broadcast()
		el.Value = nil
		el = nil
	}
}

//TwoBack -
func (q *Queue) TwoBack() *list.Element {
	q.Lock.RLock()
	for q.Container.Len() < 2 {
		q.nonEmpty.Wait()
	}
	q.Lock.RUnlock()
	return q.Container.Back()
}

func (q *Queue) isFull() bool {
	return q.Container.Len() >= q.Limit
}
