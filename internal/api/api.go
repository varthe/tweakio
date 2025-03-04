package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type APIClient struct {
	TorrentioBaseURL string
	TorrentioOptions string
	TMDBBaseURL string
	TMDBAPIKey string
	Client *http.Client
}

func NewAPIClient(torrentioBaseURL, torrentioOptions, tmdbAPIKey string) *APIClient {
	return &APIClient{
		TorrentioBaseURL: torrentioBaseURL,
		TorrentioOptions: torrentioOptions,
		TMDBBaseURL: "https://api.themoviedb.org/3",
		TMDBAPIKey: tmdbAPIKey,
		Client: &http.Client{},
	}
}

func fetchJSON(httpClient *http.Client, url string, result interface{}) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch from URL %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned an error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	return nil
}

func (c *APIClient) FetchFromTorrentio(mediaType, imdbID string, season, episode int) ([]interface{}, error) {
	url := strings.Builder{}
	fmt.Fprintf(&url, "%s%s/stream/%s/%s", c.TorrentioBaseURL, c.TorrentioOptions, mediaType, imdbID)

	if mediaType == "series" && season > 0 {
		if episode == 0 {
			episode = 1
		}
		fmt.Fprintf(&url, ":%d:%d", season, episode)
	}

	fmt.Fprintf(&url, ".json")

	log.Printf("Fetching from Torrentio: %s\n", url.String())

	var result map[string]interface{}
	if err := fetchJSON(c.Client, url.String(), &result); err != nil {
		return nil, fmt.Errorf("failed to fetch from Torrentio: %w", err)
	}

	streams, ok := result["streams"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse streams from Torrentio")
	}

	return streams, nil
}

func fetchIdFromTMDB(c *APIClient, imdbID string) (string, error) {
	url := fmt.Sprintf("%s/find/%s?api_key=%s&external_source=imdb_id", c.TMDBBaseURL, imdbID, c.TMDBAPIKey)
	var result map[string]interface{}
	if err := fetchJSON(c.Client, url, &result); err != nil {
		return "", fmt.Errorf("failed to fetch ID from TMDB: %w", err)
	}

	tvResults, ok := result["tv_results"].([]interface{})
	if !ok || len(tvResults) == 0 {
		return "", fmt.Errorf("no TMDB ID found for IMDB ID %s", imdbID)
	}
	
	tvData, ok := tvResults[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid TMDB data format")
	}

	tmdbID, ok := tvData["id"].(float64)
	if !ok {
		return "", fmt.Errorf("invalid TMDB ID format")
	}

	return strconv.FormatFloat(tmdbID, 'f', 0, 64), nil
}

func (c *APIClient) FetchEpisodesFromTMDB(imdbID string, seasonNumber int) (int) {
	if c.TMDBAPIKey == "" {
		return 10
	}

	tmdbID, err := fetchIdFromTMDB(c, imdbID)
	if err != nil {
		log.Printf("Error fetching TMDB ID for IMDB ID %s: %v", imdbID, err)
		return 10
	}

	url := fmt.Sprintf("%s/tv/%s/season/%d?api_key=%s", c.TMDBBaseURL, tmdbID, seasonNumber, c.TMDBAPIKey)

	var result map[string]interface{}
	if err := fetchJSON(c.Client, url, &result); err != nil {
		log.Printf("Error fetching episode count from TMDB ID %s: %v", tmdbID, err)
		return 10
	}

	episodes, ok := result["episodes"].([]interface{})
	if !ok {
		log.Printf("Invalid episode format for TMDB ID %s", tmdbID)
		return 10
	}

	return len(episodes)
}