package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

var userList = &sync.Map{}
var banned = &sync.Map{}
var backupInterval = time.NewTicker(time.Minute * 10)
var changed = false

func backup() {
	for {
		<-backupInterval.C
		if !changed {
			continue
		}
		up := make(map[string]interface{})
		userList.Range(func(k interface{}, v interface{}) bool {
			up[k.(string)] = v
			return true
		})
		fmt.Println(up)
		setGistFile(GistID, `users`, up)
		changed = false
		fmt.Println(`>>>>Backuped!<<<<`)
	}
}

func download() {
	for k, v := range gjson.Parse(getGistFile(GistID, `users`).String()).Value().(map[string]interface{}) {
		userList.Store(k, v)
	}
}
