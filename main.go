package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AtomURLEntry structure of the a single item in database
type AtomURLEntry struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	ShortURL       string             `json:"shortURL" bson:"shortURL"`
	DestinationURL string             `json:"destinationURL" bson:"destinationURL"`
	CreatedAt      int64              `json:"created_at" bson:"created_at"`
}

func getEnvValues(envKeyStrings [5]string) map[string]string {

	envValues := make(map[string]string)

	for _, keyString := range envKeyStrings {
		if os.Getenv(keyString) == "" {
			log.Fatal("No env value provided for " + keyString + " , check Readme")

		}
		envValues[keyString] = os.Getenv(keyString)
	}

	return envValues
}

func connectToDatabase(mangoDatabaseURL string) *mongo.Client {
	databaseURL := fmt.Sprint(mangoDatabaseURL)
	connectOptions := options.Client()
	connectOptions.ApplyURI(databaseURL)

	connectContext, errorInContext := context.WithTimeout(context.Background(), 10*time.Second)

	defer errorInContext()

	databaseClient, errInConnection := mongo.Connect(connectContext, connectOptions)

	if errInConnection != nil {
		log.Fatal("Failed to connect to DB", errInConnection)
	}

	errInPing := databaseClient.Ping(connectContext, nil)

	if errInPing != nil {
		panic("DB not found")
	}

	return databaseClient
}

func areJSONFieldsMissing(shortURL string, longURL string) error {
	lengthOfShortURL := len(strings.TrimSpace(shortURL))
	lengthOfLongURL := len(strings.TrimSpace(longURL))

	if lengthOfShortURL == 0 || lengthOfLongURL == 0 {
		if lengthOfLongURL == 0 {
			return fmt.Errorf("destination url not provided or field missing")
		}
		return fmt.Errorf("short url not provided or field missing")
	}
	return nil
}

func isDestinationURLValid(url *url.URL) error {
	if url.IsAbs() == true {
		// check if its localhost or other with port
		if url.Port() == "" {
			// check if it has http or https
			if url.Scheme == "http" || url.Scheme == "https" {
				// check if it has user name or password
				if url.User == nil {
					hostname := strings.ToLower(url.Host)
					atomurlSubdomain := "www.atomurl.ga"
					atom2Root := "atomurl.ga"
					if (hostname != atom2Root) && (hostname != atomurlSubdomain) {
						return nil
					}
					return fmt.Errorf("cannot contain atomurl")
				}
				return fmt.Errorf("username in url")
			}
			return fmt.Errorf("not a valid scheme")
		}
		return fmt.Errorf("should not have port")
	}
	return fmt.Errorf("doesnt have https or http")
}

func webAppHandler(ginContext *gin.Context) {
	ginContext.File("./web/build/index.html")
}

func notFoundHandler(ginContext *gin.Context) {
	ginContext.Redirect(http.StatusMovedPermanently, "/404")
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
		notFoundHandler(ginContext)
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

	// Valid JSON check
	errInParsingInput := ginContext.ShouldBindJSON(&atomURLEntry)
	if errInParsingInput != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Wrong JSON structure",
			"error_details": errInParsingInput.Error()})
		connectContext.Done()
		return
	}

	// Empty check
	missingJSONFields := areJSONFieldsMissing(atomURLEntry.ShortURL, atomURLEntry.DestinationURL)
	if missingJSONFields != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Empty fields", "error_details": missingJSONFields.Error()})
		connectContext.Done()
		return
	}

	destinationURL := strings.ToLower(strings.TrimSpace(atomURLEntry.DestinationURL))

	// Only destination URL checks
	decodedDestinationURL, errInDecoding := url.Parse(destinationURL)
	if errInDecoding != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Error while decoding URL",
			"error_details": errInDecoding.Error()})
		connectContext.Done()
		return
	}

	errInDestinationURL := isDestinationURLValid(decodedDestinationURL)
	if errInDestinationURL != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Destination URL not in correct format",
			"error_details": errInDestinationURL.Error()})
		connectContext.Done()
		return
	}

	shortURL := strings.ToLower(strings.TrimSpace(atomURLEntry.ShortURL))

	// Only short url checks
	validShortURLRegex := "^([a-z-]+$)"
	shortURLRegTester, errorInURLRegex := regexp.Compile(validShortURLRegex)
	if errorInURLRegex != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Short URL not in correct format",
			"error_details": errorInURLRegex.Error()})
		connectContext.Done()
		return
	}

	if shortURLRegTester.MatchString(shortURL) == false {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Short URL not in correct format",
			"error_details": "Short url did not match regex"})
		connectContext.Done()
		return
	}

	shortURLFirstChar := string(shortURL[0])
	shortURLLastChar := string(shortURL[len(shortURL)-1])
	if shortURLFirstChar == "-" || shortURLLastChar == "-" {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Short URL not in correct format",
			"error_details": "Short url cannot start or end wit hyphen"})
		connectContext.Done()
		return
	}

	createdAt := time.Now().Unix()
	atomURLEntry.CreatedAt = createdAt

	atomURLEntryToAdd := bson.M{
		"shortURL":       shortURL,
		"destinationURL": destinationURL,
		"created_at":     createdAt,
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
		port = "8001"
	}

	envKeys := [5]string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_URL", "DB_NAME"}
	envValues := getEnvValues(envKeys)

	databaseURL := fmt.Sprint(envValues["DB_HOST"], "://", envValues["DB_USER"], ":", envValues["DB_PASSWORD"], "@", envValues["DB_URL"], "/", envValues["DB_NAME"])

	database := connectToDatabase(databaseURL)
	shortURLsCollection := database.Database("atom-url-db").Collection("shorturls")

	// defining new router
	router := gin.Default()

	defaultCors := cors.DefaultConfig()

	defaultCors.AllowOrigins = []string{"https://atomurl.ga", "https://atomurlapp.herokuapp.com", "http://localhost:3000"}
	router.Use(cors.New(defaultCors))

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

	router.GET("/404", func(ginContext *gin.Context) {
		webAppHandler(ginContext)
	})

	router.NoRoute(func(ginContext *gin.Context) {
		notFoundHandler(ginContext)
	})

	err := router.Run(":" + port)
	if err != nil {
		log.Fatal("Cannot start server")
	}
}
