package app

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Kondou727/maimai-stats-tracker/internal/config"
	songdatadb "github.com/Kondou727/maimai-stats-tracker/internal/database/songdata"
)

// gets chinese song data
func pullJsonCN() ([]divingFishSong, chartStatsResponse, error) {
	songdataUrl := DIVING_FISH_API + "music_data"
	res, err := http.Get(songdataUrl)
	if err != nil {
		log.Printf("get songdata request to diving-fish server failed: %s", err)
		return nil, chartStatsResponse{}, err
	}
	defer res.Body.Close()

	chartstatsUrl := DIVING_FISH_API + "chart_stats"
	res2, err := http.Get(chartstatsUrl)
	if err != nil {
		log.Printf("get chart stats request to diving-fish server failed: %s", err)
		return nil, chartStatsResponse{}, err
	}
	defer res2.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("failed to read songdata request body: %s", err)
		return nil, chartStatsResponse{}, err
	}

	body2, err := io.ReadAll(res2.Body)
	if err != nil {
		log.Printf("failed to read chart stats request body: %s", err)
		return nil, chartStatsResponse{}, err
	}

	var songdata []divingFishSong
	var chartstats chartStatsResponse

	/*
		// Strips UTF-8 BOM if needed
		cleaned_body := bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
		cleaned_body2 := bytes.TrimPrefix(body2, []byte("\xef\xbb\xbf"))
	*/

	if err := json.Unmarshal(body, &songdata); err != nil {
		log.Printf("failed to unmarshal music_data.json: %s", err)
		return nil, chartStatsResponse{}, err
	}

	if err := json.Unmarshal(body2, &chartstats); err != nil {
		log.Printf("failed to unmarshal chart_stats.json: %s", err)
		return nil, chartStatsResponse{}, err
	}
	return songdata, chartstats, nil
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
