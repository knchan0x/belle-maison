package controller

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/knchan0x/belle-maison/cmd/web/middleware"
	"github.com/knchan0x/belle-maison/internal/cache"
	"github.com/knchan0x/belle-maison/internal/db/model/product"
	"github.com/knchan0x/belle-maison/internal/db/model/target"
	"github.com/knchan0x/belle-maison/internal/scraper"
	"gorm.io/gorm"
)

const (
	targets_cache_key = "target"
)

// add target
func AddTarget(dbClient *gorm.DB, s scraper.Scraper) func(*gin.Context) {
	return func(ctx *gin.Context) {
		productCode := ctx.GetString(middleware.Validated_ProductCode)

		// get most updated product info
		var wg sync.WaitGroup
		wg.Add(2)

		var r *scraper.Result
		go func() {
			if c, ok := cache.Get("scraper_result_" + productCode); ok {
				r = c.(*scraper.Result)
			} else {
				r = s.Scraping(productCode)[0]
				cache.Add("scraper_result_"+productCode, r, time.Hour)
			}
			wg.Done()
		}()

		var p *product.Product
		var err error
		go func() {
			p, err = product.GetProductByCode(dbClient, productCode)
			wg.Done()
		}()

		wg.Wait()

		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server rrror"})
				return
			}

			p, err = product.New(dbClient, r)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server rrror"})
				return
			}
		}

		targetStyle, err := p.Style(
			dbClient,
			ctx.GetString(middleware.Validated_TargetColour),
			ctx.GetString(middleware.Validated_TargetSize))

		// style not found
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameters"})
			return
		}

		if _, err := target.New(
			dbClient,
			productCode,
			p.ID,
			targetStyle.ID,
			uint(ctx.GetInt(middleware.Validated_TargetPrice))); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server rrror"})
			return
		}

		cache.Delete(targets_cache_key) // force update target list
		ctx.Status(http.StatusCreated)
	}
}

// get all products under tracing
// params: page, size
func GetTargets(dbClient *gorm.DB) func(*gin.Context) {
	return func(ctx *gin.Context) {

		page := ctx.GetInt(middleware.Validated_QueryPage)
		size := ctx.GetInt(middleware.Validated_QuerySize)

		var targets []target.TargetInfo
		if t, ok := cache.Get(targets_cache_key); ok {
			targets = t.([]target.TargetInfo)
		} else {
			targets = target.GetAll(dbClient)
			cache.Add(targets_cache_key, targets, time.Hour*24)
		}

		targetSize := len(targets)

		if targetSize > size {
			if size*page > targetSize {
				ctx.JSON(http.StatusOK, targets[size*(page-1):])
			} else {
				ctx.JSON(http.StatusOK, targets[size*(page-1):size*page])
			}
		} else {
			ctx.JSON(http.StatusOK, targets)
		}
	}
}

func DeleteTarget(dbClient *gorm.DB) func(*gin.Context) {
	return func(ctx *gin.Context) {

		id := ctx.GetInt(middleware.Validated_TargetId)
		t, err := target.GetById(dbClient, uint(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
				return
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		t.Delete(dbClient)
		cache.Delete(targets_cache_key) // force update target list
		ctx.Status(http.StatusNoContent)
	}
}
