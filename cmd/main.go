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

	app := fiber.New()

	cacheService := services.NewCacheService(os.Getenv("REDIS_URL"))

	ctx := context.Background()
	_, err := cacheService.Client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	log.Println("Connected to Redis successfully")
	movieService := services.NewMovieService(cfg, cacheService)
	movieHandler := handlers.NewMovieHandler(movieService)

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET",
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Get("/discover", handlers.CacheMiddleware(cacheService, 10*time.Minute), movieHandler.DiscoverMovies)
	app.Get("/popular", handlers.CacheMiddleware(cacheService, 30*time.Minute), movieHandler.TopPopularMovies)
	app.Get("/search", handlers.CacheMiddleware(cacheService, 1*time.Hour), movieHandler.SearchMovies)

	app.Listen(":8080")
}
