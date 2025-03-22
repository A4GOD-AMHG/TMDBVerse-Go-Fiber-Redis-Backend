package services

import (
	"time"

	"github.com/sony/gobreaker"
)

var cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
	Name:        "TMDB-API",
	MaxRequests: 5,
	Interval:    10 * time.Second,
	Timeout:     15 * time.Second,
	ReadyToTrip: func(counts gobreaker.Counts) bool {
		return counts.ConsecutiveFailures > 5
	},
})
