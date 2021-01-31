package main

import (
	"log"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
)

var userList = &sync.Map{}
var backupInterval = time.NewTicker(time.Minute * 1)
var changed = false

func backup() {
	defer sentry.Recover()
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
		setGistFile(GistID, `users`, up)
		changed = false
		log.Println(`>>>>Backuped!<<<<`)
	}
}

func download() {
	defer sentry.Recover()
	var list map[string]User
	json.Unmarshal([]byte(getGistFile(GistID, `users`).String()), &list)
	for k, v := range list {
		userList.Store(k, v)
	}
}
