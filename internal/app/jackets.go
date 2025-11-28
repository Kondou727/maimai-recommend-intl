package app

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/Kondou727/maimai-recommend-intl/internal/config"
	"golang.org/x/sync/semaphore"
)

func PullJackets(cfg *config.ApiConfig) error {
	var wg sync.WaitGroup

	const JACKET_PATH = "resources/jackets/"
	err := os.MkdirAll(JACKET_PATH, 0755)
	if err != nil {
		log.Printf("failed to create jacket path: %s", err)
		return err
	}

	jackets, err := cfg.SongdataDBQueries.ReturnAllJackets(context.Background())
	if err != nil {
		log.Printf("failed to get jackets: %s", err)
		return err
	}

	sem := semaphore.NewWeighted(MAX_SIMUTANOUS_DOWNLOADS)
	for _, j := range jackets {
		if err := sem.Acquire(context.Background(), 1); err != nil {
			log.Printf("Failed to acquire semaphore: %v", err)
			break
		}

		wg.Go(func() {
			defer sem.Release(1)
			link := SERVER_MUSIC_JACKET_BASE_URL + j + ".png"
			filename := JACKET_PATH + j + ".png"

			f, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
			if err != nil {
				if os.IsExist(err) {
					log.Printf("%s.png already exists, skipping...", j)
					return
				}
				log.Printf("failed to open file: %s", err)
				return
			}
			defer f.Close()

			res, err := http.Get(link)
			if err != nil {
				log.Printf("failed to get %s: %s", link, err)
				defer os.Remove(filename)
				return
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)

			if err != nil {
				log.Printf("failed to process image: %s", err)
				return
			}

			_, err = f.Write(body)
			if err != nil {
				log.Printf("failed to write to file: %s", err)
				return
			}
		})
	}
	wg.Wait()
	log.Printf("finished getting image files")
	return nil
}
