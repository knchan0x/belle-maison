package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/knchan0x/belle-maison/backend/cmd/web/middleware"
	"github.com/knchan0x/belle-maison/backend/internal/cache"
	"github.com/knchan0x/belle-maison/backend/internal/scraper"
)

const (
	cachePrefix = "scraper_result_"
)

// get product info
func GetProduct(s scraper.Scraper) func(*gin.Context) {

	return func(ctx *gin.Context) {
		productCode := ctx.GetString(middleware.Validated_ProductCode)
		var r *scraper.Result
		if c, ok := cache.Get(cachePrefix + productCode); ok {
			r = c.(*scraper.Result)
		} else {
			r = s.Scraping(productCode)[0]
			cache.Add(cachePrefix+productCode, r, time.Hour)
		}

		if r.Err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": r.Err})
			return
		}
		if r.Product == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}

		ctx.JSON(http.StatusOK, r)
	}
}
