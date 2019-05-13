package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func welcomePage(ginContext *gin.Context) {
	ginContext.JSON(200, gin.H{"name": "Hello"})
}

func redirectURLHandler(ginContext *gin.Context, shortURLMaps map[string]string) {

	shortURL := ginContext.Param("shortURL")

	shortURLMappedDest, found := shortURLMaps[shortURL]
	if found {
		ginContext.Redirect(http.StatusFound, shortURLMappedDest)
	}

	ginContext.JSON(http.StatusNotFound, gin.H{"status": "Not found"})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8000"
	}

	// defining new router
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/", func(ginContext *gin.Context) {
		welcomePage(ginContext)
	})

	shortURLMappings := make(map[string]string)
	shortURLMappings["zubair"] = "https://mzubairahmed.ml/"

	router.GET("/:shortURL", func(ginContext *gin.Context) {
		redirectURLHandler(ginContext, shortURLMappings)
	})

	err := router.Run(port)
	if err != nil {
		fmt.Printf("Cannot start server")
	}
}
