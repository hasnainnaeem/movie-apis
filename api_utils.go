package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"


	"github.com/go-redis/redis/v8"
)

type errorResponse struct {
	Message   string `json:"message"`
	ErrorCode int    `json:"error_code"`
}

type Connect struct {
	db     *sql.DB
	client *redis.Client
}

var conn Connect

func (handle APIHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var err error
	var config Config
	config, err = loadConfig(".")
	if err != nil {
		log.Fatal("Error loading config file:", err)
	}

	// get db and redis client
	var db *sql.DB
	var client *redis.Client
	db, err = getPostgresDB(config)
	if err != nil {
		panic(err)
	}
	client, err = getRedisClient(config)
	if err != nil {
		panic(err)
	}

	conn = Connect{db, client}

	resp, err := handle.endpoint(res, req, conn)

	if err != nil {
		var response errorResponse
		response.Message = resp.(string)
		response.ErrorCode = 500
		out, _ := json.Marshal(response)

		log.Println(err)

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(response.ErrorCode)
		_, err := res.Write(out)
		if err != nil {
			return
		}
	} else {
		res.Header().Set("Content-Type", "application/json")
		out, _ := json.Marshal(resp)
		_, err := res.Write(out)
		if err != nil {
			return
		}
	}
}

func makeSWAPICall(endpoint string, conn Connect) ([]byte, error) {
	var res []byte

	// check if exists in cache
	val, err := conn.client.Get(context.Background(), endpoint).Result()

	if err == redis.Nil {
		resp, err := http.Get("https://swapi.dev/api/" + endpoint)
		if err != nil {
			return nil, err
		}

		res, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		// store in cache
		err = conn.client.Set(context.Background(), endpoint, string(res), 48*time.Hour).Err()
	} else if err != nil {
		return res, nil
	} else {
		res = []byte(val)
	}

	return res, nil
}

func heightInFeet(heightInCm string) (string, error) {
	cms, err := strconv.ParseFloat(heightInCm, 64)
	if err != nil {
		return "-1", err
	}
	// convert to %d feet %d inches string format
	feet := math.Floor(cms / 30.48)
	inches := (cms / 2.54) - (feet * 12)
	return fmt.Sprintf("%d feet %0.2f inches", int(feet), inches), nil
}

func sortCharacters(sortParam string, characters []Character) {
	isOrderAsc := true
	if strings.HasPrefix(sortParam, "-") {
		isOrderAsc = false
		sortParam = strings.TrimPrefix(sortParam, "-")
	}

	// sort w.r.t param and in given order
	sort.Slice(characters, func(i, j int) bool {
		var order bool

		if sortParam == "name" {
			order = characters[i].Name < characters[j].Name
		} else if sortParam == "height" {
			order = characters[i].Height < characters[j].Height
		} else if sortParam == "gender" {
			order = characters[i].Gender < characters[j].Gender
		}
		return isOrderAsc == order
	})
}

func fetchCharacters(charactersLinks []string, filterParam string, conn Connect) ([]Character, float64, error) {
	var characters []Character
	var totalHeight float64

	for _, url := range charactersLinks {
		// fetch
		characterId := strings.TrimPrefix(url, "https://swapi.dev/api/people/")
		resp, _ := makeSWAPICall("people/"+characterId, conn)

		var character Character
		if err := json.Unmarshal(resp, &character); err != nil {
			return characters, totalHeight, err
		}
		character.HeightInFeet, _ = heightInFeet(character.Height)

		if filterParam == "" || character.Gender == filterParam {
			characters = append(characters, character)

			height, _ := strconv.ParseFloat(character.Height, 64)
			totalHeight += height
		}
	}
	return characters, totalHeight, nil
}

func swaggerFiles(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./docs/swagger.json")
}
