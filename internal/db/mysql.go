package db

import (
	"fmt"
	"log"
	"time"

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
}

var debugMode = false

// NewClient return *gorm.DB with DbSettings provided.
// It will also set MaxIdleConns=10, MaxOpenConns=20, ConnMaxLifetime=1hour implicitly.
func NewClient(settings *DbSettings) (*gorm.DB, error) {
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
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(20)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	return db, nil
}

// Auto migrate following schemas:
// Product, Style, Price, Target
func Migrate(dbClient *gorm.DB) {
	if err := dbClient.AutoMigrate(&Product{}); err != nil {
		log.Panicf("failed to migrate Product: %v", err)
	}
	if err := dbClient.AutoMigrate(&Style{}); err != nil {
		log.Panicf("failed to migrate Style: %v", err)
	}
	if err := dbClient.AutoMigrate(&Price{}); err != nil {
		log.Panicf("failed to migrate Price: %v", err)
	}
	if err := dbClient.AutoMigrate(&Target{}); err != nil {
		log.Panicf("failed to migrate Target: %v", err)
	}
}

// SetDebugMode set the LogMode of gorm's Logger
func SetDebugMode(isDebugMode bool) {
	debugMode = isDebugMode
}
