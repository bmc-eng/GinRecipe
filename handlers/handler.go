package handlers

import (
	"encoding/json"
	"fmt"
	"gintest/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type RecipesHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

var err error

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redisClient,
	}
}

// ##########################
// ###### GET Requests ######
// ##########################

// Function to list all of the recipes - Updated for MongoDB & Redis
func (handler *RecipesHandler) ListRecipesHandler(c *gin.Context) {
	val, err := handler.redisClient.Get("recipes").Result()
	if err == redis.Nil {
		log.Printf("Request to MongoDB")
		cur, err := handler.collection.Find(handler.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cur.Close(handler.ctx)
		recipes := make([]models.Recipe, 0)
		for cur.Next(handler.ctx) {
			var recipe models.Recipe
			cur.Decode(&recipe)
			recipes = append(recipes, recipe)
		}
		data, _ := json.Marshal(recipes)
		handler.redisClient.Set("recipes", string(data), 0)
		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	} else {
		log.Printf("Request to Redis")
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(val), &recipes)
		c.JSON(http.StatusOK, recipes)
	}

}

// Function to return a recipe by id
func (handler *RecipesHandler) SingleRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	// Convert to primitive object
	objectId, _ := primitive.ObjectIDFromHex(id)
	cur, err := handler.collection.Find(handler.ctx, bson.M{"_id": objectId})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(handler.ctx)
	var recipe models.Recipe
	for cur.Next(handler.ctx) {
		cur.Decode(&recipe)
	}

	c.JSON(http.StatusOK, recipe)
}

// Handle search query from API Example: http://localhost:8080/recipes/search?tag=vegetarian - updated for MongoDB
func (handler *RecipesHandler) SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecipes := make([]models.Recipe, 0)

	// Search through MongoDB for the tag
	cur, err := handler.collection.Find(handler.ctx, bson.M{"tags": tag})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(handler.ctx)
	for cur.Next(handler.ctx) {
		var recipe models.Recipe
		cur.Decode(&recipe)
		listOfRecipes = append(listOfRecipes, recipe)
	}

	c.JSON(http.StatusOK, listOfRecipes)

}

// ##########################
// ##### POST Requests ######
// ##########################

// Function to add a new recipe to Database
// Add in MongoDB functionality
func (handler *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err = handler.collection.InsertOne(handler.ctx, recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new recipe"})
		return
	}
	message := gin.H{"message": "New Recipe added", "id": recipe.ID}
	c.JSON(http.StatusOK, message)

}

// ##########################
// ###### PUT Requests ######
// ##########################

// Function to update a recipe given an id
func (handler *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	// Get the id from the PUT request
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	// Find the recipe that matches the index
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err = handler.collection.UpdateOne(handler.ctx, bson.M{"_id": objectId}, bson.D{{"$set", bson.D{
		{"name", recipe.Name},
		{"instructions", recipe.Instructions},
		{"ingredients", recipe.Ingredients},
		{"tags", recipe.Tags},
	}}})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been updated"})
}

// ##########################
// #### DELETE Requests #####
// ##########################

// Updated for Mongo DB
func (handler *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	// get the id from the parameter passed into the url
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	cur, err := handler.collection.DeleteOne(handler.ctx, bson.M{"_id": objectId})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully removed record: " + id,
		"count":   cur.DeletedCount})

}
