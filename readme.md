# Gin Recipe API Example

This is an implementation of Building Distributed Application in Gin with some enhancements to the code. 

### Feature Development
The API currently has the following features:

- Get all Recipes using GET /recipes 
- Get a single recipe using GET /recipes/id
- Create a new recipe using POST /recipes {JSON recipe}
- Update a recipe using PUT /recipe/id
- Delete a recipe using DELETE recipes/id
- Search for a recipe using GET recipes/tag

### MongoDB Set up
All functionality updated for MongoDB. The MongoDB database needs to be run and needs to have records imported. Run mongodb in Docker with the following code:

```docker run -d --name mongodb -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=password -p 27017:27017 mongo:4.4.3```

Confirm running with ```docker rs```

To run with mongoDB initialised:

```MONGO_URI="mongodb://admin:password@localhost:27017/test?authSource=admin" MONGO_DATABASE=demo go run main.go```

### Caching with Redis
Adding caching functionality to the API with redis. We use docker again to run the redis container. Ensure that the following is running before starting the application. Set the redis policy so that has a maximum size of 512MB in the redis config file.

```docker run -d -v redis:/usr/local/etc/redis --name redis -p 6379:6379 redis:6.0```


