package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/knchan0x/belle-maison/cmd/web/auth"
	"github.com/knchan0x/belle-maison/cmd/web/middleware"
	"github.com/knchan0x/belle-maison/cmd/web/route"
	"github.com/knchan0x/belle-maison/internal/config"
	"github.com/knchan0x/belle-maison/internal/db"
	"github.com/knchan0x/belle-maison/internal/email"
	"github.com/knchan0x/belle-maison/internal/scraper"
)

const (
	cookie_name = "_cookie_"
	addr        = ":80"

	filePath_root_docker    = "./page"
	filePath_root_localhost = "../../page"

	urlPath_root        = "/bellemasion"
	urlPrefix_dashboard = "/dashboard"
	urlPrefix_login     = "/login"
	urlPrefix_api       = "/api"
	urlPrefix_asset     = "/assets"
)

var (
	urlPath_dashboard = urlPath_root + urlPrefix_dashboard
	urlPath_login     = urlPath_root + urlPrefix_login
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

	// set data handler
	dataHandler := db.GetHandler(dbClient)

	// set auth
	auth.SetAdmin(config.GetString("dashboard.username"), config.GetString("dashboard.password"))
	auth.SetCookieName(cookie_name)

	// activate auth middleware
	middleware.ActivateSimpleAuth(!config.GetBool("debug.mode"))

	// configure scraper
	scraper, err := scraper.NewScraper()
	if err != nil {
		log.Fatalf("failed to initialize scraper: %v", err)
	}

	// set gin mode
	gin.SetMode(gin.ReleaseMode)

	// configure gin
	web := gin.Default()
	if config.GetBool("debug.mode") {
		web.Use(middleware.AllowCrossOrigin("http://localhost:3000"))
	}

	// add root for easier configuration root path
	root := web.Group(urlPath_root)

	// redirect / to /dashboard
	root.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, urlPath_dashboard)
	})

	fileBasePath := ""
	if _, err := os.Stat(filePath_root_docker + "/login.html"); err == nil {
		fileBasePath = filePath_root_docker
	} else {
		fileBasePath = filePath_root_localhost
	}
	root.StaticFile(urlPrefix_login, fileBasePath+"/login.html")
	root.POST(urlPrefix_login, route.Login(urlPath_dashboard))

	api := root.Group(urlPrefix_api, middleware.SimpleAuth(middleware.AuthMode_Unauthorized))

	// get product info
	api.GET("/product/:productCode",
		middleware.Validate(middleware.ProductCode),
		route.GetProduct(scraper))

	// POST content: colour, size
	api.POST("/target/:productCode",
		middleware.Validate(middleware.ProductCode),
		middleware.Validate(middleware.TargetColour),
		middleware.Validate(middleware.TargetSize),
		middleware.Validate(middleware.TargetPrice),
		route.AddTarget(scraper, dataHandler))

	api.DELETE("/target/:targetId",
		middleware.Validate(middleware.TargetId),
		route.DeleteTarget(dataHandler))

	api.PATCH("/target/:targetId",
		middleware.Validate(middleware.TargetId),
		middleware.Validate(middleware.ProductId),
		middleware.Validate(middleware.TargetColour),
		middleware.Validate(middleware.TargetSize),
		route.UpdateTarget(dataHandler))

	// get all products under tracing
	api.GET("/targets",
		middleware.Validate(middleware.QueryPageSize),
		route.GetTargets(dataHandler))

	dashboard := root.Group(urlPrefix_dashboard, middleware.SimpleAuth(middleware.AuthMode_Redirect, urlPath_login))
	dashboard.StaticFile("/", fileBasePath+"/index.html")

	root.Static(urlPrefix_asset, fileBasePath+"/assets")

	log.Printf("Running web server on %s...", addr)
	if err := web.Run(addr); err != nil {
		log.Fatalf("Unable to start API Server: %v", err)
	}
}
