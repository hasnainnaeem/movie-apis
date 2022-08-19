# Go Assessment Task
- API endpoints to fetch movie details and to comment on movies. 
- Details are fetched from `https://swapi.dev`and cached using `redis`. 
- Comments are saved in postgres database.
- **Demo link:** https://apis-for-movies.herokuapp.com/
- **API endpoints and their details:** https://apis-for-movies.herokuapp.com/api/docs

## Setup and Running 
- **Run app:** run `go run .` in the root directory. 
- **Config variables:** `app.env* file contains all the config variables including `redis` and `postgres` credentials. The port variable is accessed through environment variable named `PATH`, use `config.AppPort` in `main.go` to access it from the config file.

## Task Details
Create a small set of rest API endpoints using Golang to do the following
- List an array of movies containing the name, opening crawl and comment count
- Add a new comment for a movie
- List the comments for a movie
- Get list of characters for a movie

More details: https://incredible-passenger-16e.notion.site/Assessment-2e389b0d9ccd4b9aa0494449b234c2ca
