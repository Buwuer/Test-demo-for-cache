package main

import (
	"github.com/gin-gonic/gin"
	"webcache/api"
)

func main() {
	r := gin.Default()
	r.POST("/lottery", api.LotteryApi)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
