package handlers

import (
	"github.com/A4GOD-AMHG/TMDBZone-Go-Fiber-Backend/internal/config"
	"github.com/A4GOD-AMHG/TMDBZone-Go-Fiber-Backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

type MovieHandler struct {
	service *services.MovieService
}

func NewMovieHandler(cfg *config.Config) *MovieHandler {
	return &MovieHandler{
		service: services.NewMovieService(cfg),
	}
}

// @Summary      Get discover movies
// @Description  Get movies from discover endpoint
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        page  query     string  false  "Page number"
// @Success      200   {object}  []models.Movie
// @Router       /discover [get]
func (h *MovieHandler) DiscoverMovies(c *fiber.Ctx) error {
	page := c.Query("page", "1")

	movies, err := h.service.GetDiscoverMovies(page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(movies)
}

// @Summary      Get top popular movies
// @Description  Get 3 most popular movies
// @Tags         movies
// @Accept       json
// @Produce      json
// @Success      200  {object}  []models.Movie
// @Router       /popular [get]
func (h *MovieHandler) TopPopularMovies(c *fiber.Ctx) error {
	movies, err := h.service.GetTopPopularMovies()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(movies)
}

// @Summary      Search movies
// @Description  Search movies by title
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        q     query     string  true   "Search query"
// @Param        page  query     string  false  "Page number"
// @Success      200   {object}  []models.Movie
// @Router       /search [get]
func (h *MovieHandler) SearchMovies(c *fiber.Ctx) error {
	query := c.Query("q")
	page := c.Query("page", "1")

	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Search query parameter 'q' is required",
		})
	}

	movies, err := h.service.SearchMovies(query, page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(movies)
}
