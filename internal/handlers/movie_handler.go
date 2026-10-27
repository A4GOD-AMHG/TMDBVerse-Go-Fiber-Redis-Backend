package handlers

import (
	"log"
	"strconv"
	"strings"

	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

type MovieHandler struct {
	service *services.MovieService
}

func NewMovieHandler(service *services.MovieService) *MovieHandler {
	return &MovieHandler{
		service: service,
	}
}

// @Summary      Get discover movies
// @Description  Get movies from discover endpoint sorted by popularity
// @Tags         movies
// @Produce      json
// @Param        page  query     int  false  "Page number (1-500)"
// @Success      200   {array}   models.Movie
// @Failure      400   {object}  models.ErrorResponse  "Invalid page"
// @Failure      502   {object}  models.ErrorResponse  "TMDB error"
// @Router       /discover [get]
func (h *MovieHandler) DiscoverMovies(c *fiber.Ctx) error {
	page, ok := pageParam(c)
	if !ok {
		return respondError(c, fiber.StatusBadRequest, "Query parameter 'page' must be a number between 1 and 500")
	}

	movies, err := h.service.GetDiscoverMovies(page)
	if err != nil {
		return respondServiceError(c, err)
	}

	return c.JSON(movies)
}

// @Summary      Get top popular movies
// @Description  Get the 3 most popular movies
// @Tags         movies
// @Produce      json
// @Success      200  {array}   models.Movie
// @Failure      502  {object}  models.ErrorResponse  "TMDB error"
// @Router       /popular [get]
func (h *MovieHandler) TopPopularMovies(c *fiber.Ctx) error {
	movies, err := h.service.GetTopPopularMovies()
	if err != nil {
		return respondServiceError(c, err)
	}

	return c.JSON(movies)
}

// @Summary      Search movies
// @Description  Search movies by title
// @Tags         movies
// @Produce      json
// @Param        q     query     string  true   "Search query"
// @Param        page  query     int     false  "Page number (1-500)"
// @Success      200   {array}   models.Movie
// @Failure      400   {object}  models.ErrorResponse  "Missing query or invalid page"
// @Failure      502   {object}  models.ErrorResponse  "TMDB error"
// @Router       /search [get]
func (h *MovieHandler) SearchMovies(c *fiber.Ctx) error {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		return respondError(c, fiber.StatusBadRequest, "Search query parameter 'q' is required")
	}

	page, ok := pageParam(c)
	if !ok {
		return respondError(c, fiber.StatusBadRequest, "Query parameter 'page' must be a number between 1 and 500")
	}

	movies, err := h.service.SearchMovies(query, page)
	if err != nil {
		return respondServiceError(c, err)
	}

	return c.JSON(movies)
}

// @Summary      Get trending movies
// @Description  Get top 5 trending movies based on searches
// @Tags         movies
// @Produce      json
// @Success      200  {array}   models.Movie
// @Failure      500  {object}  models.ErrorResponse  "Redis error"
// @Router       /trending [get]
func (h *MovieHandler) TrendingMovies(c *fiber.Ctx) error {
	movies, err := h.service.GetTrendingMovies()
	if err != nil {
		log.Printf("trending movies failed: %v", err)
		return respondError(c, fiber.StatusInternalServerError, "Failed to load trending movies")
	}
	return c.JSON(movies)
}

// @Summary      Get movie details
// @Description  Get complete details for a specific movie
// @Tags         movies
// @Produce      json
// @Param        id   path      int  true  "Movie ID"
// @Success      200  {object}  models.Movie
// @Failure      400  {object}  models.ErrorResponse  "Invalid ID"
// @Failure      404  {object}  models.ErrorResponse  "Movie not found"
// @Failure      502  {object}  models.ErrorResponse  "TMDB error"
// @Router       /movies/{id} [get]
func (h *MovieHandler) MovieDetails(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return respondError(c, fiber.StatusBadRequest, "Invalid ID")
	}

	movie, err := h.service.GetMovieDetails(id)
	if err != nil {
		return respondServiceError(c, err)
	}

	return c.JSON(movie)
}
