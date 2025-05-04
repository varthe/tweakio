package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"tweakio/config"
	"tweakio/internal/api"
	"tweakio/internal/cache"
	"tweakio/internal/parser"
	"tweakio/internal/torznab"
)

func main() {
	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := parser.CompileRegex(); err != nil {
		log.Fatalf("Failed to compile regex: %v", err)
	}

	httpClient := api.NewAPIClient(cfg.Torrentio.BaseURL, cfg.Torrentio.Options, cfg.TMDB.APIKey)
	episodeCache := cache.CreateEpisodeCache(cfg.TMDB.CacheSize)

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		handleProwlarrRequest(w, r, httpClient, episodeCache)
	})

	log.Println("Tweakio is running on :3185")
	log.Fatal(http.ListenAndServe(":3185", nil))
}

func handleProwlarrRequest(w http.ResponseWriter, r *http.Request, httpClient *api.APIClient, episodeCache *cache.EpisodeCache) {
	query := r.URL.Query()
	t := query.Get("t")
	imdbID := "tt" + query.Get("imdbid")
	season, _ := strconv.Atoi(query.Get("season"))
	episode, _ := strconv.Atoi(query.Get("ep"))

	log.Printf("Incoming request: t=%s, imdbID=%s, season=%d, episode=%d\n", t, imdbID, season, episode)

	if t == "caps" || (t == "search" && imdbID == "tt") {
		capsResponse, err := torznab.GenerateCapsResponse(t)
		if err != nil {
			msg := fmt.Sprintf("Error getting Torznab response: %v", err)
			log.Println(msg)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		writeResponse(w, capsResponse)
		return
	}

	if t == "rss" {
		emptyRSS := `<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel></channel></rss>`
		writeResponse(w, emptyRSS)
		return
	}

	if imdbID == "" {
		msg := "Missing required paramenter: imdbid"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	mediaType := "movie"
	if t == "tvsearch" {
		mediaType = "series"
	}

	results, err := httpClient.FetchFromTorrentio(mediaType, imdbID, season, episode)
	if err != nil {
		msg := fmt.Sprintf("Error fetching from Torrentio: %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	var parsedResults []parser.TorrentioResult
	for _, result := range results {
		torrentioResult := parser.ParseResult(result, t, imdbID, httpClient, episodeCache)
		if torrentioResult != nil {
			parsedResults = append(parsedResults, *torrentioResult)
		}
	}

	torznabResponse, err := torznab.ConvertToTorznab(parsedResults, "http://tweakio:3185/api")
	if err != nil {
		msg := fmt.Sprintf("Error generating Torznab response: %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	writeResponse(w, torznabResponse)
}

func writeResponse(w http.ResponseWriter, response string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(response))
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
