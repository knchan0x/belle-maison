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
		log.Panicf("failed to connect database: %v", err)
	}

	// config email service
	email.ConfigEmailService(&email.EmailSetting{
		ServiceProvider: config.GetString("email.provider"),
		Host:            config.GetString("email.smpt.host"),
		Port:            config.GetInt("email.smpt.port"),
		User:            config.GetString("email.sender.username"),
		Password:        config.GetString("email.sender.password"),
		Recipients:      config.GetStringSlice("email.recipients"),
	})

	if err := email.TestEmailService(config.GetString("email.sender.username")); err != nil {
		log.Panicf("failed to connect email service: %v", err)
	}

	// migrate schemas
	db.Migrate(dbClient)

	// set data handler
	dataHandler := db.GetHandler(dbClient)

	// set api gateway and dashboard
	SetDebugMode(config.GetBool("debug.mode"))
	ActivateSimpleAuth(!config.GetBool("debug.mode"))
	web := NewHandler(dataHandler, &User{
		Username: config.GetString("dashboard.username"),
		Password: config.GetString("dashboard.password"),
	})
	web.Run(":80")
}
