package parser

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"tweakio/internal/api"
	"tweakio/internal/cache"
	"tweakio/internal/logger"
)

var Regexes *RegexPatterns

type RegexPatterns struct {
	Size          *regexp.Regexp
	Season        *regexp.Regexp
	SeasonRange   *regexp.Regexp
	SingleEpisode *regexp.Regexp
	EpisodeRange  *regexp.Regexp
	Episode       *regexp.Regexp
	Info          *regexp.Regexp
}

type TorrentioResult struct {
	Title    string
	Link     string
	Size     float64
	InfoHash string
	Peers    int
	Category int
	Source   string
}

func CompileRegex() error {
	regexes := &RegexPatterns{}
	var err error

	if regexes.Season, err = regexp.Compile("(?i)\\b(?:season\\s*|s)(\\d{1,2})(?:\\s+(?:complete|full|pack|multi)|\\.[a-z]+)?\\b"); err != nil {
		return err
	}
	if regexes.SeasonRange, err = regexp.Compile("(?i)\\b(?:s(?:eason)?\\s*\\d{1,2}(?:-s?(?:eason)?\\s*\\d{1,2})|s\\d{1,2}e\\d{1,2}-\\d{1,2})\\b"); err != nil {
		return err
	}
	if regexes.SingleEpisode, err = regexp.Compile("(?i)\\bs\\d{1,2}e\\d{1,2}\\b"); err != nil {
		return err
	}
	if regexes.EpisodeRange, err = regexp.Compile("\\bS?\\d{1,2}E(\\d{1,3})-(\\d{1,3})\\b"); err != nil {
		return err
	}
	if regexes.Episode, err = regexp.Compile("\\b[eE]\\d{2,3}\\b"); err != nil {
		return err
	}
	if regexes.Info, err = regexp.Compile("👤\\s*(\\d+)\\s*💾\\s*([\\d.]+)\\s*(GB|MB)\\s*⚙️\\s*(.+)"); err != nil {
		return err
	}

	Regexes = regexes
	return nil
}

func ParseResult(result any, mediaType, imdbID string, httpClient *api.APIClient, episodeCache *cache.EpisodeCache) (*TorrentioResult, error) {
	parsedResult, ok := result.(map[string]any)
	if !ok {
		return nil, errors.New("invalid result format")
	}

	title, ok := parsedResult["title"].(string)
	if !ok {
		return nil, errors.New("missing title from result")
	}

	cleanTitle := GetCleanTitle(title)

	torrentioResult := &TorrentioResult{
		Title:    cleanTitle,
		InfoHash: parsedResult["infoHash"].(string),
		Category: 5000,
	}

	parseInfo(title, torrentioResult)

	if mediaType != "tvsearch" {
		torrentioResult.Category = 2000
		return torrentioResult, nil
	}

	if start, end, found := GetSeasonRange(cleanTitle); found {
		episodes := GetOrFetchEpisodes(imdbID, start, end, httpClient, episodeCache)
		torrentioResult.Size *= float64(episodes)
		return torrentioResult, nil
	}
	if season, found := GetSeasonNumber(cleanTitle); found {
		episodes := GetOrFetchEpisodes(imdbID, season, season, httpClient, episodeCache)
		torrentioResult.Size *= float64(episodes)
		return torrentioResult, nil
	}
	if start, end, found := GetEpisodeRange(cleanTitle); found {
		episodes := end - start + 1
		torrentioResult.Size *= float64(episodes)
		return torrentioResult, nil
	}
	return torrentioResult, nil
}

func GetOrFetchEpisodes(imdbID string, start, end int, httpClient *api.APIClient, episodeCache *cache.EpisodeCache) int {
	episodes := 0
	for i := start; i <= end; i++ {
		if episodeCache == nil {
			episodes += 10
			continue
		}
		
		if seasonEpisodes, exists := episodeCache.Get(imdbID, i); exists {
			episodes += seasonEpisodes
			continue
		} 
		
		seasonEpisodes, err := httpClient.FetchEpisodesFromTMDB(imdbID, i)
		if err != nil {
			logger.Error("Error fetching espiodes from TMDB", err)
		}
		episodeCache.Set(imdbID, i, seasonEpisodes)
		episodes += seasonEpisodes
	}
	return episodes
}

func parseInfo(title string, torrentioResult *TorrentioResult) {
	peers := 0
	size := float64(0)
	source := "Unknown"

	if match := Regexes.Info.FindStringSubmatch(title); len(match) == 5 {
		peers, _ = strconv.Atoi(match[1])
		size, _ = strconv.ParseFloat(match[2], 64)
		sizeUnit := match[3]
		source = match[4]

		if sizeUnit == "MB" {
			size /= 1024
		}
	}

	torrentioResult.Peers = peers
	torrentioResult.Size = size
	torrentioResult.Source = source
}

func GetCleanTitle(title string) string {
	return strings.Split(title, "\n")[0]
}

func GetSeasonRange(title string) (int, int, bool) {
	if match := Regexes.SeasonRange.FindStringSubmatch(title); len(match) > 2 {
		start, _ := strconv.Atoi(match[1])
		end, _ := strconv.Atoi(match[2])
		return start, end, true
	}
	return 0, 0, false
}

func GetSeasonNumber(title string) (int, bool) {
	if match := Regexes.Season.FindStringSubmatch(title); len(match) > 1 {
		season, _ := strconv.Atoi(match[1])
		return season, true
	}
	return 0, false
}

func GetEpisodeRange(title string) (start, end int, found bool) {
	if match := Regexes.EpisodeRange.FindStringSubmatch(title); len(match) == 3 {
		start, _ = strconv.Atoi(match[1])
		end, _ = strconv.Atoi(match[2])
		return start, end, true
	}
	return 0, 0, false
}
