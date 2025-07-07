package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"tweakio/internal/logger"
)

type APIClient struct {
	TorrentioBaseURL string
	TorrentioOptions string
	TMDBBaseURL      string
	TMDBAPIKey       string
	Client           *http.Client
}

type userAgentTransport struct {
	transport http.RoundTripper
}

func NewAPIClient(torrentioBaseURL, torrentioOptions, tmdbAPIKey string) *APIClient {
	return &APIClient{
		TorrentioBaseURL: torrentioBaseURL,
		TorrentioOptions: torrentioOptions,
		TMDBBaseURL:      "https://api.themoviedb.org/3",
		TMDBAPIKey:       tmdbAPIKey,
		Client: &http.Client{
			Transport: &userAgentTransport{transport: http.DefaultTransport},
		},
	}
}

func (u *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	return u.transport.RoundTrip(req)
}

func fetchJSON(httpClient *http.Client, url string, apiKey string, result interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for URL %s: %w", url, err)
	}

	if apiKey != "" && strings.HasPrefix(apiKey, "eyJ") {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := httpClient.Do(req)
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

	logger.Info("TORRENTIO", "Fetching results from: %s", url.String())

	var result map[string]any
	if err := fetchJSON(c.Client, url.String(), "", &result); err != nil {
		return nil, err
	}

	streams, ok := result["streams"].([]any)
	if !ok {
		return nil, errors.New("invalid result structure")
	}

	return streams, nil
}

func fetchIdFromTMDB(c *APIClient, imdbID string) (string, error) {
	baseUrl := fmt.Sprintf("%s/find/%s?external_source=imdb_id", c.TMDBBaseURL, imdbID)
	if !strings.HasPrefix(c.TMDBAPIKey, "eyJ") {
		baseUrl = fmt.Sprintf("%s&api_key=%s", baseUrl, c.TMDBAPIKey)
	}

	var result map[string]any
	if err := fetchJSON(c.Client, baseUrl, c.TMDBAPIKey, &result); err != nil {
		return "", fmt.Errorf("failed to fetch TMDB ID: %w", err)
	}

	tvResults, ok := result["tv_results"].([]any)
	if !ok || len(tvResults) == 0 {
		return "", fmt.Errorf("no TMDB ID found for IMDB ID %s", imdbID)
	}

	tvData, ok := tvResults[0].(map[string]any)
	if !ok {
		return "", fmt.Errorf("invalid TMDB data format")
	}

	tmdbID, ok := tvData["id"].(float64)
	if !ok {
		return "", fmt.Errorf("invalid TMDB ID format")
	}

	return strconv.FormatFloat(tmdbID, 'f', 0, 64), nil
}

func (c *APIClient) FetchTVShowDetails(imdbID string) (map[string]any, error) {
	logger.Info("TMDB", "Fetching TV show details")

	tmdbID, err := fetchIdFromTMDB(c, imdbID)
	if err != nil {
		return nil, err
	}

	baseUrl := fmt.Sprintf("%s/tv/%s", c.TMDBBaseURL, tmdbID)
	if !strings.HasPrefix(c.TMDBAPIKey, "eyJ") {
		baseUrl = fmt.Sprintf("%s?api_key=%s", baseUrl, c.TMDBAPIKey)
	}

	var result map[string]any
	if err := fetchJSON(c.Client, baseUrl, c.TMDBAPIKey, &result); err != nil {
		return nil, fmt.Errorf("failed to fetch TV show details: %w", err)
	}

	return result, nil
}

func (c *APIClient) FetchEpisodesFromTMDB(imdbID string, seasonNumber int) (int, error) {
	logger.Info("TMDB", "Fetching episode count for season %d", seasonNumber)

	result, err := c.FetchTVShowDetails(imdbID)
	if err != nil {
		return 10, err
	}

	seasons, ok := result["seasons"].([]any)
	if !ok {
		return 10, fmt.Errorf("failed to get seasons from response")
	}

	for _, season := range seasons {
		seasonData, ok := season.(map[string]any)
		if !ok {
			continue
		}

		seasonNum, ok := seasonData["season_number"].(float64)
		if !ok || int(seasonNum) != seasonNumber {
			continue
		}

		episodeCount, ok := seasonData["episode_count"].(float64)
		if !ok {
			return 10, fmt.Errorf("failed to get episode count for season %d", seasonNumber)
		}

		return int(episodeCount), nil
	}

	return 10, fmt.Errorf("season %d not found", seasonNumber)
}
