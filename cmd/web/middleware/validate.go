package middleware

import (
	"log"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

const (
	productIdLength    = 7
	pattern_product_id = "productId"
)

var patterns map[string]*regexp.Regexp

// TODO: refactor, pattern class?
func SetPatterns() {
	productIdPattern, err := regexp.Compile(`\d{7}`)
	if err != nil {
		log.Fatal("invalid regexp: productIdPattern")
	}
	patterns = map[string]*regexp.Regexp{pattern_product_id: productIdPattern}
}

func ValidateProductCode(code string) bool {
	if len(patterns) == 0 {
		log.Fatal("no pattern provided")
	}

	if code != "" && len(code) == productIdLength && patterns[pattern_product_id].MatchString(code) {
		return true
	}

	return false
}

type ValidateType int

const (
	ProductCode ValidateType = iota
)

const (
	Validated_Product_Code = "Validated_Product_Code"
)

// TODO: comment
func Validate(t ValidateType) func(*gin.Context) {
	switch t {
	case ProductCode:
		return func(ctx *gin.Context) {
			if !ValidateProductCode(ctx.Param("productCode")) {
				ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "invalid id"}) // 404 Not Found -> ID Not Found
			}
			ctx.Set(Validated_Product_Code, ctx.Param("productCode")) // TODO: use it
			ctx.Next()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.Next()
		}
	}
}
