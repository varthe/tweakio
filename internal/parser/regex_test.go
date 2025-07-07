package parser

import (
	"regexp"
	"testing"
)

func TestCompiledRegexPatterns(t *testing.T) {
	if err := CompileRegex(); err != nil {
		t.Fatalf("Failed to compile regex: %v", err)
	}

	tests := map[string][]struct {
		input    string
		expected bool
	}{
		"season": {
			{"Season 3", true}, {"S02", true}, {"S02 COMPLETE", true},
			{"Season 2 Pack", true}, {"S02.MULTi", true}, {"S02E01", false},
			{"CHUNGITE.2023.S01.COMPLETE.1080p.NF.WEB-DL.DD5.1.Atmos.H.264", true},
		},
		"season_range": {
			{"S01-S03", true}, {"Season 1-3", true}, {"S02E01-09", false},
			{"S03-05", true}, {"S1-3", true}, {"S1-E3", false},
			{"ONE.CHUNGUS.2023.S01-7.2160p.NF.WEB-DL.DDP5.1.Atmos.H.265", true},
		},
		"single_episode": {
			{"S02E05", true}, {"s01e01", true}, {"S1E1", true},
			{"RETURN.OF.THE.CHUNGUS.2023.S01E01.720p.NF.WEB-DL.DDP5", true},
		},
		"episode_range": {
			{"Chungus.S01E01-09.2160p.UHD.BluRay.DDP.5.1.ITA.ATMOS.ENG.DV.HDR.x265-G66", true},
		},
		"episode": {
			{"E05", true}, {"e10", true}, {"E105", true},
		},
		"info": {
			{"👤 5 💾 20 GB ⚙️ CornHub", true},
			{"💾 15 GB ⚙️ CornHub", false},
			{"👤 999 💾 1.2 MB ⚙️ CornHub", true},
			{"⚙️ CornHub", false},
			{"", false},
			{"👤 0 💾 0 GB ⚙️ Unknown", true},
			{"👤 1500 💾 2.33 GB ⚙️ CornHub", true},
		},
	}

	failed := false
	for pattern, cases := range tests {
		t.Run(pattern, func(t *testing.T) {
			var regex *regexp.Regexp
			switch pattern {
			case "season":
				regex = Regexes.Season
			case "season_range":
				regex = Regexes.SeasonRange
			case "single_episode":
				regex = Regexes.SingleEpisode
			case "episode_range":
				regex = Regexes.EpisodeRange
			case "episode":
				regex = Regexes.Episode
			case "info":
				regex = Regexes.Info
			default:
				t.Fatalf("Unknown regex pattern: %s", pattern)
			}

			for _, test := range cases {
				if match := regex.MatchString(test.input); match != test.expected {
					t.Errorf("Input '%s': expected %v, got %v",
						test.input, test.expected, match)
					failed = true
				}
			}
		})
	}

	if !failed {
		t.Log("✅ All regex tests passed!")
	}
}

func TestRegexCapturing(t *testing.T) {
	if err := CompileRegex(); err != nil {
		t.Fatalf("Failed to compile regex: %v", err)
	}

	tests := []struct {
		input    string
		expected []string
	}{
		{"S02E05", []string{"S02E05", "02", "05"}},
		{"s01e01", []string{"s01e01", "01", "01"}},
		{"S1E1", []string{"S1E1", "1", "1"}},
		{"RETURN.OF.THE.CHUNGUS.2023.S01E01.720p", []string{"S01E01", "01", "01"}},
		{"Family Guy S21E01 1080p WEB", []string{"S21E01", "21", "01"}},
		{"Not a match", nil},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			matches := Regexes.SingleEpisode.FindStringSubmatch(test.input)

			if test.expected == nil {
				if matches != nil {
					t.Errorf("Input '%s': expected no match, but got %v", test.input, matches)
				}
				return
			}

			if matches == nil {
				t.Errorf("Input '%s': expected match %v, but got no match",
					test.input, test.expected)
				return
			}

			if len(matches) != len(test.expected) {
				t.Errorf("Input '%s': expected %d captures, got %d",
					test.input, len(test.expected), len(matches))
				return
			}

			for i, expectedMatch := range test.expected {
				if matches[i] != expectedMatch {
					t.Errorf("Input '%s': capture group %d expected '%s', got '%s'",
						test.input, i, expectedMatch, matches[i])
				}
			}
		})
	}

	seasonEpisodeTests := []struct {
		input          string
		expectedSeason int
		expectedEp     int
		expectedFound  bool
	}{
		{"S02E05", 2, 5, true},
		{"s01e01", 1, 1, true},
		{"S1E1", 1, 1, true},
		{"Family Guy S21E01", 21, 1, true},
		{"No match here", 0, 0, false},
	}

	for _, test := range seasonEpisodeTests {
		t.Run("GetSeasonEpisode: "+test.input, func(t *testing.T) {
			season, episode, found := GetSeasonEpisode(test.input)
			if found != test.expectedFound {
				t.Errorf("Input '%s': expected found=%v, got %v",
					test.input, test.expectedFound, found)
			}

			if found {
				if season != test.expectedSeason {
					t.Errorf("Input '%s': expected season=%d, got %d",
						test.input, test.expectedSeason, season)
				}
				if episode != test.expectedEp {
					t.Errorf("Input '%s': expected episode=%d, got %d",
						test.input, test.expectedEp, episode)
				}
			}
		})
	}
}
