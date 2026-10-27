package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/docs"
	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/config"
	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/handlers"
	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/services"
	"github.com/go-redis/redis/v8"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

// @title           TMDBVerse-Go-Fiber-Redis-Backend
// @version         1.1.0
// @description     API proxy for The Movie DB with Redis caching
// @host            localhost:8080
// @BasePath        /
func main() {
	log.SetOutput(os.Stdout)

	cfg := config.LoadConfig()
	if cfg.AccessToken == "" {
		log.Fatal("TMDB_API_ACCESS_TOKEN environment variable is required")
	}

	rdb, err := newRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Invalid REDIS_URL: %v", err)
	}
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")

	cacheService := services.NewCacheService(rdb)
	movieService := services.NewMovieService(cfg, cacheService, rdb)
	movieHandler := handlers.NewMovieHandler(movieService)

	app := fiber.New(fiber.Config{ErrorHandler: handlers.ErrorHandler})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET",
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Get("/discover", handlers.CacheMiddleware(cacheService, 10*time.Minute), movieHandler.DiscoverMovies)
	app.Get("/popular", handlers.CacheMiddleware(cacheService, 30*time.Minute), movieHandler.TopPopularMovies)
	app.Get("/movies/:id", handlers.CacheMiddleware(cacheService, 30*time.Minute), movieHandler.MovieDetails)
	app.Get("/search", movieHandler.SearchMovies)
	app.Get("/trending", movieHandler.TrendingMovies)

	log.Fatal(app.Listen(":" + cfg.Port))
}

// newRedisClient accepts both "redis://host:port/db" URLs and plain "host:port" addresses.
func newRedisClient(redisURL string) (*redis.Client, error) {
	if strings.Contains(redisURL, "://") {
		opts, err := redis.ParseURL(redisURL)
		if err != nil {
			return nil, err
		}
		return redis.NewClient(opts), nil
	}
	return redis.NewClient(&redis.Options{Addr: redisURL}), nil
}
