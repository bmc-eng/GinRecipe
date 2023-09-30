package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func IndexHandler(c *gin.Context) {
	name := c.Params.ByName("name")
	id := uuid.New()
	c.JSON(http.StatusOK, gin.H{
		"user": "Hello " + name,
		"id":   id.String(),
	})
}

func main() {
	r := gin.Default()
	r.GET(":name", IndexHandler)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
