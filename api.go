package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Film struct {
	EpisodeID     int    `json:"episode_id"`
	Title         string `json:"title"`
	OpeningCrawl  string `json:"opening_crawl"`
	ReleaseDate   string `json:"release_date"`
	TotalComments int    `json:"total_comments"`
}

type Comment struct {
	Content   string    `json:"content"`
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
	MovieID   string    `json:"movie_id"`
}

type Character struct {
	Name         string `json:"name"`
	Gender       string `json:"gender"`
	BirthYear    string `json:"birth_year"`
	HairColor    string `json:"hair_color"`
	Height       string `json:"height"`
	HeightInFeet string `json:"height_in_feet"`
}

type getFilmsResponse struct {
	Count   int    `json:"count"`
	Results []Film `json:"results"`
}

type getCommentsResponse struct {
	Count    int       `json:"count"`
	Comments []Comment `json:"comments"`
}

type postCommentRequest struct {
	MovieID int    `json:"movie_id"`
	Comment string `json:"comment"`
}

type getCharactersResponse struct {
	Count        int         `json:"count"`
	Height       string      `json:"height"`
	HeightInFeet string      `json:"height_in_feet"`
	Characters   []Character `json:"characters"`
}

// @tags		Film
// @Summary     Get films list
// @Description Fetch list of movies and their details in chronological order.
// @Accept      json
// @Produce     json
// @Success     200 {object} getFilmsResponse
// @Failure     500 {object} errorResponse
// @Router      /api/films [get]
func getFilms(res http.ResponseWriter, req *http.Request, conn Connect) (any, error) {
	// fetch movies
	val, err := makeSWAPICall("films", conn)
	if err != nil {
		return "Error fetching films from SWAPI", err
	}

	var films getFilmsResponse
	err = json.Unmarshal(val, &films)
	if err != nil {
		return "Error unmarshaling films response:", err
	}

	var commentsCount int
	for i, f := range films.Results {
		rows, err := conn.db.Query(`SELECT count(*) FROM comments WHERE movie_id=` + strconv.Itoa(f.EpisodeID))
		if err != nil {
			log.Println("Error reading comments from DB.")
			commentsCount = -1
		} else {
			defer rows.Close()
			if rows.Next() {
				err = rows.Scan(&commentsCount)
			} else {
				commentsCount = 0
			}
		}
		films.Results[i].TotalComments = commentsCount
	}

	sort.Slice(films.Results, func(i, j int) bool {
		return films.Results[i].ReleaseDate < films.Results[j].ReleaseDate
	})

	films.Count = len(films.Results)

	return films, nil
}

// @tags		Comment
// @Summary     Get all comments for a movie.
// @Description Fetch comments for a specific movie, provide no 'movie_id' if comments for all movies needed
// @Accept      json
// @Produce     json
// @Param       movie_id   path      int  true  "Movie ID"
// @Success     200 {object} getCommentsResponse
// @Failure     500 {object} errorResponse
// @Router      /api/comments/{movie_id} [get]
func getComments(w http.ResponseWriter, r *http.Request, connect Connect) (any, error) {
	id := strings.TrimPrefix(r.URL.Path, "/api/comments/")
	log.Println(id)
	var comments getCommentsResponse
	var query string
	if id == "" {
		// all
		query = "SELECT * FROM comments ORDER BY timestamp DESC"
	} else {
		query = fmt.Sprintf(`SELECT * FROM comments 
			WHERE movie_id=%s ORDER BY timestamp DESC `, id)
	}
	rows, err := connect.db.Query(query)
	if err != nil {
		log.Printf("Error reading comments from DB.")
		return "Internal error", err
	}
	defer rows.Close()
	for rows.Next() {
		var c Comment
		var id int
		err := rows.Scan(&id, &c.MovieID, &c.Content, &c.IP, &c.Timestamp)
		if err != nil {
			log.Printf("Error reading comment from DB.")
			return "Server-side Error", err
		}
		comments.Comments = append(comments.Comments, c)
	}
	comments.Count = len(comments.Comments)

	return comments, nil
}

// @tags		Comments
// @Summary     Post comment on a movie
// @Description Submit comment for a movie identified by 'movie_id' in request body.
// @Accept      json
// @Produce     json
// @Param       body body postCommentRequest true "Movie ID"
// @Success     200 {object} getCommentsResponse
// @Failure     500 {object} errorResponse
// @Router      /api/comment/ [post]
func postComment(w http.ResponseWriter, r *http.Request, connect Connect) (any, error) {
	decoder := json.NewDecoder(r.Body)

	var comment postCommentRequest
	err := decoder.Decode(&comment)
	if err != nil {
		log.Printf("Comment decode error.")
		return "Server-side Error", err
	}
	if len(comment.Comment) > 500 {
		return "Comment must be less than 500 characters", err
	}
	if comment.MovieID == 0 || comment.Comment == "" {
		return "movie_id or comment missing", err
	}
	ip := r.RemoteAddr

	_, err = connect.db.Exec(fmt.Sprintf(`INSERT INTO comments (movie_id, comment, commenter_ip, timestamp)
		VALUES (%d, '%s', '%s', (NOW() AT TIME ZONE 'utc'));`, comment.MovieID, comment.Comment, ip))
	if err != nil {
		log.Printf("Comment insertion error.")
		return "Server-side Error", err
	}
	response := struct {
		message string
		ip      string
	}{
		"Comment successfully submitted.",
		ip,
	}
	return response, nil
}

// @tags		Character
// @Summary     Fetch movie characters.
// @Description Fetch characters of a movie specified by ID along with their combined height.
// @Accept      json
// @Produce     json
// @Param       movie_id  path     int true "Movie ID"
// @Param       sort  query     string false "Supported values: 'name', 'gender', or 'height', use '-' prefix for descending order."
// @Param       filter  query     string false "Filter by gender"
// @Success     200 {object} getCharactersResponse
// @Failure     500 {object} errorResponse
// @Router      /api/characters/{movie_id} [get]
func getCharacters(w http.ResponseWriter, r *http.Request, connect Connect) (any, error) {
	id := strings.TrimPrefix(r.URL.Path, "/api/characters/")
	sortParam := r.URL.Query().Get("sort")
	filterParam := r.URL.Query().Get("filter")

	resp, err := makeSWAPICall("films/"+id, connect)
	if err != nil {
		return "Error fetching film", err
	}

	var film struct{ Characters []string }
	err = json.Unmarshal(resp, &film)

	if err != nil {
		return "Error in films data", err
	}

	characters, totalHeight, err := fetchCharacters(film.Characters, filterParam, connect)
	if err != nil {
		return "Error fetching characters.", err
	}
	sortCharacters(sortParam, characters)

	var response getCharactersResponse
	response.Count = len(characters)
	response.Height = fmt.Sprintf("%d", int(totalHeight))
	response.HeightInFeet, _ = heightInFeet(response.Height)
	response.Characters = characters

	return response, nil
}
