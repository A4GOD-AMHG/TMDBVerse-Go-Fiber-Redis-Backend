package main

import (
	"context"
	"log"
	"os"
	"time"

	_ "github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/docs"
	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/config"
	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/handlers"
	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/services"
	"github.com/go-redis/redis/v8"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
)

// @title           TMDBZone-Go-Fiber-Backend
// @version         1.0
// @description     API proxy for The Movie DB
// @host            localhost:8080
// @BasePath        /
func main() {
	cfg := config.LoadConfig()
	if cfg.AccessToken == "" {
		panic("TMDB_API_ACCESS_TOKEN environment variable is required")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0,
	})

	log.SetOutput(os.Stdout)

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}

	log.Println("Connected to Redis successfully")

	cacheService := services.NewCacheService(rdb)
	movieService := services.NewMovieService(cfg, cacheService, rdb)
	movieHandler := handlers.NewMovieHandler(movieService)

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET",
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Get("/discover", handlers.CacheMiddleware(cacheService, 10*time.Minute), movieHandler.DiscoverMovies)
	app.Get("/popular", handlers.CacheMiddleware(cacheService, 30*time.Minute), movieHandler.TopPopularMovies)
	app.Get("/search", handlers.CacheMiddleware(cacheService, 1*time.Hour), movieHandler.SearchMovies)
	app.Get("/trending", handlers.CacheMiddleware(cacheService, 30*time.Minute), movieHandler.TrendingMovies)

	app.Listen(":8080")
}
