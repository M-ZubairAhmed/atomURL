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

func corsMiddleware() gin.HandlerFunc {
	return func(ginContext *gin.Context) {
		ginContext.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		ginContext.Writer.Header().Set("Access-Control-Max-Age", "86400")
		ginContext.Writer.Header().Set("Access-Control-Allow-Methods", "GET , POST")
		ginContext.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
		ginContext.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		ginContext.Next()
	}
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

func webAppHandler(ginContext *gin.Context) {
	ginContext.File("./web/build/index.html")
}

func redirectURLHandler(ginContext *gin.Context, dbCollection *mongo.Collection) {
	var atomURLEntry AtomURLEntry
	shortURLEntered := ginContext.Param("shortURL")

	connectContext, cancelContext := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelContext()

	filterOptions := bson.M{
		"shortURL": shortURLEntered,
	}

	errInFinding := dbCollection.FindOne(connectContext, filterOptions).Decode(&atomURLEntry)
	if errInFinding != nil {
		ginContext.JSON(http.StatusNotFound, gin.H{"error": "Not found",
			"error_details": errInFinding.Error()})
		connectContext.Done()
		return
	}

	ginContext.Redirect(http.StatusFound, atomURLEntry.DestinationURL)
	connectContext.Done()
}

func addURLHandler(ginContext *gin.Context, dbCollection *mongo.Collection) {
	var atomURLEntry AtomURLEntry
	var atomURLExistingEntry AtomURLEntry

	connectContext, cancelContext := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelContext()

	errInParsingInput := ginContext.ShouldBindJSON(&atomURLEntry)
	if errInParsingInput != nil {
		ginContext.JSON(http.StatusNotAcceptable, gin.H{"error": "Wrong JSON structure",
			"error_details": errInParsingInput.Error()})
		connectContext.Done()
		return
	}

	if isInputJsonValid(atomURLEntry.ShortURL, atomURLEntry.DestinationURL) == false {
		ginContext.JSON(http.StatusNotAcceptable, gin.H{"error": "Empty fields"})
		connectContext.Done()
		return
	}

	atomURLEntry.ShortURL = strings.TrimSpace(atomURLEntry.ShortURL)
	atomURLEntry.DestinationURL = strings.TrimSpace(atomURLEntry.DestinationURL)
	createdAt := time.Now().Unix()
	atomURLEntry.CreatedAt = createdAt

	atomURLEntryToAdd := bson.M{
		"shortURL":       atomURLEntry.ShortURL,
		"destinationURL": atomURLEntry.DestinationURL,
		"created_at":     atomURLEntry.CreatedAt,
	}

	// Checking if short url is taken
	atomURLEntryToSearch := bson.M{
		"shortURL": atomURLEntry.ShortURL,
	}
	errInFinding := dbCollection.FindOne(connectContext, atomURLEntryToSearch).Decode(&atomURLExistingEntry)
	if errInFinding == nil {
		ginContext.JSON(http.StatusConflict, gin.H{"error": "short url already taken",
			"error_details": "Duplicate short url already in database"})
		connectContext.Done()
		return
	}

	// Adding to database
	addedAtomURLEntry, errInAdding := dbCollection.InsertOne(connectContext, atomURLEntryToAdd)
	if errInAdding != nil {
		ginContext.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Error while saving to database", "error_details": errInAdding.Error()})
		connectContext.Done()
		return
	}

	ginContext.JSON(http.StatusCreated, gin.H{"data": addedAtomURLEntry})
	connectContext.Done()
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

	// defining new router
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	router.Static("/asset-manifest.json", "./web/build/asset-manifest.json")
	router.Static("/static", "./web/build/static")

	router.GET("/", func(ginContext *gin.Context) {
		webAppHandler(ginContext)
	})

	router.GET("/go/:shortURL", func(ginContext *gin.Context) {
		redirectURLHandler(ginContext, shortURLsCollection)
	})

	router.POST("/api/add", func(ginContext *gin.Context) {
		addURLHandler(ginContext, shortURLsCollection)
	})

	router.NoRoute(func(ginContext *gin.Context) {
		webAppHandler(ginContext)
	})

	err := router.Run(":" + port)
	if err != nil {
		fmt.Printf("Cannot start server")
	}
}
