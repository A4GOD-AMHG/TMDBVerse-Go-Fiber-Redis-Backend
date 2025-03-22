package main

import (
	_ "github.com/A4GOD-AMHG/TMDBZone-Go-Fiber-Backend/docs"
	"github.com/A4GOD-AMHG/TMDBZone-Go-Fiber-Backend/internal/config"
	"github.com/A4GOD-AMHG/TMDBZone-Go-Fiber-Backend/internal/handlers"

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

	movieHandler := handlers.NewMovieHandler(cfg)

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET",
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	app.Get("/discover", movieHandler.DiscoverMovies)
	app.Get("/popular", movieHandler.TopPopularMovies)
	app.Get("/search", movieHandler.SearchMovies)

	app.Listen(":8080")
}
