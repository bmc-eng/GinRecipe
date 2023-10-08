package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gintest/handlers"
	"gintest/models"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Global Variables
var recipesHandler *handlers.RecipesHandler

// var recipes []Recipe
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

	// Set up Redis Cache
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	status := redisClient.Ping()
	fmt.Println(status)

	// Set up the recipesHandler
	recipesHandler = handlers.NewRecipesHandler(ctx, collection, redisClient)

}

// Function to initialise the database -
// This code does not need to be run once the database is set up
func InitializeDatabase() {
	// Read the contents of a JSON file containing all the recipes
	recipes := make([]models.Recipe, 0)
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

func main() {
	r := gin.Default()
	r.POST("/recipes", recipesHandler.NewRecipeHandler)
	r.GET("/recipes", recipesHandler.ListRecipesHandler)
	r.GET("/recipes/:id", recipesHandler.SingleRecipeHandler)
	r.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
	r.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	r.GET("recipes/search", recipesHandler.SearchRecipeHandler)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
