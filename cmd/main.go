package main

import (
	"log"
	"os"

	"github.com/knchan0x/belle-maison/cmd/scheduler"
	"github.com/knchan0x/belle-maison/cmd/web"
	"github.com/knchan0x/belle-maison/internal/db"
	"github.com/knchan0x/belle-maison/internal/email"
	"github.com/spf13/viper"
)

func init() {
	log.Println("Initializing, please wait...")

	path := ""
	_, err := os.Stat("../config.yaml")
	if err == nil {
		path = "../config.yaml"
	} else {
		path = "./config.yaml"
	}

	viper.SetConfigFile(path)
	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalln("Config file not found")
		} else {
			log.Fatalf("Fatal error config file: %s \n", err)
		}
	}
}

func main() {

	// set db connection
	db.SetDebugMode(viper.GetBool("debug.mode"))
	dbClient, err := db.NewClient(&db.DbSettings{
		Host:     viper.GetString("mysql.host"),
		Port:     viper.GetString("mysql.port"),
		DB:       viper.GetString("mysql.db"),
		User:     viper.GetString("mysql.user"),
		Password: viper.GetString("mysql.password"),
	})

	if err != nil {
		panic("failed to connect database")
	}

	// config email service

	provider := viper.GetString("email.provider")
	if provider == "outlook" || provider == "Outlook" {
		provider = email.Outlook
	}
	email.ConfigEmailService(&email.EmailSetting{
		ServiceProvider: provider,
		Host:            viper.GetString("email.smpt.host"),
		Port:            viper.GetInt("email.smpt.port"),
		User:            viper.GetString("email.sender.username"),
		Password:        viper.GetString("email.sender.password"),
		Recipients:      viper.GetStringSlice("email.recipients"),
	})

	// migrate schemas
	db.Migrate(dbClient)

	// set data handler
	dataHandler := db.GetHandler(dbClient)

	// set schedule
	s := scheduler.NewScheduler(dataHandler)
	s.RunAsync()

	// set api gateway and dashboard
	web.SetDebugMode(viper.GetBool("debug.mode"))
	web.ActivateSimpleAuth(!viper.GetBool("debug.mode"))
	web := web.NewHandler(dataHandler, &web.User{
		Username: viper.GetString("dashboard.username"),
		Password: viper.GetString("dashboard.password"),
	})
	web.Run(":80")
}
