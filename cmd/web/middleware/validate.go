package middleware

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	maxInt = 1<<31 - 1

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
	ProductId
	TargetId
	TargetColour
	TargetSize
	TargetPrice
	QueryPageSize
)

const (
	Validated_ProductCode  = "Validated_ProductCode"
	Validated_ProductId    = "Validated_ProductId"
	Validated_TargetId     = "Validated_TargetId"
	Validated_TargetColour = "Validated_TargetColour"
	Validated_TargetSize   = "Validated_TargetSize"
	Validated_TargetPrice  = "Validated_TargetPrice"
	Validated_QueryPage    = "Validated_QueryPage"
	Validated_QuerySize    = "Validated_QuerySize"
)

// TODO: comment
// Validate processes handler after Validations completed
func Validate(t ValidateType) func(*gin.Context) {
	switch t {
	case ProductCode:
		return validateProductCode()
	case ProductId:
		return validateProductId()
	case TargetId:
		return validateTargetId()
	case TargetColour:
		return validateTargetColour()
	case TargetSize:
		return validateTargetSize()
	case TargetPrice:
		return validateTargetPrice()
	case QueryPageSize:
		return validateQueryPageSize()
	default:
		return byPass()
	}
}

func byPass() func(*gin.Context) {
	return func(ctx *gin.Context) {
		ctx.Next()
	}
}

func validateProductCode() func(*gin.Context) {
	return func(ctx *gin.Context) {
		code := ctx.Param("productCode")
		if !ValidateProductCode(code) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "invalid id"}) // 404 Not Found -> ID Not Found
			return
		}
		ctx.Set(Validated_ProductCode, code)
		ctx.Next()
	}
}

func validateTargetId() func(*gin.Context) {
	return func(ctx *gin.Context) {
		targetId := ctx.Param("targetId")
		id, err := strconv.ParseUint(targetId, 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "invalid target id"})
			return
		}
		ctx.Set(Validated_TargetId, id)
		ctx.Next()
	}
}

func validatePostForm(ctx *gin.Context, item, ctxKey string) {
	v, ok := ctx.GetPostForm(item)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("missing %s", item)})
		return
	}
	ctx.Set(ctxKey, v)
}

// parseInt parses context key stored from string to int.
// It must be used after the context key has been validated
// and saved into context.
func parseInt(ctx *gin.Context, ctxKey, errMsg string) {
	v := ctx.GetString(ctxKey)
	if v != "" {
		intV, err := strconv.Atoi(v)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errMsg})
			return
		}
		ctx.Set(ctxKey, intV)
	}
}

func validateProductId() func(*gin.Context) {
	return func(ctx *gin.Context) {
		validatePostForm(ctx, "productId", Validated_ProductId)
		parseInt(ctx, Validated_ProductId, "invalid product id")
		ctx.Next()
	}
}

func validateTargetColour() func(*gin.Context) {
	return func(ctx *gin.Context) {
		validatePostForm(ctx, "colour", Validated_TargetColour)
		ctx.Next()
	}
}

func validateTargetSize() func(*gin.Context) {
	return func(ctx *gin.Context) {
		validatePostForm(ctx, "size", Validated_TargetSize)
		ctx.Next()
	}
}

func validateTargetPrice() func(*gin.Context) {
	return func(ctx *gin.Context) {
		validatePostForm(ctx, "price", Validated_TargetPrice)
		parseInt(ctx, Validated_TargetPrice, "invalid target price")
		ctx.Next()
	}
}

func validateQueryPageSize() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p := ctx.Query("page")
		s := ctx.Query("size")

		page, size := 0, maxInt
		if p != "" || s != "" {
			page, errPage := strconv.Atoi(p)
			size, errSize := strconv.Atoi(s)

			if errPage != nil || errSize != nil || page <= 0 || size <= 0 {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid parameters"})
				return
			}
		}

		ctx.Set(Validated_QueryPage, page)
		ctx.Set(Validated_QuerySize, size)
		ctx.Next()
	}
}
