package handlers

import (
	"time"

	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

func CacheMiddleware(cache *services.CacheService, ttl time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cacheKey := c.OriginalURL()

		if cached, err := cache.Get(cacheKey); err == nil {
			c.Response().Header.SetContentType("application/json")
			return c.Send(cached)
		}

		err := c.Next()

		if err == nil && c.Response().StatusCode() == 200 {
			cache.Set(cacheKey, c.Response().Body(), ttl)
		}

		return err
	}
}
