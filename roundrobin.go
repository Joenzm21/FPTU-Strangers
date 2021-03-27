package main

import (
	"log"
	"math"
	"sync"

	"github.com/getsentry/sentry-go"
)

var queue = NewQueue(Limit)
var roundCounter = 0
var rrLock = &sync.RWMutex{}
var update = sync.NewCond(rrLock.RLocker())

func startRR() {
	defer sentry.Recover()
	for {
		first := queue.TwoBack()
		prev := first.Prev()
		rrLock.RLock()
		request1 := first.Value.(*FindingRequest)
		if request1.Session.State != `finding` {
			queue.Remove(first)
			continue
		}
		success := false
		for {
			for prev != nil && prev.Value == nil {
				old := prev
				prev = old.Prev()
				queue.Remove(old)
			}
			if prev == nil {
				break
			}
			request2 := prev.Value.(*FindingRequest)
			if request2.Session.State != `finding` {
				queue.Remove(prev)
				continue
			}
			if isSuitable(request1, request2) {
				request1.Session.State, request2.Session.State = `chating`, `chating`
				request1.Session.StateInfo, request2.Session.StateInfo = request2.Psid, request1.Psid
				notify := templates.Get(`notify`).Value().([]interface{})
				sendText(request1.Psid, notify...)
				sendText(request2.Psid, notify...)
				success = true
				roundCounter = 0
				queue.Remove(first)
				queue.Remove(prev)
				log.Println("Paired ", request1.Psid, request2.Psid)
				break
			}
			prev = prev.Prev()
		}
		if !success {
			if !queue.isFull() && request1 != nil {
				roundCounter++
				request1.Old = true
				queue.Lock.Lock()
				queue.Container.MoveToFront(first)
				queue.Lock.Unlock()
			} else if first != nil {
				queue.Remove(first)
				if request1 != nil {
					dropRequest(request1)
				}
			}
		}
		for roundCounter >= queue.Container.Len() {
			update.Wait()
			roundCounter = 0
		}
		rrLock.RUnlock()
	}
}

func dropRequest(request *FindingRequest) {
	request.Session.State = `idle`
	request.Session.StateInfo = nil
	sendText(request.Psid, templates.Get(`getstarted.onDrop`).Value().([]interface{})...)
	log.Println("Dropped request of ", request.Psid)
}

func isSuitable(request1 *FindingRequest, request2 *FindingRequest) bool {
	if request1.Old && request2.Old {
		return false
	}
	return request1.User.Gender == request2.Gender &&
		request2.User.Gender == request1.Gender &&
		int(math.Abs(float64(request1.User.Year-request2.Year))) <= MaxAgeDiff &&
		int(math.Abs(float64(request2.User.Year-request1.Year))) <= MaxAgeDiff
}
