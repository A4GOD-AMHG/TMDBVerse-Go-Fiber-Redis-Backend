package models

type Movie struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	Overview         string  `json:"overview"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	OriginalLanguage string  `json:"original_language"`
	VoteAverage      float64 `json:"vote_average"`
}

type TrendingMovie struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	PosterPath  string `json:"poster_path"`
	SearchCount int    `json:"search_count"`
}

type MovieResponse struct {
	Page       int     `json:"page"`
	Results    []Movie `json:"results"`
	TotalPages int     `json:"total_pages"`
}
