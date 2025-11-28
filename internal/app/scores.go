package app

import (
	"context"
	"database/sql"
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Kondou727/maimai-recommend-intl/internal/config"
	scoresdb "github.com/Kondou727/maimai-recommend-intl/internal/database/scores"
	"github.com/gocarina/gocsv"
	"github.com/pressly/goose"
)

type RawScores struct {
	SongName    string `csv:"Song"`
	ChartType   string `csv:"Chart"`
	Difficulty  string `csv:"Difficulty"`
	Achievement string `csv:"Achv"`
	FCAP        string `csv:"FC/AP"`
	Sync        string `csv:"Sync"`
	DXStar      int    `csv:"DX ✦"`
	DXPercent   string `csv:"DX %"`
}

func scoresTSVToStruct(tsvPath string) ([]RawScores, error) {
	tsvFile, err := os.Open(tsvPath)
	if err != nil {
		return nil, err
	}
	csvreader := csv.NewReader(tsvFile)
	csvreader.Comma = '\t'
	csvreader.LazyQuotes = true
	var rs []RawScores
	if err := gocsv.UnmarshalCSV(csvreader, &rs); err != nil {
		return nil, err
	}

	return rs, nil
}

func ImportTSVToDB(input string, cfg *config.ApiConfig) error {
	log.Print("processing tsv..")
	scores, err := scoresTSVToStruct(input)
	if err != nil {
		return err
	}

	for _, s := range scores {
		achievementInt, err := percentToInt(s.Achievement)
		if err != nil {
			return err
		}

		dxpercentInt, err := percentToInt(s.DXPercent)
		if err != nil {
			return err
		}

		_, err = cfg.ScoresDBQueries.CreateScore(context.Background(), scoresdb.CreateScoreParams{
			SongName:    s.SongName,
			ChartType:   s.ChartType,
			Difficulty:  s.Difficulty,
			Achievement: int64(achievementInt),
			FcAp:        s.FCAP,
			Sync:        s.Sync,
			DxStar:      int64(s.DXStar),
			DxPercent:   int64(dxpercentInt),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func percentToInt(s string) (int, error) {
	return strconv.Atoi(strings.ReplaceAll(strings.ReplaceAll(s, "%", ""), ".", ""))
}

func LoadScoresDB() (*sql.DB, error) {
	scoresDB, err := sql.Open("sqlite", DBFILE)
	if err != nil {
		return nil, err
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}
	if err := goose.Up(scoresDB, "sql/scores/schema"); err != nil {
		return nil, err
	}
	return scoresDB, nil

}
