package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		id := uuid.New()
		c.JSON(http.StatusOK, gin.H{
			"message": id.String(),
		})
	})
	r.GET("/home", func(c *gin.Context) {
		message := "Hello"
		c.JSON(http.StatusOK, gin.H{
			"message": message,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
