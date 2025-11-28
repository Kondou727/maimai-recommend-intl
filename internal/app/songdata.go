package app

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Kondou727/maimai-recommend-intl/internal/config"
	songdatadb "github.com/Kondou727/maimai-recommend-intl/internal/database/songdata"
	"github.com/pressly/goose"
)

// Fills the songdata.db
func PopulateSongData(cfg *config.ApiConfig) error {

	songs, records, err := pullJsonINTL()
	if err != nil {
		log.Printf("failed to pull official json")
		return err
	}
	log.Printf("populating songdata...")
	for _, r := range records {
		string_level := string(r.Level)
		cleaned_level := strings.ReplaceAll(strings.ReplaceAll(string_level, ".0", ""), ".6", "+")

		params := songdatadb.CreateSongParams{
			ID:      r.ID,
			Title:   r.Title,
			Artist:  r.Artist,
			Genre:   r.Genre,
			Img:     r.Img,
			Release: r.Release,
			Version: r.Version,
			IsDx:    r.IsDx,
			Diff:    r.Diff,
			Level:   cleaned_level,
			Const:   string(r.Const),
			IsUtage: false,
			IsBuddy: sql.NullString{String: "", Valid: true},
		}
		err := cfg.SongdataDBQueries.CreateSong(context.Background(), params)
		if err != nil {
			log.Printf("CreateSong failed: %s", err)
			return err
		}
	}

	for _, s := range songs {
		if s.Catcode == "宴会場" {
			params := songdatadb.CreateSongParams{
				ID:      s.Sort,
				Title:   s.Title,
				Artist:  s.Artist,
				Genre:   s.Catcode,
				Img:     strings.ReplaceAll(s.ImageURL, ".png", ""),
				Release: "",
				Version: "",
				IsDx:    true,
				Level:   s.LevUtage,
				Const:   "",
				IsUtage: true,
				IsBuddy: sql.NullString{String: s.Buddy, Valid: true},
			}
			err := cfg.SongdataDBQueries.CreateSong(context.Background(), params)
			if err != nil {
				log.Printf("CreateSong failed: %s", err)
				return err
			}
		}
	}
	log.Printf("finished populating songdata")
	return nil

}

// Pulls the jsons
func pullJsonINTL() (maimai_songs, reiwa_songs, error) {
	res, err := http.Get(MAIMAI_SONGS_JSON_LINK)
	if err != nil {
		log.Printf("get request to official maimai server failed: %s", err)
		return nil, nil, err
	}
	defer res.Body.Close()

	res2, err := http.Get(REIWA_JSON_LINK)
	if err != nil {
		log.Printf("get request to reiwa server failed: %s", err)
		return nil, nil, err
	}
	defer res2.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("failed to read get request body: %s", err)
		return nil, nil, err
	}

	body2, err := io.ReadAll(res2.Body)
	if err != nil {
		log.Printf("failed to read get request body: %s", err)
		return nil, nil, err
	}

	var songs maimai_songs
	var songs2 reiwa_songs

	cleaned_body := bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
	cleaned_body2 := bytes.TrimPrefix(body2, []byte("\xef\xbb\xbf"))

	if err := json.Unmarshal(cleaned_body, &songs); err != nil {
		log.Printf("failed to unmarshal maimai_songs.json: %s", err)
		return nil, nil, err
	}

	if err := json.Unmarshal(cleaned_body2, &songs2); err != nil {
		log.Printf("failed to unmarshal maimai_record.json: %s", err)
		return nil, nil, err
	}
	return songs, songs2, nil

}

func LoadSongdataDB() (*sql.DB, error) {
	songdataDB, err := sql.Open("sqlite", "songdata.db")
	if err != nil {
		return nil, err
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}
	if err := goose.Up(songdataDB, "sql/songdata/schema"); err != nil {
		return nil, err
	}
	return songdataDB, nil

}

type reiwa_songs []struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	Artist    string      `json:"artist"`
	Genre     string      `json:"genre"`
	Img       string      `json:"img"`
	Release   string      `json:"release"`
	Version   string      `json:"version"`
	IsDx      bool        `json:"is_dx"`
	Diff      string      `json:"diff"`
	Level     json.Number `json:"level"`
	Const     json.Number `json:"const"`
	IsUnknown int         `json:"is_unknown"`
}
type maimai_songs []struct {
	Artist     string `json:"artist"`
	Catcode    string `json:"catcode"`
	ImageURL   string `json:"image_url"`
	Release    string `json:"release"`
	LevBas     string `json:"lev_bas,omitempty"`
	LevAdv     string `json:"lev_adv,omitempty"`
	LevExp     string `json:"lev_exp,omitempty"`
	LevMas     string `json:"lev_mas,omitempty"`
	Sort       string `json:"sort"`
	Title      string `json:"title"`
	TitleKana  string `json:"title_kana"`
	Version    string `json:"version"`
	LevRemas   string `json:"lev_remas,omitempty"`
	DxLevBas   string `json:"dx_lev_bas,omitempty"`
	DxLevAdv   string `json:"dx_lev_adv,omitempty"`
	DxLevExp   string `json:"dx_lev_exp,omitempty"`
	DxLevMas   string `json:"dx_lev_mas,omitempty"`
	Key        string `json:"key,omitempty"`
	DxLevRemas string `json:"dx_lev_remas,omitempty"`
	Date       string `json:"date,omitempty"`
	LevUtage   string `json:"lev_utage,omitempty"`
	Kanji      string `json:"kanji,omitempty"`
	Comment    string `json:"comment,omitempty"`
	Buddy      string `json:"buddy,omitempty"`
}
