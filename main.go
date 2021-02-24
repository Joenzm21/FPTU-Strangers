package main

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

var templates gjson.Result

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Debug:   true,
		Release: "fptu-strangers@1.0.0",
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()
	initMenu()
	download()
	payload, _ := ioutil.ReadFile(`templates.json`)
	templates = gjson.ParseBytes(payload)
	go startRR()
	go backup()
	startServer()
}
func startServer() {
	defer sentry.Recover()
	router := gin.New()
	router.Use(sentrygin.New(sentrygin.Options{}))
	router.POST(`/webhook`, handleRequest)
	router.GET(`/webhook`, func(c *gin.Context) {
		queries := c.Request.URL.Query()
		mode := queries[`hub.mode`][0]
		token := queries[`hub.verify_token`][0]
		challenge := queries[`hub.challenge`][0]
		if mode == `subscribe` && token == VerifyToken {
			c.AbortWithStatusJSON(200, challenge)
			return
		}
		c.AbortWithStatus(403)
	})
	router.GET(`/`, func(c *gin.Context) {
		c.Writer.WriteString("Server is running...")
		c.AbortWithStatus(200)
	})
	router.Run()
}
