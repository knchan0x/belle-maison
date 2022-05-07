package web

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/knchan0x/belle-maison/cmd/db"
	"github.com/knchan0x/belle-maison/internal/cache"
	"github.com/knchan0x/belle-maison/internal/crawler"
	"gorm.io/gorm"
)

var debugMode = false

var simpleAuthSuspended = false

func SetDebugMode(isDebugMode bool) {
	debugMode = isDebugMode
}

func ActivateSimpleAuth(isActivate bool) {
	simpleAuthSuspended = !isActivate
}

const (
	productIdLength = 7
	maxInt          = 1<<31 - 1

	pattern_product_id = "productId"
	targets_cache_key  = "target"
)

type User struct {
	Username string
	Password string
}

// API Gateway Handler, it contains *gin.Engine
type APIHandler struct {
	admin       *User
	web         *gin.Engine
	crawler     crawler.Crawler
	dataHandler db.Handler
	patterns    map[string]*regexp.Regexp
}

// NewHandler returns new APIHandler, it will initialize
// and attach a new *gin.Engine by using gin.Default()
func NewHandler(dataHandler db.Handler, admin *User) *APIHandler {
	productIdPattern, err := regexp.Compile(`\d{7}`)
	if err != nil {
		log.Fatal("invalid regexp: productIdPattern")
	}

	gin.SetMode(gin.ReleaseMode)

	c, err := crawler.NewCrawler()
	if err != nil {
		log.Fatalf("failed to initialize crawler: %v", err)
	}

	return &APIHandler{
		admin:       admin,
		web:         gin.Default(),
		crawler:     c,
		dataHandler: dataHandler,
		patterns:    map[string]*regexp.Regexp{pattern_product_id: productIdPattern},
	}
}

// Run adds path handlers to gin engine and runs it.
// It will automatically attach an allowCrossOrigin
// middleware during debug mode sets to true
func (h *APIHandler) Run(addr string) {
	if debugMode {
		h.web.Use(allowCrossOrigin)
	}

	// redirect / to /dashboard
	h.web.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/dashboard")
	})

	fileBasePath := ""
	_, err := os.Stat("./page/login.html")
	if err == nil {
		fileBasePath = "./page"
	} else {
		fileBasePath = "../page"
	}
	h.web.StaticFile("/login", fileBasePath+"/login.html")
	h.web.POST("/login", h.login)

	api := h.web.Group("/api", simpleAuthCheck(AuthMode_Unauthorized))
	api.GET("/product/:productCode", h.getProduct) // get product info
	api.POST("/target/:productCode", h.addTarget)  // POST content: colour, size
	api.DELETE("/target/:targetId", h.deleteTarget)
	api.PATCH("/target/:targetId", h.updateTarget)
	api.GET("/targets", h.getTargets) // get all products under tracing

	dashboard := h.web.Group("/dashboard", simpleAuthCheck(AuthMode_Redirect))
	dashboard.StaticFile("/", fileBasePath+"/index.html")

	h.web.Static("assets", fileBasePath+"/assets")

	log.Printf("Running web server on %s...", addr)
	if err := h.web.Run(addr); err != nil {
		log.Fatalf("Unable to start API Server: %v", err)
	}
}

// get product info
func (h *APIHandler) getProduct(ctx *gin.Context) {
	productCode := ctx.Param("productCode")
	if !h.verifyProductCode(productCode) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "invalid id"}) // 404 Not Found -> ID Not Found
		return
	}

	var r *crawler.Result
	if c, ok := cache.Get("crawler_result_" + productCode); ok {
		r = c.(*crawler.Result)
	} else {
		r = h.crawler.RetrieveProduct(productCode)
		cache.Add("crawler_result_"+productCode, r, time.Hour)
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

// add target
func (h *APIHandler) addTarget(ctx *gin.Context) {
	productCode := ctx.Param("productCode")
	if !h.verifyProductCode(productCode) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

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

	var r *crawler.Result
	go func() {
		if c, ok := cache.Get("crawler_result_" + productCode); ok {
			r = c.(*crawler.Result)
		} else {
			r = h.crawler.RetrieveProduct(productCode)
			cache.Add("crawler_result_"+productCode, r, time.Hour)
		}
		wg.Done()
	}()

	var p *db.Product
	var err error
	go func() {
		p, err = h.dataHandler.GetProductAndStylesByProductCode(productCode)
		wg.Done()
	}()

	wg.Wait()

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		p, err = h.dataHandler.CreateProduct(r)
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
	if _, err := h.dataHandler.GetTargetByStyleId(newTarget.StyleID); !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "target exists"})
		return
	}

	// save target
	err = h.dataHandler.AddTarget(&newTarget)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.Delete(targets_cache_key) // delete cache
	ctx.Status(http.StatusCreated)
}

