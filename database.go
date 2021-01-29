package main

import (
	"fmt"
	"sync"
	"time"
)

var userList = &sync.Map{}
var banned = &sync.Map{}
var backupInterval = time.NewTicker(time.Minute * 1)
var changed = false

func backup() {
	for {
		<-backupInterval.C
		if !changed {
			continue
		}
		up := make(map[string]User)
		userList.Range(func(k interface{}, v interface{}) bool {
			up[k.(string)] = v.(User)
			return true
		})
		fmt.Println(up)
		setGistFile(GistID, `users`, up)
		changed = false
		fmt.Println(`>>>>Backuped!<<<<`)
	}
}

func download() {
	var list map[string]User
	json.Unmarshal([]byte(getGistFile(GistID, `users`).String()), &list)
	for k, v := range list {
		userList.Store(k, v)
	}
}
