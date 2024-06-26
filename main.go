package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"gintest/handlers"
	"gintest/models"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	redisStore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Global Variables
var recipesHandler *handlers.RecipesHandler
var authHandler *handlers.AuthHandler

// var recipes []Recipe
var ctx context.Context
var err error
var client *mongo.Client
var collection *mongo.Collection

// ##########################
// ##### Initial setup ######
// ##########################
func init() {

	// Add Auth key to the os.env
	os.Setenv("JWT_SECRET", EnvVariable("JWT_SECRET"))

	// Connect to MongoDB for the recipes
	ctx = context.Background()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(EnvVariable("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	collection = client.Database(EnvVariable("MONGO_DATABASE")).Collection("recipes")
	collectionUsers := client.Database(EnvVariable("MONGO_DATABASE")).Collection("users")

	log.Println("Connected to MongoDB")

	// Initialize Database - To add check to see if DB is empty
	InitializeDatabase(collection)
	InitializeUsers(collectionUsers)

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
	authHandler = handlers.NewAuthHandler(ctx, collectionUsers)

}

// ###########################
// ##### HELPER FUNCTION #####
// ###########################

// Function to initialise the database -
// This code does not need to be run once the database is set up
func InitializeDatabase(collection *mongo.Collection) {
	// Read the contents of a JSON file containing all the recipes
	recipes := make([]models.Recipe, 0)
	file, _ := os.ReadFile("backup/recipes.json")
	err := json.Unmarshal([]byte(file), &recipes)
	if err != nil {
		log.Fatal("Failed to convert json file to recipes object: ", err)
	}

	var listOfRecipes []interface{}
	for _, recipe := range recipes {
		listOfRecipes = append(listOfRecipes, recipe)
	}

	fmt.Println(recipes)
	insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Inserted recipes: ", len(insertManyResult.InsertedIDs))
}

// This is to initially populate the users database
func InitializeUsers(collection *mongo.Collection) {
	users := map[string]string{
		"admin":   "fCRmh4Q2J7Rseqkz",
		"bclarke": "test123",
		"test":    "test123",
	}

	h := sha256.New()
	for username, password := range users {
		collection.InsertOne(ctx, bson.M{
			"username": username,
			"password": string(h.Sum([]byte(password))),
		})
	}
	log.Println("Users inserted correctly")
}

func EnvVariable(key string) string {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func main() {
	r := gin.Default()

	store, _ := redisStore.NewStore(10, "tcp",
		"localhost:6379", "", []byte("secret"))
	r.Use(sessions.Sessions("recipes_api", store))

	// Add Authentication section
	authorized := r.Group("/")
	authorized.Use(authHandler.AuthMiddleware())
	{
		authorized.POST("/recipes", recipesHandler.NewRecipeHandler)
		authorized.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
		authorized.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	}

	// Get functionality does not need authorisation
	r.GET("/recipes", recipesHandler.ListRecipesHandler)
	r.GET("/recipes/:id", recipesHandler.SingleRecipeHandler)
	r.GET("/recipes/search", recipesHandler.SearchRecipeHandler)

	// Allow the user to sign in outside requiring authentication
	r.POST("/signin", authHandler.SignInHandler)
	r.POST("/refresh", authHandler.RefreshHandler)
	r.POST("/signout", authHandler.SignOutHandler)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