// get all products under tracing
// params: page, size
func (h *APIHandler) getTargets(ctx *gin.Context) {
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
		targets = h.dataHandler.GetTargets()
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

func (h *APIHandler) updateTarget(ctx *gin.Context) {

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
		target, errTarget = h.dataHandler.GetTargetById(uint(id))
		wg.Done()
	}()

	go func() {
		styles, errStyles = h.dataHandler.GetStylesByProductId(uint(pid))
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
	h.dataHandler.UpdateTarget(target)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.Delete(targets_cache_key) // delete cache
	ctx.Status(http.StatusNoContent)
}

func (h *APIHandler) deleteTarget(ctx *gin.Context) {
	targetId := ctx.Param("targetId")
	id, err := strconv.ParseUint(targetId, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid target id"})
		return
	}

	if err := h.dataHandler.DeleteTargetById(uint(id)); err != nil {
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

func (h *APIHandler) verifyProductCode(code string) bool {
	if code != "" && len(code) == productIdLength && h.patterns[pattern_product_id].MatchString(code) {
		return true
	}

	return false
}

// login handler. Insecure. Please set SSL.
func (h *APIHandler) login(ctx *gin.Context) {

	user := ctx.PostForm("username")
	pw := ctx.PostForm("password")

	if user != h.admin.Username || pw != h.admin.Password {
		ctx.String(http.StatusUnauthorized, "username and/or password incorrect.")
	}

	// generate token
	md5 := md5.New()
	md5.Write([]byte(time.Now().Format("2006-01-02 15:04:05") + "belle-masion"))
	token := hex.EncodeToString(md5.Sum(nil))

	cache.Add("token", token, time.Hour)
	ctx.SetCookie("_cookie", token, 60*60, "/", "", true, true)
	ctx.Redirect(http.StatusFound, "/dashboard")
}

// allowCrossOrigin middleware handles CORS issues
// when debuging Vue app in http://localhost:3000
func allowCrossOrigin(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "http://localhost:3000")
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, PATCH, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type")
	ctx.Header("Access-Control-Max-Age", "86400")
	if ctx.Request.Method == http.MethodOptions {
		ctx.Status(http.StatusOK)
		return
	}
	ctx.Next()
}

type AuthMode string

const (
	AuthMode_Redirect     AuthMode = "Redirect"
	AuthMode_Unauthorized AuthMode = "Unauthorized"
)

// simpleAuthCheck returns gin middleware with mode specified.
// This middleware will check is the user permit to access
//
// - AuthMode_Redirect = redirect to login page
// - AuthMode_Unauthorized = return JSON with 401 unauthorized
func simpleAuthCheck(mode AuthMode) func(ctx *gin.Context) {
	if !simpleAuthSuspended {
		if mode == AuthMode_Redirect {
			return func(ctx *gin.Context) {
				t, err := ctx.Cookie("_cookie")
				token, ok := cache.Get("token")
				if err != nil || !ok || t != token.(string) {
					ctx.Redirect(http.StatusFound, "/login")
					return
				}
				ctx.Next()
			}
		}
		if mode == AuthMode_Unauthorized {
			return func(ctx *gin.Context) {
				t, err := ctx.Cookie("_cookie")
				token, ok := cache.Get("token")
				if err != nil || !ok || t != token.(string) {
					ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
				ctx.Next()
			}
		}
	}

	return func(ctx *gin.Context) {
		ctx.Next() // by pass
	}
}
