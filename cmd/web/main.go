package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/knchan0x/belle-maison/cmd/web/auth"
	"github.com/knchan0x/belle-maison/cmd/web/controller"
	"github.com/knchan0x/belle-maison/cmd/web/middleware"
	"github.com/knchan0x/belle-maison/cmd/web/user"
	"github.com/knchan0x/belle-maison/internal/config"
	"github.com/knchan0x/belle-maison/internal/db"
	"github.com/knchan0x/belle-maison/internal/email"
	"github.com/knchan0x/belle-maison/internal/scraper"
)

const (
	cookie_name = "_cookie_"

	filePath_root_docker    = "./static"
	filePath_root_localhost = "./static"

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

	addr := ":80"
	if config.GetInt("web.port") != 0 {
		addr = fmt.Sprintf(":%d", config.GetInt("web.port"))
	}

	// config db connection
	db.SetDebugMode(config.GetBool("debug"))
	dbClient, err := db.NewGORMClient(&db.DbSettings{
		Host:     config.GetString("mysql.host"),
		Port:     config.GetString("mysql.port"),
		DB:       config.GetString("mysql.db"),
		User:     config.GetString("mysql.user"),
		Password: config.GetString("mysql.password"),
		PoolSize: 20,
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

	// set user
	user.SetAdmin(config.GetString("admin.username"), config.GetString("admin.password"))

	// set auth
	auth.SetCookieName(cookie_name)

	// activate auth middleware
	middleware.ActivateRolePermit(!config.GetBool("debug"))

	// configure scraper
	scraper, err := scraper.NewScraper()
	if err != nil {
		log.Fatalf("failed to initialize scraper: %v", err)
	}

	// set gin mode
	gin.SetMode(gin.ReleaseMode)

	// configure gin
	web := gin.Default()
	if config.GetBool("debug") {
		web.Use(middleware.AllowCrossOrigin("http://localhost:3000"))
	}

	// add root for easier configure root path
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
	root.Static(urlPrefix_asset, fileBasePath+"/assets")

	root.StaticFile(urlPrefix_login, fileBasePath+"/login.html")    // GET
	root.POST(urlPrefix_login, controller.Login(urlPath_dashboard)) // POST

	dashboard := root.Group(urlPrefix_dashboard,
		middleware.AccessControl(middleware.Admin, middleware.AuthMode_Redirect, urlPath_login))
	dashboard.StaticFile("/", fileBasePath+"/index.html")

	api := root.Group(urlPrefix_api,
		middleware.AccessControl(middleware.Admin, middleware.AuthMode_Unauthorized))

	// get product info
	api.GET("/product/:productCode",
		middleware.Validate(middleware.ProductCode),
		controller.GetProduct(scraper))

	// POST content: colour, size
	api.POST("/target/:productCode",
		middleware.Validate(middleware.ProductCode),
		middleware.Validate(middleware.TargetColour),
		middleware.Validate(middleware.TargetSize),
		middleware.Validate(middleware.TargetPrice),
		controller.AddTarget(dbClient, scraper))

	api.DELETE("/target/:targetId",
		middleware.Validate(middleware.TargetId),
		controller.DeleteTarget(dbClient))

	// get all products under tracing
	api.GET("/targets",
		middleware.Validate(middleware.QueryPageSize),
		controller.GetTargets(dbClient))

	// set up server
	srv := &http.Server{
		Addr:    addr,
		Handler: web,
	}

	// async, prevent to block the current thread
	// that will handle graceful shutdown
	go func() {
		log.Printf("Running server on %s...", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Unable to start server: %v", err)
		}
	}()

	// wait for interrupt signal
	quit := make(chan os.Signal, 1)

	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	// context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	// close db connection before exit
	if sqlDB, err := dbClient.DB(); err == nil {
		sqlDB.Close()
	}

	log.Println("Server exiting")
}
