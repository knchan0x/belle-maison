package route

import (
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/knchan0x/belle-maison/internal/cache"
	"github.com/knchan0x/belle-maison/internal/db"
	"github.com/knchan0x/belle-maison/internal/scraper"
	"gorm.io/gorm"
)

const (
	maxInt            = 1<<31 - 1
	targets_cache_key = "target"
)

// add target
func AddTarget(s scraper.Scraper, dataHandler db.Handler) func(*gin.Context) {
	return func(ctx *gin.Context) {
		productCode := ctx.Param("productCode")

		colour, colourOK := ctx.GetPostForm("colour")
		size, sizeOK := ctx.GetPostForm("size")
		price, priceOK := ctx.GetPostForm("price")
		if !colourOK || !sizeOK || !priceOK {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "empty target product info"})
			return
		}

		// get most updated product info
		var wg sync.WaitGroup
		wg.Add(2)

		var r *scraper.Result
		go func() {
			if c, ok := cache.Get("scraper_result_" + productCode); ok {
				r = c.(*scraper.Result)
			} else {
				r = s.ScrapingProduct(productCode)
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

		intPrice, err := strconv.Atoi(price)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid target price"})
			return
		}

		// create target
		newTarget := db.Target{
			ProductCode: productCode,
			ProductID:   p.ID,
			TargetPrice: uint(intPrice),
		}

		targetStyle := colour + "-" + size
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
		p := ctx.Query("page")
		s := ctx.Query("size")

		page, size := 0, maxInt
		if p != "" || s != "" {
			page, errPage := strconv.Atoi(p)
			size, errSize := strconv.Atoi(s)

			if errPage != nil || errSize != nil || page <= 0 || size <= 0 {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameters"})
				return
			}
		}

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

		// validate
		targetId := ctx.Param("targetId")
		id, err := strconv.ParseUint(targetId, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "invalid target id"})
			return
		}

		productId, OK := ctx.GetPostForm("productId")
		if !OK {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "product id missing"})
			return
		}

		pid, err := strconv.ParseUint(productId, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
			return
		}

		colour, colourOK := ctx.GetPostForm("colour")
		size, sizeOK := ctx.GetPostForm("size")
		if !colourOK || !sizeOK {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "empty target info"})
			return
		}

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
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "target not found"})
				return
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		if errStyles != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		cache.Delete(targets_cache_key) // delete cache
		ctx.Status(http.StatusNoContent)
	}
}

func DeleteTarget(dataHandler db.Handler) func(*gin.Context) {
	return func(ctx *gin.Context) {
		targetId := ctx.Param("targetId")
		id, err := strconv.ParseUint(targetId, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid target id"})
			return
		}

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
