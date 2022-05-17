package main

import (
	"log"

	"github.com/knchan0x/belle-maison/internal/config"
	"github.com/knchan0x/belle-maison/internal/db"
	"github.com/knchan0x/belle-maison/internal/email"
)

func init() {
	log.Println("Initializing, please wait...")
	if err := config.LoadConfig(); err != nil {
		log.Fatalln(err)
	}
}

func main() {

	// config db connection
	db.SetDebugMode(config.GetBool("debug.mode"))
	dbClient, err := db.NewGORMClient(&db.DbSettings{
		Host:     config.GetString("mysql.host"),
		Port:     config.GetString("mysql.port"),
		DB:       config.GetString("mysql.db"),
		User:     config.GetString("mysql.user"),
		Password: config.GetString("mysql.password"),
	})
	if err != nil {
		panic("failed to connect database")
	}

	// config email service
	email.ConfigService(&email.EmailSetting{
		ServiceProvider: config.GetString("email.provider"),
		Host:            config.GetString("email.smpt.host"),
		Port:            config.GetInt("email.smpt.port"),
		User:            config.GetString("email.sender.username"),
		Password:        config.GetString("email.sender.password"),
		Recipients:      config.GetStringSlice("email.recipients"),
	})

	if err := email.Test(config.GetString("email.sender.username")); err != nil {
		log.Panicf("failed to connect email service: %v", err)
	}

	// migrate schemas
	db.Migrate(dbClient)

	// set schedule
	s := NewScheduler(dbClient)
	if _, err := s.Every(1).Day().At("00:00").Tag("schedule-tasks").Do(s.assignJobs); err != nil {
		log.Printf("schedule-tasks: %v", err)
	}
	if _, err := s.Every(1).Day().At("23:59").Tag("clean-tasks").Do(s.cleanJobs); err != nil {
		log.Printf("clean-tasks: %v", err)
	}
	if _, err := s.Every(1).Day().At("04:00").Tag("daily-report").Do(s.GenerateDailyReport); err != nil {
		log.Printf("daily-report: %v", err)
	}
	if _, err := s.Every(1).Hour().At("00:00").Tag("crawling").Do(s.StartScraping); err != nil {
		log.Printf("scraping: %v", err)
	}

	// start scheduler
	log.Println("scraper starts working...")
	s.StartBlocking()
}
