package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/models"
	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/sony/gobreaker"
)

func respondError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(models.ErrorResponse{Error: message})
}

// respondServiceError maps service/upstream failures to a consistent JSON error.
func respondServiceError(c *fiber.Ctx, err error) error {
	var upstream *services.UpstreamError
	switch {
	case errors.As(err, &upstream) && upstream.Status == http.StatusNotFound:
		return respondError(c, fiber.StatusNotFound, "Movie not found")
	case errors.Is(err, gobreaker.ErrOpenState), errors.Is(err, gobreaker.ErrTooManyRequests):
		return respondError(c, fiber.StatusServiceUnavailable, "TMDB is temporarily unavailable, please try again later")
	default:
		log.Printf("%s %s failed: %v", c.Method(), c.OriginalURL(), err)
		return respondError(c, fiber.StatusBadGateway, "Failed to fetch data from TMDB")
	}
}

// pageParam reads the optional ?page= query (TMDB accepts 1-500).
func pageParam(c *fiber.Ctx) (int, bool) {
	page, err := strconv.Atoi(c.Query("page", "1"))
	return page, err == nil && page >= 1 && page <= 500
}

// ErrorHandler returns Fiber errors (404 route, panics, ...) as JSON.
func ErrorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	var fe *fiber.Error
	if errors.As(err, &fe) {
		status = fe.Code
	}
	return respondError(c, status, err.Error())
}
