package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/config"
	"github.com/A4GOD-AMHG/TMDBVerse-Go-Fiber-Redis-Backend/internal/models"
	"github.com/go-redis/redis/v8"
)

const (
	apiBaseURL    = "https://api.themoviedb.org/3"
	searchZSetKey = "movie_searches"
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
	redis *redis.Client
}

func NewMovieService(cfg *config.Config, cache *CacheService, rdb *redis.Client) *MovieService {
	return &MovieService{
		cfg:   cfg,
		cache: cache,
		redis: rdb,
	}
}

type upstreamResponse struct {
	status int
	body   []byte
}

func (s *MovieService) makeRequest(url string) ([]byte, error) {
	result, err := cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", "Bearer "+s.cfg.AccessToken)

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		// Only 5xx count as breaker failures; 4xx (e.g. unknown movie) are client errors.
		if resp.StatusCode >= http.StatusInternalServerError {
			return nil, &UpstreamError{Status: resp.StatusCode}
		}
		return upstreamResponse{status: resp.StatusCode, body: body}, nil
	})
	if err != nil {
		return nil, err
	}

	res := result.(upstreamResponse)
	if res.status != http.StatusOK {
		return nil, &UpstreamError{Status: res.status}
	}
	return res.body, nil
}

func (s *MovieService) fetchMovieList(url string) ([]models.Movie, error) {
	body, err := s.makeRequest(url)
	if err != nil {
		return nil, err
	}

	var response models.MovieResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	if response.Results == nil {
		return []models.Movie{}, nil
	}
	return response.Results, nil
}

func (s *MovieService) logSearch(movie models.Movie) {
	ctx := context.Background()
	if err := s.redis.ZIncrBy(ctx, searchZSetKey, 1, strconv.Itoa(movie.ID)).Err(); err != nil {
		log.Printf("error increasing search counter: %v", err)
		return
	}
	s.redis.Expire(ctx, searchZSetKey, 30*time.Minute)
}

func (s *MovieService) GetMovieDetails(id int) (*models.Movie, error) {
	cacheKey := fmt.Sprintf("movie:%d", id)

	var movie models.Movie
	if cached, err := s.cache.Get(cacheKey); err == nil && json.Unmarshal(cached, &movie) == nil {
		return &movie, nil
	}

	body, err := s.makeRequest(fmt.Sprintf("%s/movie/%d", apiBaseURL, id))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &movie); err != nil {
		return nil, err
	}

	if data, err := json.Marshal(movie); err == nil {
		s.cache.Set(cacheKey, data, time.Hour)
	}

	return &movie, nil
}

func (s *MovieService) GetTrendingMovies() ([]models.Movie, error) {
	results, err := s.redis.ZRevRangeWithScores(context.Background(), searchZSetKey, 0, 4).Result()
	if err != nil {
		return nil, err
	}

	trendingMovies := make([]models.Movie, 0, len(results))
	for _, result := range results {
		idStr, ok := result.Member.(string)
		if !ok {
			continue
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		movie, err := s.GetMovieDetails(id)
		if err != nil {
			continue
		}

		movie.SearchCount = int(result.Score)
		trendingMovies = append(trendingMovies, *movie)
	}

	return trendingMovies, nil
}

func (s *MovieService) SearchMovies(query string, page int) ([]models.Movie, error) {
	movies, err := s.searchMovies(query, page)
	if err != nil {
		return nil, err
	}

	// Count every search (cached or not) towards trending.
	if len(movies) > 0 {
		go s.logSearch(movies[0])
	}
	return movies, nil
}

func (s *MovieService) searchMovies(query string, page int) ([]models.Movie, error) {
	cacheKey := fmt.Sprintf("search:%s:%d", query, page)
	if cached, err := s.cache.Get(cacheKey); err == nil {
		var movies []models.Movie
		if json.Unmarshal(cached, &movies) == nil {
			return movies, nil
		}
	}

	movies, err := s.fetchMovieList(fmt.Sprintf(
		"%s/search/movie?query=%s&include_adult=false&language=en-US&page=%d",
		apiBaseURL, url.QueryEscape(query), page,
	))
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(movies); err == nil {
		s.cache.Set(cacheKey, data, time.Hour)
	}
	return movies, nil
}

func (s *MovieService) GetDiscoverMovies(page int) ([]models.Movie, error) {
	return s.fetchMovieList(fmt.Sprintf(
		"%s/discover/movie?include_adult=false&include_video=false&language=en-US&page=%d&sort_by=popularity.desc",
		apiBaseURL, page,
	))
}

func (s *MovieService) GetTopPopularMovies() ([]models.Movie, error) {
	movies, err := s.fetchMovieList(fmt.Sprintf("%s/movie/popular?language=en-US&page=1", apiBaseURL))
	if err != nil {
		return nil, err
	}

	if len(movies) > 3 {
		return movies[:3], nil
	}
	return movies, nil
}
