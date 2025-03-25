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
	metadataKey   = "movie_metadata"
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

func (s *MovieService) logSearch(movie models.Movie) error {
	ctx := context.Background()

	member := fmt.Sprintf("%d", movie.ID)

	err := s.redis.ZIncrBy(ctx, searchZSetKey, 1, member).Err()
	if err != nil {
		log.Printf("Error incrementing search count: %v", err)
		return err
	}

	movieData := map[string]interface{}{
		"title":       movie.Title,
		"poster_path": movie.PosterPath,
	}
	err = s.redis.HSet(ctx, metadataKey, member, movieData).Err()
	if err != nil {
		log.Printf("Error saving movie metadata: %v", err)
	}

	s.redis.Expire(ctx, searchZSetKey, 30*time.Minute)
	s.redis.Expire(ctx, metadataKey, 30*time.Minute)

	return nil
}

func (s *MovieService) GetTrendingMovies() ([]models.TrendingMovie, error) {
	ctx := context.Background()

	results, err := s.redis.ZRevRangeWithScores(ctx, searchZSetKey, 0, 4).Result()
	if err != nil {
		return nil, err
	}

	var trendingMovies []models.TrendingMovie

	for _, result := range results {
		member := result.Member.(string)

		movieData, err := s.redis.HGet(ctx, metadataKey, member).Result()
		if err != nil {
			continue
		}

		var data map[string]string
		if err := json.Unmarshal([]byte(movieData), &data); err != nil {
			continue
		}
		id, _ := strconv.Atoi(member)
		movie := models.TrendingMovie{
			ID:          id,
			Title:       data["title"],
			PosterPath:  data["poster_path"],
			SearchCount: int(result.Score),
		}
		trendingMovies = append(trendingMovies, movie)
	}

	return trendingMovies, nil
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

	if len(response.Results) > 0 {
		go func() {
			s.logSearch(response.Results[0])
		}()
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
