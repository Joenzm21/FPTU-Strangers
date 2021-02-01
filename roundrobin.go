package main

import (
	"math"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
)

var queue = NewQueue(Limit)
var roundCounter = 0
var rrLock = &sync.Mutex{}
var update = sync.NewCond(rrLock)

func startRR() {
	defer sentry.Recover()
	rrLock.Lock()
	for {
		request1 := queue.Dequeue().(*FindingRequest)
		next := queue.Back()
		success := false
		for next != nil {
			request2 := next.Value.(*FindingRequest)
			if isSuitable(request1, request2) {
				queue.Remove(next)
				request1.Session.State, request2.Session.State = `chating`, `chating`
				request1.Session.StateInfo, request2.Session.StateInfo = request2.Psid, request1.Psid
				notify := templates.Get(`notify`).Value().([]interface{})
				sendText(request1.Psid, notify...)
				sendText(request2.Psid, notify...)
				success = true
				roundCounter = 0
				break
			}
			next = next.Next()
		}
		if !success {
			if !queue.isFull() && time.Now().Sub(request1.Time) < 5*time.Minute {
				request1.Old = true
				queue.Enqueue(request1)
				roundCounter++
			} else {
				dropRequest(request1)
			}
		}
		for roundCounter >= queue.Container.Len() {
			update.Wait()
			roundCounter = 0
		}
	}
}

func dropRequest(request *FindingRequest) {
	request.Session.State = `idle`
	request.Session.StateInfo = nil
	sendText(request.Psid, templates.Get(`getstarted.onDrop`).Value().([]interface{})...)
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
