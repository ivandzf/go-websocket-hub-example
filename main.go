package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

var port = "8080"

func main() {
	hub := newHub()
	go hub.run()

	router := gin.Default()
	router.GET("/connect/:name", func(c *gin.Context) {
		listen(hub, c)
	})

	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		panic(err)
	}
}