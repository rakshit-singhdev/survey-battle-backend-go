package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"survey-battle-backend-go/internal/config"
	"survey-battle-backend-go/internal/database"
	"survey-battle-backend-go/internal/redis"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MongoDB
	mongoDB, err := database.ConnectMongo(context.Background(), cfg.MongoURI)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := mongoDB.Disconnect(shutdownCtx); err != nil {
			log.Printf("MongoDB disconnect error: %v", err)
		}
	}()

	log.Println("MongoDB connected")

	// Connect to Redis
	redisStore, err := redis.Connect(context.Background(), cfg.RedisURL, cfg.Redis)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := redisStore.Disconnect(); err != nil {
			log.Printf("Redis disconnect error: %v", err)
		}
	}()

	log.Println("Redis connected")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"message":"Family Feud Go backend"}`))
	})

	log.Printf("Go backend listening on :%s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatal(err)
	}
}
