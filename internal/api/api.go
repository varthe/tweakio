package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"tweakio/internal/logger"
)

type APIClient struct {
	TorrentioURL *url.URL
	TMDBBaseURL  string
	TMDBAPIKey   string
	Client       *http.Client
}

type userAgentTransport struct {
	base http.RoundTripper
}

func NewAPIClient(torrentioURL, proxyURL *url.URL, tmdbAPIKey string) *APIClient {
	baseTransport := http.DefaultTransport
	if proxyURL != nil {
		baseTransport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	return &APIClient{
		TorrentioURL: torrentioURL,
		TMDBBaseURL:  "https://api.themoviedb.org/3",
		TMDBAPIKey:   tmdbAPIKey,
		Client: &http.Client{
			Transport: &userAgentTransport{
				base: baseTransport,
			},
		},
	}
}

func (u *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	return u.base.RoundTrip(req)
}

func fetchJSON(httpClient *http.Client, url string, apiKey string, result any) error {
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

func (c *APIClient) FetchFromTorrentio(mediaType, imdbID string, season, episode int) ([]any, error) {
	url := *c.TorrentioURL

	url.Path = path.Join(url.Path, "stream", "mediaType", imdbID)

	if mediaType == "series" {
		if season == 0 {
			season = 1
		}
		if episode == 0 {
			episode = 1
		}
		url.Path += fmt.Sprintf(":%d:%d", season, episode)
	}

	url.Path += ".json"

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
