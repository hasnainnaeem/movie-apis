package main

import (
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"log"
)

// loadConfig reads configuration from file or environment variables.
func loadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return config, err
}

func getPostgresDB(config Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		config.DBHost, config.DBPort, config.DBUser, config.DBPass,
		config.DBName, config.DBSSLMode)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	var doesCommentsTableExist string
	// create comments table if it doesn't exist
	if err := db.QueryRow(`SELECT EXISTS (SELECT FROM information_schema.tables 
       WHERE  table_name   = 'film_comments')`).Scan(&doesCommentsTableExist); err != nil {
		return nil, err
	}
	if doesCommentsTableExist == "false" {
		log.Println("Creating table")
		_, err := db.Exec(`CREATE TABLE film_comments (id SERIAL, comment VARCHAR(500), commenter_ip VARCHAR(40), timestamp TIMESTAMP, movie_id INT);`)
		if err != nil {
			return nil, err
		}
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, err
}

func getRedisClient(config Config) (*redis.Client, error) {
	// redis

	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisHost,
		Password: config.RedisPassword,
		DB:       0,
	})

	var err error
	_, err = client.Ping(client.Context()).Result()
	if err != nil {
		return nil, err
	}

	return client, err
}
