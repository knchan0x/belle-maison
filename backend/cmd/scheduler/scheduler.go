package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/knchan0x/belle-maison/backend/internal/crawler"
	"github.com/knchan0x/belle-maison/backend/internal/db/model/product"
	"github.com/knchan0x/belle-maison/backend/internal/db/model/target"
	"github.com/knchan0x/belle-maison/backend/internal/email"
	"gorm.io/gorm"
)

// scheduler runs jobs according to pre-defined schedule.
// It is a wrapper of *gocron.Scheduler.
type scheduler struct {
	*gocron.Scheduler
	crawler  crawler.Crawler
	dbClient *gorm.DB
	jobs     []string // tasks pending to perform
}

// NewScheduler returns new scheduler
func NewScheduler(dbClient *gorm.DB) *scheduler {
	c, err := crawler.NewCrawler()
	if err != nil {
		log.Fatalf("failed to initialize crawler: %v", err)
	}

	s := &scheduler{
		crawler:  c,
		dbClient: dbClient,
		jobs:     []string{},
	}
	s.Scheduler = gocron.NewScheduler(time.UTC)
	return s
}

// StartScraping activates crawler to preform scraping tasks
func (s *scheduler) StartScraping() {
	log.Println("Start scraping...")
	if s.jobs == nil || len(s.jobs) <= 0 {
		log.Println("No job need to be performed, end scraping")
		return
	}

	// fetch
	results := s.crawler.Scraping(s.jobs...)
	s.cleanJobs()

	for _, result := range results {
		if result.Err != nil && result.Err != crawler.PRODUCT_NOT_FOUND {
			s.jobs = append(s.jobs, result.ProductCode)
			continue
		}

		p, err := product.GetProductByCode(s.dbClient, result.ProductCode)
		if err != nil && err != gorm.ErrRecordNotFound {
			s.jobs = append(s.jobs, result.ProductCode)
			continue
		}
		if err == gorm.ErrRecordNotFound {
			if _, err := product.New(s.dbClient, result); err != nil {
				s.jobs = append(s.jobs, result.ProductCode)
			}
			continue
		}

		if err := p.Update(s.dbClient, result); err != nil {
			s.jobs = append(s.jobs, result.ProductCode)
		}
	}
	log.Println("Done")
}

const (
	LowStockThreshold = 9
)

func (s *scheduler) GenerateDailyReport() {
	log.Println("Generating daily report...")
	targets := target.GetAll(s.dbClient)
	emailMsg := ""
	for _, target := range targets {
		// hit target
		if target.Price <= target.TargetPrice {
			if emailMsg == "" {
				emailMsg += "The following products have achieved your target price: \n"
			}
			emailMsg += fmt.Sprintf("%s: target price: %d, current price: %d\n", target.Name, target.TargetPrice, target.Price)
		}
		// low stock
		if target.Price >= target.TargetPrice && target.Stock <= LowStockThreshold {
			if emailMsg == "" {
				emailMsg += "The following products have not achieved your target price but the stock is low now: \n"
			}
			emailMsg += fmt.Sprintf("%s: target price: %d, current price: %d", target.Name, target.TargetPrice, target.Price)
		}
	}

	if emailMsg != "" {
		if err := email.SendEmail("Belle Maison Price Tracker", emailMsg); err != nil {
			log.Println(err)
		}
	}

	log.Println("Done")
}

func (s *scheduler) cleanJobs() {
	s.jobs = []string{}
}

func (s *scheduler) assignJobs() {
	targets := target.GetList(s.dbClient)
	s.jobs = append(s.jobs, targets...)
}
