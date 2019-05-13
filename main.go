package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AtomURLEntry struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	ShortURL       string             `json:"shortURL" bson:"shortURL"`
	DestinationURL string             `json:"destinationURL" bson:"destinationURL"`
	CreatedAt      int64              `json:"created_at" bson:"created_at"`
}

func connectToDatabase(mangoDatabaseURL string) *mongo.Client {
	databaseURL := fmt.Sprint(mangoDatabaseURL)
	connectOptions := options.Client()
	connectOptions.ApplyURI(databaseURL)

	connectContext, errorInContext := context.WithTimeout(context.Background(), 10*time.Second)

	defer errorInContext()

	databaseClient, errInConnection := mongo.Connect(connectContext, connectOptions)

	if errInConnection != nil {
		panic("Failed to connect to DB")
	}

	errInPing := databaseClient.Ping(connectContext, nil)

	if errInPing != nil {
		panic("DB not found")
	}

	return databaseClient
}

func isInputJsonValid(shortURL string, longURL string) bool {
	lengthOfShortURL := len(strings.TrimSpace(shortURL))
	lengthOfLongURL := len(strings.TrimSpace(longURL))

	if lengthOfShortURL == 0 || lengthOfLongURL == 0 {
		return false
	}
	return true
}

func welcomePage(ginContext *gin.Context) {

	ginContext.JSON(200, gin.H{"name": "Hello"})
}

func redirectURLHandler(ginContext *gin.Context, dbCollection *mongo.Collection, connectContext context.Context) {
	var atomURLEntry AtomURLEntry
	shortURLEntered := ginContext.Param("shortURL")

	filterOptions := bson.M{
		"shortURL": shortURLEntered,
	}

	err := dbCollection.FindOne(connectContext, filterOptions).Decode(&atomURLEntry)
	if err != nil {
		ginContext.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	}

	ginContext.Redirect(http.StatusFound, atomURLEntry.DestinationURL)
}

func addURLHandler(ginContext *gin.Context, dbCollection *mongo.Collection, connectContext context.Context) {
	var atomURLEntry AtomURLEntry

	errInParsingInput := ginContext.ShouldBindJSON(&atomURLEntry)
	if errInParsingInput != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Wrong JSON structure"})
	}

	if isInputJsonValid(atomURLEntry.ShortURL, atomURLEntry.DestinationURL) == false {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Empty fields"})
	}

	atomURLEntry.ShortURL = strings.TrimSpace(atomURLEntry.ShortURL)
	atomURLEntry.DestinationURL = strings.TrimSpace(atomURLEntry.DestinationURL)
	createdAt := time.Now().Unix()
	atomURLEntry.CreatedAt = createdAt

	atomURLEntryToAdd := bson.M{
		"shortURL":       atomURLEntry.ShortURL,
		"destinationURL": atomURLEntry.DestinationURL,
		"create_at":      atomURLEntry.CreatedAt,
	}

	addedAtomURLEntry, errInAdding := dbCollection.InsertOne(connectContext, atomURLEntryToAdd)
	if errInAdding != nil {
		ginContext.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while saving to database"})
	}

	ginContext.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "data": addedAtomURLEntry})
}

func main() {
	port := os.Getenv("PORT")
	if os.Getenv("PORT") == "" {
		port = "8000"
	}

	databaseUserName := os.Getenv("DB_USER")
	if os.Getenv("DB_USER") == "" {
		log.Fatal("No Database user name provided")
	}

	databaseUserPassword := os.Getenv("DB_PASSWORD")
	if os.Getenv("DB_PASSWORD") == "" {
		log.Fatal("No Database user's password provided")
	}

	databaseAddress := os.Getenv("DB_URL")
	if os.Getenv("DB_URL") == "" {
		log.Fatal("No Database URL provided")
	}

	databaseName := os.Getenv("DB_NAME")
	if os.Getenv("DB_NAME") == "" {
		log.Fatal("NO Database name provided")
	}

	databaseURL := fmt.Sprint("mongodb://" + databaseUserName + ":" + databaseUserPassword + "@" + databaseAddress + "/" + databaseName)

	database := connectToDatabase(databaseURL)
	shortURLsCollection := database.Database("atom-url-db").Collection("shorturls")

	connectContext, errorInContext := context.WithTimeout(context.Background(), 30*time.Second)
	defer errorInContext()

	// defining new router
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/", func(ginContext *gin.Context) {
		welcomePage(ginContext)
	})

	router.GET("/:shortURL", func(ginContext *gin.Context) {
		redirectURLHandler(ginContext, shortURLsCollection, connectContext)
	})

	router.POST("/", func(ginContext *gin.Context) {
		addURLHandler(ginContext, shortURLsCollection, connectContext)
	})

	err := router.Run(":" + port)
	if err != nil {
		fmt.Printf("Cannot start server")
	}
}
