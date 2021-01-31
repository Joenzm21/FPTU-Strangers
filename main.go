package main

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

var templates gjson.Result

func main() {
	initMenu(`getstarted.json`)
	initMenu(`persistentmenu.json`)
	download()
	payload, _ := ioutil.ReadFile(`templates.json`)
	templates = gjson.ParseBytes(payload)
	go startRR()
	go backup()
	startServer()
}
func startServer() {
	router := gin.Default()
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
