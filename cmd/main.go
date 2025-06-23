package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"tweakio/config"
	"tweakio/internal/api"
	"tweakio/internal/cache"
	"tweakio/internal/logger"
	"tweakio/internal/parser"
	"tweakio/internal/torznab"
)

func main() {
	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		logger.Fatal("Failed to load config", err)
	}

	if err := parser.CompileRegex(); err != nil {
		logger.Fatal("Failed to compile regex", err)
	}

	httpClient := api.NewAPIClient(cfg.Torrentio.BaseURL, cfg.Torrentio.Options, cfg.TMDB.APIKey)
	episodeCache := cache.CreateEpisodeCache(cfg.TMDB.CacheSize)

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		handleProwlarrRequest(w, r, httpClient, episodeCache)
	})

	logger.Info("Tweakio is running on port 3185")
	if err := http.ListenAndServe(":3185", nil); err != nil {
		logger.Fatal("Failed to start", err)
	}
}

func handleProwlarrRequest(w http.ResponseWriter, r *http.Request, httpClient *api.APIClient, episodeCache *cache.EpisodeCache) {
	query := r.URL.Query()
	t := query.Get("t")
	imdbID := "tt" + query.Get("imdbid")
	season, _ := strconv.Atoi(query.Get("season"))
	episode, _ := strconv.Atoi(query.Get("ep"))

	msg := fmt.Sprintf("Incoming request: type=%s, imdbID=%s, season=%d, episode=%d", t, imdbID, season, episode)
	logger.Info(msg)

	if t == "caps" || (t == "search" && imdbID == "tt") {
		capsResponse, err := torznab.GenerateCapsResponse(t)
		if err != nil {
			sendError(w, "Error generating caps response", err)
			return
		}
		sendResponse(w, capsResponse)
		return
	}

	if t == "rss" {
		emptyRSS := `<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel></channel></rss>`
		sendResponse(w, emptyRSS)
		return
	}

	if imdbID == "" {
		sendError(w, "Missing required parameter", errors.New("imdbID"))
		return
	}

	mediaType := "movie"
	if t == "tvsearch" {
		mediaType = "series"
	}

	results, err := httpClient.FetchFromTorrentio(mediaType, imdbID, season, episode)
	if err != nil {
		sendError(w, "Error fetching from Torrentio", err)
		return
	}

	var parsedResults []parser.TorrentioResult
	for _, result := range results {
		torrentioResult, err := parser.ParseResult(result, t, imdbID, httpClient, episodeCache)
		if err != nil {
			logger.Error("Error parsing result from Torrentio", err)
		} else {
			parsedResults = append(parsedResults, *torrentioResult)
		}
	}

	torznabResponse, err := torznab.ConvertToTorznab(parsedResults, "http://tweakio:3185/api")
	if err != nil {
		sendError(w, "Error generating Torznab response", err)
		return
	}

	sendResponse(w, torznabResponse)
}

func sendResponse(w http.ResponseWriter, response string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(response))
	if err != nil {
		logger.Error("Error writing response", err)
	}
}

func sendError(w http.ResponseWriter, msg string, err error) {
	logger.Error(msg, err)
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}