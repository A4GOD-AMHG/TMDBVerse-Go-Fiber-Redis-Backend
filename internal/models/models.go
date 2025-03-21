package models

type Movie struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	PosterPath  string  `json:"poster_path"`
	ReleaseDate string  `json:"release_date"`
	Popularity  float64 `json:"popularity"`
}

type MovieResponse struct {
	Page       int     `json:"page"`
	Results    []Movie `json:"results"`
	TotalPages int     `json:"total_pages"`
}
