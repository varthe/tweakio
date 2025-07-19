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
	if regexes.SeasonRange, err = regexp.Compile("(?i)\\b(?:s(?:eason)?\\s*(\\d{1,2})(?:-s?(?:eason)?\\s*(\\d{1,2}))|s(\\d{1,2})-s(\\d{1,2}))\\b"); err != nil {
		return err
	}
	if regexes.SingleEpisode, err = regexp.Compile("(?i)\\bs(\\d{1,2})e(\\d{1,2})\\b"); err != nil {
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

	logger.Debug("PARSER", "Processing result: title=%s, infoHash=%v", title, parsedResult["infoHash"])

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

	if season, episode, found := GetSeasonEpisode(cleanTitle); found {
		logger.Debug("PARSER", "Found season and episode in title '%s': S%dE%d", cleanTitle, season, episode)
		episodes := GetOrFetchEpisodes(imdbID, season, season, httpClient, episodeCache)
		torrentioResult.Size *= float64(episodes)
		return torrentioResult, nil
	}
	if start, end, found := GetSeasonRange(cleanTitle); found {
		logger.Debug("PARSER", "Found season range in title '%s': start=%d, end=%d", cleanTitle, start, end)
		episodes := GetOrFetchEpisodes(imdbID, start, end, httpClient, episodeCache)
		torrentioResult.Size *= float64(episodes)
		return torrentioResult, nil
	}
	if season, found := GetSeasonNumber(cleanTitle); found {
		logger.Debug("PARSER", "Found season number in title '%s': season=%d", cleanTitle, season)
		episodes := GetOrFetchEpisodes(imdbID, season, season, httpClient, episodeCache)
		torrentioResult.Size *= float64(episodes)
		return torrentioResult, nil
	}
	if start, end, found := GetEpisodeRange(cleanTitle); found {
		logger.Debug("PARSER", "Found episode range in title '%s': start=%d, end=%d", cleanTitle, start, end)
		episodes := end - start + 1
		torrentioResult.Size *= float64(episodes)
		return torrentioResult, nil
	}

	logger.Debug("PARSER", "No season or episode information found in title '%s'", cleanTitle)
	return torrentioResult, nil
}

func GetOrFetchEpisodes(imdbID string, start, end int, httpClient *api.APIClient, episodeCache *cache.EpisodeCache) int {
	if episodeCache == nil {
		return 10 * (end - start + 1)
	}

	episodes := 0
	missing := false

	for i := start; i <= end; i++ {
		if seasonEpisodes, exists := episodeCache.Get(imdbID, i); exists {
			episodes += seasonEpisodes
		} else {
			missing = true
		}
	}

	if !missing {
		return episodes
	}

	tvDetails, err := httpClient.FetchTVShowDetails(imdbID)
	if err != nil {
		logger.Error("TMDB", "Error fetching TV show details from TMDB: %v", err)
		return 10 * (end - start + 1)
	}

	seasons, ok := tvDetails["seasons"].([]any)
	if !ok {
		logger.Error("TMDB", "Failed to get seasons from response")
		return 10 * (end - start + 1)
	}

	for _, season := range seasons {
		seasonData, ok := season.(map[string]any)
		if !ok {
			logger.Warn("TMDB", "Could not parse season data from response for IMDB ID %s", imdbID)
			continue
		}

		seasonNumRaw, ok := seasonData["season_number"].(float64)
		if !ok {
			logger.Warn("TMDB", "Could not find season number in season data for IMDB ID %s", imdbID)
			continue
		}
		seasonNum := int(seasonNumRaw)

		// Rechecking the cache allows us to store all seasons from the API response.
		// e.g. if start=1 and end=3 but the show has 5 seasons,
		// we cache them all now to prevent redundant fetches later.
		if _, exists := episodeCache.Get(imdbID, seasonNum); exists {
			continue
		}

		episodeCountRaw, ok := seasonData["episode_count"].(float64)
		if !ok {
			logger.Warn("TMDB", "Could not find episode count for season %d in season data for IMDB ID %s", seasonNum, imdbID)
			continue
		}
		episodeCount := int(episodeCountRaw)

		episodeCache.Set(imdbID, seasonNum, episodeCount)
		episodes += episodeCount
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
	if match := Regexes.SeasonRange.FindStringSubmatch(title); len(match) > 0 {
		if match[1] != "" && match[2] != "" {
			start, _ := strconv.Atoi(match[1])
			end, _ := strconv.Atoi(match[2])
			return start, end, true
		}

		if match[3] != "" && match[4] != "" {
			start, _ := strconv.Atoi(match[3])
			end, _ := strconv.Atoi(match[4])
			return start, end, true
		}
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

func GetSeasonEpisode(title string) (season, episode int, found bool) {
	if match := Regexes.SingleEpisode.FindStringSubmatch(title); len(match) == 3 {
		season, _ = strconv.Atoi(match[1])
		episode, _ = strconv.Atoi(match[2])
		return season, episode, true
	}
	return 0, 0, false
}
