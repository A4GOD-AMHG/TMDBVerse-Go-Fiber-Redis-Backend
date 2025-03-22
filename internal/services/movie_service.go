package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/A4GOD-AMHG/TMDBZone-Go-Fiber-Backend/internal/config"
	"github.com/A4GOD-AMHG/TMDBZone-Go-Fiber-Backend/internal/models"
)

const (
	apiBaseURL = "https://api.themoviedb.org/3"
)

type MovieService struct {
	cfg *config.Config
}

func NewMovieService(cfg *config.Config) *MovieService {
	return &MovieService{cfg: cfg}
}

func (s *MovieService) makeRequest(url string) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.cfg.AccessToken))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (s *MovieService) SearchMovies(query string, page string) ([]models.Movie, error) {
	url := fmt.Sprintf(
		"%s/search/movie?query=%s&include_adult=false&language=en-US&page=%s",
		apiBaseURL,
		url.QueryEscape(query),
		page,
	)

	body, err := s.makeRequest(url)
	if err != nil {
		return nil, err
	}

	var response models.MovieResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Results, nil
}

func (s *MovieService) GetDiscoverMovies(page string) ([]models.Movie, error) {
	url := fmt.Sprintf(
		"%s/discover/movie?include_adult=false&include_video=false&language=en-US&page=%s&sort_by=popularity.desc",
		apiBaseURL,
		page,
	)

	body, err := s.makeRequest(url)
	if err != nil {
		return nil, err
	}

	var response models.MovieResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Results, nil
}

func (s *MovieService) GetTopPopularMovies() ([]models.Movie, error) {
	url := fmt.Sprintf("%s/movie/popular?language=en-US&page=1", apiBaseURL)

	body, err := s.makeRequest(url)
	if err != nil {
		return nil, err
	}

	var response models.MovieResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if len(response.Results) > 3 {
		return response.Results[:3], nil
	}
	return response.Results, nil
}
