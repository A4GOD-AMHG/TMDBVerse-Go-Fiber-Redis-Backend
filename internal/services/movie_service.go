package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/config"
	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/models"
)

const (
	apiBaseURL = "https://api.themoviedb.org/3"
)

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     30 * time.Second,
	},
}

type MovieService struct {
	cfg   *config.Config
	cache *CacheService
}

func NewMovieService(cfg *config.Config, cache *CacheService) *MovieService {
	return &MovieService{
		cfg:   cfg,
		cache: cache,
	}
}

func (s *MovieService) makeRequest(url string) ([]byte, error) {
	result, err := cb.Execute(func() (interface{}, error) {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.cfg.AccessToken))

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		return io.ReadAll(resp.Body)
	})

	if err != nil {
		return nil, err
	}

	return result.([]byte), nil
}

func (s *MovieService) SearchMovies(query string, page string) ([]models.Movie, error) {
	cacheKey := fmt.Sprintf("search:%s:%s", query, page)
	if cached, err := s.cache.Get(cacheKey); err == nil {
		var movies []models.Movie
		if json.Unmarshal(cached, &movies) == nil {
			return movies, nil
		}
	}

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

	if data, err := json.Marshal(response.Results); err == nil {
		s.cache.Set(cacheKey, data, time.Hour)
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
