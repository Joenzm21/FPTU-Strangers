package main

import (
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

//Postback -
type Postback struct {
	Type    string   `json:"template_type"`
	Text    string   `json:"text"`
	Buttons []Button `json:"buttons"`
}

//QAState -
type QAState struct {
	Template      gjson.Result
	Answers       []interface{}
	Counter       int
	CheckFuncs    []func(answer string) bool
	LastStateInfo interface{}
	OnDone        func()
	OnCancel      func(oldState interface{})
}

//FindingRequest -
type FindingRequest struct {
	Psid    string
	Year    int
	Gender  string
	User    User
	Session *Session
	Time    time.Time
	Old     bool
}

//Button -
type Button struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Payload string `json:"payload"`
}

//Session -
type Session struct {
	State     string
	StateInfo interface{}
	Timeout   *time.Timer
	Lock      *sync.Mutex
}

//CancelingState -
type CancelingState struct {
	OnYes         func()
	OnNo          func()
	LastState     string
	LastStateInfo interface{}
	NewState      string
	NewStateInfo  interface{}
}

//User -
type User struct {
	Gender     string `json:"gender"`
	Year       int    `json:"year"`
	Scam       int    `json:"scam"`
	Unfriendly int    `json:"unfriendly"`
}
