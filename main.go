package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt"`
}

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
}

func IndexHandler(c *gin.Context) {
	name := c.Params.ByName("name")
	id := uuid.New()
	c.JSON(http.StatusOK, gin.H{
		"user": "Hello " + name,
		"id":   id.String(),
	})
}

func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}
	recipe.ID = uuid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	c.JSON(http.StatusOK, recipe)

}

func main() {
	r := gin.Default()
	r.GET(":name", IndexHandler)
	r.POST("/recipes", NewRecipeHandler)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
