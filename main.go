package main

import (
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"
)

type APIHandler struct {
	endpoint func(http.ResponseWriter, *http.Request, Connect) (any, error)
}

// @title APIs for Movie Details
// @version 1.0
// @description API to get details of movies and to comment on them.
// @contact.name Hasnain Naeem
// @BasePath /
func main() {
	// API endpoints
	// get
	http.Handle("/api/films", APIHandler{getFilms})
	http.Handle("/api/characters/", APIHandler{getCharacters})
	http.Handle("/api/comments/", APIHandler{getComments})
	// post
	http.Handle("/api/comment", APIHandler{postComment})

	// docs API
	http.Handle("/docs/", httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.json")))
	http.HandleFunc("/docs/swagger.json", swaggerFiles)
	
	// serve app
	port := os.Getenv("PORT")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
