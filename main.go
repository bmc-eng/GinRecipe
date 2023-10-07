package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Recipe struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

// Global Variables
var recipes []Recipe
var ctx context.Context
var err error
var client *mongo.Client
var collection *mongo.Collection

// ##########################
// ##### Initial setup ######
// ##########################
func init() {

	// Connect to MongoDB
	ctx = context.Background()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	collection = client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	log.Println("Connected to MongoDB")

}

// Function to initialise the database -
// This code does not need to be run once the database is set up
func InitializeDatabase() {
	// Read the contents of a JSON file containing all the recipes
	recipes = make([]Recipe, 0)
	file, _ := os.ReadFile("backup/recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)

	var listOfRecipes []interface{}
	for _, recipe := range recipes {
		listOfRecipes = append(listOfRecipes, recipe)
	}

	insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Inserted recipes: ", len(insertManyResult.InsertedIDs))
}

// ##########################
// ###### GET Requests ######
// ##########################

// Function to list all of the recipes - Updated for MongoDB
func ListRecipesHandler(c *gin.Context) {
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(ctx)
	recipes := make([]Recipe, 0)
	for cur.Next(ctx) {
		var recipe Recipe
		cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, recipes)
}

// Function to return a recipe by id
func SingleRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	// Convert to primitive object
	objectId, _ := primitive.ObjectIDFromHex(id)
	cur, err := collection.Find(ctx, bson.M{"_id": objectId})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(ctx)
	var recipe Recipe
	for cur.Next(ctx) {
		cur.Decode(&recipe)
	}

	c.JSON(http.StatusOK, recipe)
}

// Handle search query from API Example: http://localhost:8080/recipes/search?tag=vegetarian - updated for MongoDB
func SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecipes := make([]Recipe, 0)

	// Search through MongoDB for the tag
	cur, err := collection.Find(ctx, bson.M{"tags": tag})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var recipe Recipe
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
func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err = collection.InsertOne(ctx, recipe)
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
func UpdateRecipeHandler(c *gin.Context) {
	// Get the id from the PUT request
	id := c.Param("id")
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	// Find the recipe that matches the index
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectId}, bson.D{{"$set", bson.D{
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
func DeleteRecipeHandler(c *gin.Context) {
	// get the id from the parameter passed into the url
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	cur, err := collection.DeleteOne(ctx, bson.M{"_id": objectId})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully removed record: " + id,
		"count":   cur.DeletedCount})

}

func main() {
	r := gin.Default()
	r.POST("/recipes", NewRecipeHandler)
	r.GET("/recipes", ListRecipesHandler)
	r.GET("/recipes/:id", SingleRecipeHandler)
	r.PUT("/recipes/:id", UpdateRecipeHandler)
	r.DELETE("/recipes/:id", DeleteRecipeHandler)
	r.GET("recipes/search", SearchRecipeHandler)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
