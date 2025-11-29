package app

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Kondou727/maimai-recommend-intl/internal/config"
	songdatadb "github.com/Kondou727/maimai-recommend-intl/internal/database/songdata"
)

var NoUpdate = errors.New("no update")
var songdataJson = "resources/divingfish_songdata.json"
var chartstatsJson = "resources/divingfish_chartstats.json"

// gets chinese song data
func pullJsonCN() ([]divingFishSong, chartStatsResponse, error) {
	etags, err := loadEtags()
	if err != nil {
		return nil, chartStatsResponse{}, err
	}

	var songdata []divingFishSong
	songdataUrl := DIVING_FISH_API + "music_data"
	etag, body, err := sendGetRequestWithEtag(songdataUrl, etags.Songdata)
	if err == NoUpdate {
		body, err = os.ReadFile(songdataJson)
		if err != nil {
			log.Printf("failed to read %s: %s", songdataJson, err)
			return nil, chartStatsResponse{}, err
		}
	} else if err != nil {
		log.Printf("failed to send get request: %s", err)
		return nil, chartStatsResponse{}, err
	} else {
		if err := json.Unmarshal(body, &songdata); err != nil {
			log.Printf("failed to unmarshal music_data.json: %s", err)
			return nil, chartStatsResponse{}, err
		}

		err = os.WriteFile(songdataJson, body, 0644)
		if err != nil {
			return nil, chartStatsResponse{}, err
		}
	}

	var chartstats chartStatsResponse
	chartstatsUrl := DIVING_FISH_API + "chart_stats"
	etag2, body2, err := sendGetRequestWithEtag(chartstatsUrl, etags.Chartstats)
	if err == NoUpdate {
		body, err = os.ReadFile(chartstatsJson)
		if err != nil {
			log.Printf("failed to read %s: %s", chartstatsJson, err)
			return nil, chartStatsResponse{}, err
		}
	} else if err != nil {
		log.Printf("failed to send get request: %s", err)
		return nil, chartStatsResponse{}, err
	} else {
		if err := json.Unmarshal(body2, &chartstats); err != nil {
			log.Printf("failed to unmarshal chart_stats.json: %s", err)
			return nil, chartStatsResponse{}, err
		}

		err = os.WriteFile(chartstatsJson, body2, 0644)
		if err != nil {
			return nil, chartStatsResponse{}, err
		}
	}

	err = saveEtags(divingFishEtags{
		Songdata:   etag,
		Chartstats: etag2,
	})
	if err != nil {
		return nil, chartStatsResponse{}, err
	}

	return songdata, chartstats, nil
}

func sendGetRequestWithEtag(url string, etag string) (string, []byte, error) {
	client := &http.Client{
		Timeout: HTTP_REQUEST_TIMEOUT,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("failed make get request: %s", err)
		return etag, nil, err
	}

	req.Header.Set("If-None-Match", etag)

	res, err := client.Do(req)
	if err != nil {
		log.Printf("failed to get %s: %s", url, err)
		return etag, nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotModified {
		log.Printf("no new update on server, using saved json...")
		return etag, nil, NoUpdate
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("failed to read body of %s: %s", url, err)
		return res.Header.Get("etag"), nil, NoUpdate
	}

	return res.Header.Get("etag"), body, nil
}
func loadEtags() (divingFishEtags, error) {
	var cfg divingFishEtags
	f, err := os.Open(DIVING_FISH_ETAGS_JSON)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&cfg)
	return cfg, err
}

func saveEtags(t divingFishEtags) error {
	jsondata, err := json.Marshal(t)
	if err != nil {
		log.Printf("failed to marshal etags: %s", err)
		return err
	}

	err = os.WriteFile(DIVING_FISH_ETAGS_JSON, jsondata, 0644)
	if err != nil {
		log.Printf("failed to write etags: %s", err)
		return err
	}
	return nil
}

// fills the songdata.db
func PopulateSongDataCN(cfg *config.ApiConfig) error {
	songs, chartstats, err := pullJsonCN()
	if err != nil {
		log.Printf("failed to pull data from diving-fish server")
		return err
	}

	log.Printf("populating songdataCN...")
	for _, song := range songs {
		for _, d := range song.Level {
			songid, err := strconv.Atoi(song.ID)
			if err != nil {
				log.Printf("CreateSongCN for %s failed: %s", song.ID, err)
				return err
			}
			params := songdatadb.CreateSongCNParams{
				ID:    int64(songid),
				Diff:  d,
				Title: song.Title,
				IsDx:  strings.Contains(song.Type, "DX"),
			}
			err = cfg.SongdataDBQueries.CreateSongCN(context.Background(), params)
			if err != nil {
				log.Printf("CreateSongCN failed: %s", err)
				return err
			}
		}
	}
	log.Printf("finished populating songdataCN")
	log.Printf("populating chartstatsCN...")
	for key, stat := range chartstats.Charts {
		songid, err := strconv.Atoi(key)
		if err != nil {
			log.Printf("CreateStat for %s failed: %s", key, err)
			return err
		}
		for _, s := range stat {
			if int64(s.Cnt) == 0 {
				continue
			}
			params := songdatadb.CreateChartStatCNParams{
				SongID: int64(songid),
				Diff:   s.Diff,
				Count:  int64(s.Cnt),
				StdDev: s.StdDev,
			}
			err := cfg.SongdataDBQueries.CreateChartStatCN(context.Background(), params)
			if err != nil {
				log.Printf("CreateStat for %s failed: %s", key, err)
				return err
			}
		}
	}
	log.Printf("finished populating chartstatsCN")
	return nil
}

type divingFishSong struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Type  string   `json:"type"`
	Level []string `json:"level"`
}

type divingFishChartStat struct {
	Cnt     float64 `json:"cnt"`
	Diff    string  `json:"diff"`
	FitDiff float64 `json:"fit_diff"`
	StdDev  float64 `json:"std_dev"`
}

type chartStatsResponse struct {
	Charts map[string][]divingFishChartStat `json:"charts"`
}

type divingFishEtags struct {
	Songdata   string `json:"songdata"`
	Chartstats string `json:"chartstats"`
}
