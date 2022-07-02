package db

import (
	"fmt"
	"log"
	"time"

	p "github.com/knchan0x/belle-maison/internal/db/model/product"
	t "github.com/knchan0x/belle-maison/internal/db/model/target"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DbSettings struct {
	Host     string
	Port     string
	DB       string
	User     string
	Password string
	Mode     string
	PoolSize int
}

var debugMode = false

// NewGORMClient return *gorm.DB with DbSettings provided.
// It will also set MaxIdleConns=10, MaxOpenConns=20, ConnMaxLifetime=1hour implicitly.
func NewGORMClient(settings *DbSettings) (*gorm.DB, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True", settings.User, settings.Password, settings.Host, settings.Port, settings.DB)

	var level logger.LogLevel
	if debugMode {
		level = logger.Info
	} else {
		level = logger.Error
	}

	db, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{
		Logger: logger.Default.LogMode(level),
	})
	if err != nil {
		return nil, err
	}

	// set to default if no provided
	if settings.PoolSize == 0 {
		settings.PoolSize = 10
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(settings.PoolSize / 2)
		sqlDB.SetMaxOpenConns(settings.PoolSize)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	return db, nil
}

// Auto migrate following schemas:
// Product, Style, Price, Target
func Migrate(dbClient *gorm.DB) {
	if err := dbClient.AutoMigrate(&p.Product{}); err != nil {
		log.Panicf("failed to migrate Product: %v", err)
	}
	if err := dbClient.AutoMigrate(&p.Style{}); err != nil {
		log.Panicf("failed to migrate Style: %v", err)
	}
	if err := dbClient.AutoMigrate(&p.Price{}); err != nil {
		log.Panicf("failed to migrate Price: %v", err)
	}
	if err := dbClient.AutoMigrate(&t.Target{}); err != nil {
		log.Panicf("failed to migrate Target: %v", err)
	}
}

// SetDebugMode set the LogMode of gorm's Logger
func SetDebugMode(isDebugMode bool) {
	debugMode = isDebugMode
}
