package main

import (
	"context"
	"fmt"

	"github.com/magic_bubblez/link-hoarder/internal/database"
	"github.com/magic_bubblez/link-hoarder/internal/scraper"
)

type ScrapeJob struct {
	URL    string
	LinkID string
}

var ScrapeJobs chan ScrapeJob

func StartScrapeWorkers(numWorkers int, bufferSize int) {
	ScrapeJobs = make(chan ScrapeJob, bufferSize)
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for job := range ScrapeJobs {
				processScrapeJob(workerID, job)
			}
		}(i)
	}
}

func processScrapeJob(workerID int, job ScrapeJob) {
	bgctx := context.Background()
	title, image, err := scraper.ScrapeTitle(job.URL)
	if err != nil {
		fmt.Printf("[Worker %d] Failed to scrape %s: %v\n", workerID, job.URL, err)
		return
	}
	var titleStr, imageStr string
	if title != nil {
		titleStr = *title
	}
	if image != nil {
		imageStr = *image
	}

	if titleStr != "" || imageStr != "" {
		if err := database.UpdateLinkData(bgctx, job.LinkID, titleStr, imageStr); err != nil {
			fmt.Printf("[Worker %d] DB update failed: %v\n", workerID, err)
		} else {
			fmt.Printf("[Worker %d] Updated link %s\n", workerID, job.LinkID)
		}
	}
}
