package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/knchan0x/belle-maison/internal/db"
	"github.com/knchan0x/belle-maison/internal/email"
	"github.com/knchan0x/belle-maison/internal/scraper"
)

// scheduler runs jobs according to pre-defined schedule
type scheduler struct {
	scheduler   *gocron.Scheduler
	scraper     scraper.Scraper
	dataHandler db.Handler
	jobs        []string // tasks pending to perform
}

// NewScheduler returns new scheduler
func NewScheduler(dataHandler db.Handler) *scheduler {
	c, err := scraper.NewScraper()
	if err != nil {
		log.Fatalf("failed to initialize scraper: %v", err)
	}

	return &scheduler{
		scheduler:   gocron.NewScheduler(time.UTC),
		scraper:     c,
		dataHandler: dataHandler,
		jobs:        []string{},
	}
}

// Run runs Scheduler, it will block the current thread.
// Schedule defined as following
//
// - clean-tasks at 23:59 UTC and schedule tasks for coming day at 00:00 UTC
//
// - generate daily report at 04:00 UTC
//
// - crawling (run tasks) every hour starting from 00:30 UTC
//
func (s *scheduler) Run() {
	_, err := s.scheduler.Every(1).Day().At("00:00").Tag("schedule-tasks").Do(s.assignJobs)
	if err != nil {
		log.Printf("schedule-tasks: %v", err)
	}

	_, err = s.scheduler.Every(1).Day().At("23:59").Tag("clean-tasks").Do(s.cleanJobs)
	if err != nil {
		log.Printf("clean-tasks: %v", err)
	}

	_, err = s.scheduler.Every(1).Day().At("04:00").Tag("daily-report").Do(s.GenerateDailyReport)
	if err != nil {
		log.Printf("daily-report: %v", err)
	}

	_, err = s.scheduler.Every(1).Hour().At("00:00").Tag("crawling").Do(s.StartScraping)
	if err != nil {
		log.Printf("crawling: %v", err)
	}

	log.Println("scraper starts working...")
	s.scheduler.StartBlocking()
}

// StartScraping activates scraper to preform crawling tasks
func (s *scheduler) StartScraping() {
	log.Println("start crawling...")
	if s.jobs == nil || len(s.jobs) <= 0 {
		log.Println("no job, end crawling")
		return
	}

	var results []*scraper.Result
	results = s.scraper.ScrapingProducts(s.jobs...)

	for _, result := range results {
		if result.Err != nil && result.Err != scraper.PRODUCT_NOT_FOUND {
			s.jobs = append(s.jobs, result.ProductCode)
		} else {
			err := s.dataHandler.UpdateProduct(result)
			if err != nil && result.Err != db.EMPTY_PRODUCT {
				s.jobs = append(s.jobs, result.ProductCode)
			}
			if result.Err == db.EMPTY_PRODUCT {
				log.Println(db.EMPTY_PRODUCT)
			}
		}
	}
	log.Println("done")
}

func (s *scheduler) GenerateDailyReport() {
	log.Println("generating daily report...")
	targets := s.dataHandler.GetTargets()
	emailMsg := ""
	for _, target := range targets {
		// meet target
		if target.Price <= target.TargetPrice {
			if emailMsg == "" {
				emailMsg += "The following products have achieved your target price: \n"
			}
			emailMsg += fmt.Sprintf("%s: target price: %d, current price: %d\n", target.Name, target.TargetPrice, target.Price)
		}
		// low stock
		if target.Price >= target.TargetPrice && target.Stock < 10 {
			if emailMsg == "" {
				emailMsg += "The following products have not achieved your target price but the stock is low now: \n"
			}
			emailMsg += fmt.Sprintf("%s: target price: %d, current price: %d", target.Name, target.TargetPrice, target.Price)
		}
	}

	if emailMsg != "" {
		if err := email.SendEmail("Belle Masion Price Tracker", emailMsg); err != nil {
			log.Println(err)
		}
	}

	log.Println("done")
}

func (s *scheduler) cleanJobs() {
	s.jobs = []string{}
}

func (s *scheduler) assignJobs() {
	targets := s.dataHandler.GetTargetList()
	s.jobs = append(s.jobs, targets...)
}
