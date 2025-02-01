package parser

import (
	"regexp"
	"testing"
	"tweakio/config"
)

func TestCompiledRegexPatterns(t *testing.T) {
	cfg, err := config.LoadConfig("../../config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if err := CompileRegex(cfg); err != nil {
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
			{"S01-S03", true}, {"Season 1-3", true}, {"S02E01-09", true},
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
