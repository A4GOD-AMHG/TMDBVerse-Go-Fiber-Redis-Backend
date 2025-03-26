package models

type Movie struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	Overview         string  `json:"overview"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	OriginalLanguage string  `json:"original_language"`
	VoteAverage      float64 `json:"vote_average"`
	SearchCount      int     `json:"search_count,omitempty"`
}

type MovieResponse struct {
	Page       int     `json:"page"`
	Results    []Movie `json:"results"`
	TotalPages int     `json:"total_pages"`
}
