package route

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/knchan0x/belle-maison/cmd/web/middleware"
	"github.com/knchan0x/belle-maison/internal/cache"
	"github.com/knchan0x/belle-maison/internal/db"
	"github.com/knchan0x/belle-maison/internal/scraper"
	"gorm.io/gorm"
)

const (
	targets_cache_key = "target"
)

// add target
func AddTarget(s scraper.Scraper, dataHandler db.Handler) func(*gin.Context) {
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

		var p *db.Product
		var err error
		go func() {
			p, err = dataHandler.GetProductAndStylesByProductCode(productCode)
			wg.Done()
		}()

		wg.Wait()

		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			p, err = dataHandler.CreateProduct(r)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		// create target
		newTarget := db.Target{
			ProductCode: productCode,
			ProductID:   p.ID,
			TargetPrice: uint(ctx.GetInt(middleware.Validated_TargetPrice)),
		}

		targetStyle := ctx.GetString(middleware.Validated_TargetColour) + "-" + ctx.GetString(middleware.Validated_TargetSize)
		for i := range p.Styles {
			if p.Styles[i].Colour+"-"+p.Styles[i].Size == targetStyle {
				newTarget.StyleID = p.Styles[i].ID
			}
		}

		// style not found
		if newTarget.StyleID == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameters"})
			return
		}

		// check duplicate
		if _, err := dataHandler.GetTargetByStyleId(newTarget.StyleID); !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "target exists"})
			return
		}

		// save target
		err = dataHandler.AddTarget(&newTarget)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		cache.Delete(targets_cache_key) // delete cache
		ctx.Status(http.StatusCreated)
	}
}

// get all products under tracing
// params: page, size
func GetTargets(dataHandler db.Handler) func(*gin.Context) {
	return func(ctx *gin.Context) {

		page := ctx.GetInt(middleware.Validated_QueryPage)
		size := ctx.GetInt(middleware.Validated_QuerySize)

		var targets []db.TargetInfo
		if t, ok := cache.Get(targets_cache_key); ok {
			targets = t.([]db.TargetInfo)
		} else {
			targets = dataHandler.GetTargets()
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

func UpdateTarget(dataHandler db.Handler) func(*gin.Context) {
	return func(ctx *gin.Context) {

		id := ctx.GetInt(middleware.Validated_TargetId)
		pid := ctx.GetInt(middleware.Validated_ProductId)
		colour := ctx.GetString(middleware.Validated_TargetColour)
		size := ctx.GetString(middleware.Validated_TargetSize)

		// build and check target
		var wg sync.WaitGroup
		wg.Add(2)

		var target *db.Target
		var styles map[string]*db.Style
		var errTarget, errStyles error

		go func() {
			target, errTarget = dataHandler.GetTargetById(uint(id))
			wg.Done()
		}()

		go func() {
			styles, errStyles = dataHandler.GetStylesByProductId(uint(pid))
			wg.Done()
		}()

		wg.Wait()

		if errTarget != nil {
			if errors.Is(errTarget, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "target not found"})
				return
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": errTarget.Error()})
				return
			}
		}

		if errStyles != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": errStyles.Error()})
			return
		}

		if target.ProductID != uint(pid) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product code"})
			return
		}

		targetStyle := colour + "-" + size
		for i := range styles {
			if styles[i].Colour+"-"+styles[i].Size == targetStyle {
				if target.StyleID != styles[i].ID {
					target.StyleID = styles[i].ID
				} else {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": "same style"})
					return
				}
			}
		}

		// save
		dataHandler.UpdateTarget(target)
		cache.Delete(targets_cache_key) // delete cache
		ctx.Status(http.StatusNoContent)
	}
}

func DeleteTarget(dataHandler db.Handler) func(*gin.Context) {
	return func(ctx *gin.Context) {

		id := ctx.GetInt(middleware.Validated_TargetId)
		if err := dataHandler.DeleteTargetById(uint(id)); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
				return
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		cache.Delete(targets_cache_key) // delete cache
		ctx.Status(http.StatusNoContent)
	}
}
