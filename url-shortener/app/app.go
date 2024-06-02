package app

import (
	"log"
	"net/http"
	"os"
	"url-shortener/internal/adapters/handler"
	repository "url-shortener/internal/adapters/repository"
	"url-shortener/internal/domain/services"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func StartApp() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal()
	}

	postgresConfig := repository.PostgresConfig{
		Host:            os.Getenv("DB_HOST"),
		User:            os.Getenv("DB_USER"),
		Password:        os.Getenv("DB_PASSWORD"),
		DBName:          os.Getenv("DB_NAME"),
		Port:            5432,
		ApplicationName: os.Getenv("APP_NAME"),
		SSL:             false,
	}

	postgresClient, err := repository.GetPGClient(postgresConfig)

	if err != nil {
		log.Fatal("cant connect to postgres")
	}

	redisConig := repository.RedisConfig{
		Address:  os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		Port:     os.Getenv("REDIS_PORT"),
		DB:       0,
	}

	redisClient, err := repository.GetRedisClient(redisConig)

	if err != nil {
		log.Fatal("cant connect to redis")
	}

	shortenerService := services.NewShortenerService(redisClient, postgresClient)
	router := mux.NewRouter()
	serExpHttpHandler := handler.NewHTTPHandler(
		router, shortenerService,
	)

	serExpHttpHandler.RegisterRoutes()

	serverAddr := ":8081"
	log.Printf("Server started %s\n", serverAddr)
	go func() {
		if err := http.ListenAndServe(serverAddr, router); err != nil {
			log.Fatalf("Error starting the server: %v", err)
		}
	}()

	select {}
}
