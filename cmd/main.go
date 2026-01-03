package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"tweakio/config"
	"tweakio/internal/api"
	"tweakio/internal/cache"
	"tweakio/internal/logger"
	"tweakio/internal/parser"
	"tweakio/internal/torznab"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("TWEAKIO", "Failed to load config: %v", err)
		os.Exit(1)
	}

	if logger.DebugEnabled {
		tmpConfig := *cfg
		tmpConfig.TMDB.APIKey = "<REDACTED>"
		logger.Debug("TWEAKIO", "Config loaded:\n%+v", tmpConfig)
	}

	if err := parser.CompileRegex(); err != nil {
		logger.Error("TWEAKIO", "Failed to compile regex: %v", err)
		os.Exit(1)
	}

	httpClient := api.NewAPIClient(cfg.TorrentioURL, cfg.ProxyURL, cfg.TMDB.APIKey)

	var episodeCache *cache.EpisodeCache
	if cfg.TMDB.APIKey != "" {
		episodeCache = cache.CreateEpisodeCache(cfg.TMDB.CacheSize)
	}

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		handleProwlarrRequest(w, r, httpClient, episodeCache)
	})

	logger.Info("TWEAKIO", "Running on port 3185")
	if err := http.ListenAndServe(":3185", nil); err != nil {
		logger.Error("TWEAKIO", "Failed to start: %v", err)

	}
}

func handleProwlarrRequest(w http.ResponseWriter, r *http.Request, httpClient *api.APIClient, episodeCache *cache.EpisodeCache) {
	query := r.URL.Query()
	t := query.Get("t")
	imdbID := query.Get("imdbid")
	season, _ := strconv.Atoi(query.Get("season"))
	episode, _ := strconv.Atoi(query.Get("ep"))

	logger.Info("TWEAKIO", "Received request: type=%s, imdbID=%s, season=%d, episode=%d", t, imdbID, season, episode)

	if t == "rss" {
		sendResponse(w, torznab.RssResponse())
		return
	}

	if t == "caps" {
		sendResponse(w, torznab.CapsResponse())
		return
	}

	if imdbID == "" {
		fakeResults, err := torznab.GenerateFakeResults()
		if err != nil {
			logger.Error("TWEAKIO", "Error generating placeholder results: %v", err)
			sendError(w)
			return
		}
		sendResponse(w, fakeResults)
		return
	}

	if !strings.HasPrefix(imdbID, "tt") {
		imdbID = "tt" + imdbID
	}

	mediaType := "movie"
	if t == "tvsearch" {
		mediaType = "series"
	}

	results, err := httpClient.FetchFromTorrentio(mediaType, imdbID, season, episode)
	if err != nil {
		logger.Error("TORRENTIO", "Error fetching results: %v", err)
		sendError(w)
		return
	}

	parseStart := time.Now()

	var parsedResults []parser.TorrentioResult
	for _, result := range results {
		torrentioResult, err := parser.ParseResult(result, t, imdbID, httpClient, episodeCache)
		if err != nil {
			logger.Error("TORRENTIO", "Error parsing result: %v", err)
		} else {
			parsedResults = append(parsedResults, *torrentioResult)
		}
	}

	torznabResponse, err := torznab.ConvertToTorznab(parsedResults, "http://tweakio:3185/api")
	if err != nil {
		logger.Error("TWEAKIO", "Error convertng results to Torznab: %v", err)
		sendError(w)
		return
	}

	parseDuration := time.Since(parseStart).Seconds()
	logger.Info("TWEAKIO", "Processed %d results in %f sec", len(parsedResults), parseDuration)

	sendResponse(w, torznabResponse)
}

func sendResponse(w http.ResponseWriter, response string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(response))
	if err != nil {
		logger.Error("TWEAKIO", "Error writing response: %v", err)
	}
}

func sendError(w http.ResponseWriter) {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}
