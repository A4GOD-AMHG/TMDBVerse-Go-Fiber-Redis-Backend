package services

import "fmt"

// UpstreamError is returned when TMDB answers with a non-200 status.
type UpstreamError struct {
	Status int
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("TMDB responded with status %d", e.Status)
}
