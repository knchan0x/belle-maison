package db

import (
	"errors"

	"github.com/knchan0x/belle-maison/internal/scraper"
	"gorm.io/gorm"
)

type Handler interface {
	// Targets
	GetTargets() []TargetInfo
	GetTargetList() []string

	// Target
	GetTargetByProductCode(string) (*Target, error)
	GetTargetByStyleId(uint) (*Target, error)
	GetTargetById(uint) (*Target, error)
	AddTarget(*Target) error
	UpdateTarget(*Target)
	DeleteTargetById(uint) error

	// Product
	GetProductByProductCode(string) (*Product, error)
	GetProductAndStylesByProductCode(string) (*Product, error)
	CreateProduct(*scraper.Result) (*Product, error)
	UpdateProduct(*scraper.Result) error

	// Styles
	GetStylesByProductId(uint) (map[string]*Style, error)
}

var handler Handler

// GetHandler returns the data handler
func GetHandler(dbClient *gorm.DB) Handler {
	if handler != nil {
		return handler
	}

	handler = &dataHandler{
		dbClient: dbClient,
	}

	return handler
}

type dataHandler struct {
	dbClient *gorm.DB
}

type TargetInfo struct {
	ID          uint
	ProductCode string
	Name        string
	Colour      string
	Size        string
	ImageUrl    string
	TargetPrice uint
	Price       uint
	Stock       uint
}

// Get all targets' product info
func (h *dataHandler) GetTargets() (results []TargetInfo) {
	latestPrice := h.dbClient.Table("prices").
		Select("style_id, MAX(updated_at) as latest").
		Group("style_id")

	priceList := h.dbClient.Table("prices").
		Select("prices.style_id, prices.price, prices.stock, prices.updated_at").
		Joins("INNER JOIN (?) latestPrice ON prices.style_id = latestPrice.style_id AND prices.updated_at = latestPrice.latest", latestPrice).
		Group("prices.style_id, prices.price, prices.stock")

	styles := h.dbClient.Table("styles").
		Select("styles.id, styles.product_id, styles.colour, styles.size, styles.image_url, priceList.price, priceList.stock").
		Joins("LEFT JOIN (?) priceList ON styles.id = priceList.style_id", priceList)

	products := h.dbClient.Table("products").
		Select("products.id AS productId, products.name, styleList.id AS styleId, styleList.colour, styleList. `size`, styleList.image_url, styleList.price, styleList.stock").
		Joins("RIGHT JOIN (?) styleList ON styleList.product_id = products.id", styles)

	r := h.dbClient.Table("targets").
		Select("targets.id, targets.product_code, targets.target_price, productList. `name`, productList.colour, productList. `size`, productList.image_url, productList.price, productList.stock").
		Joins("LEFT JOIN (?) productList ON productList.styleId = targets.style_id", products).
		Where("targets.deleted_at IS NULL").
		Order("targets.id DESC").
		Scan(&results)

	if r.Error == nil && r.RowsAffected > 0 {
		return results
	}
	return nil
}

// Get all targets' product code
func (h *dataHandler) GetTargetList() []string {
	t := []Target{}
	r := h.dbClient.Select("productCode").Find(&t)
	if r.Error == nil && r.RowsAffected > 0 {
		return targetToList(t)
	}
	return nil
}

func targetToList(targets []Target) (list []string) {
	if targets == nil || len(targets) <= 0 {
		return
	}

	for _, target := range targets {
		list = append(list, target.ProductCode)
	}
	return
}

func (h *dataHandler) GetTargetByProductCode(productCode string) (*Target, error) {
	t := Target{}
	r := h.dbClient.Where("product_code = ?", productCode).Limit(1).Find(&t)
	if r.Error == nil && r.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &t, r.Error
}

func (h *dataHandler) GetTargetByStyleId(styleId uint) (*Target, error) {
	t := Target{}
	r := h.dbClient.Where("style_id = ?", styleId).Limit(1).Find(&t)
	if r.Error == nil && r.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &t, r.Error
}

func (h *dataHandler) GetTargetById(id uint) (*Target, error) {
	t := Target{}
	r := h.dbClient.Where("id = ?", id).Limit(1).Find(&t)
	if r.Error == nil && r.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &t, r.Error
}

func (h *dataHandler) AddTarget(t *Target) error {
	r := h.dbClient.Create(t)
	if r.Error != nil {
		return r.Error
	}

	return nil
}

func (h *dataHandler) UpdateTarget(t *Target) {
	h.dbClient.Save(t)
}

func (h *dataHandler) DeleteTargetById(id uint) error {
	t, err := h.GetTargetById(id)
	if err != nil {
		return err
	}

	h.dbClient.Delete(t)
	return nil
}

// return product only, corresponsing styles, price and stock will not included
func (h *dataHandler) GetProductByProductCode(productCode string) (*Product, error) {
	p := Product{}
	r := h.dbClient.Where("product_code = ?", productCode).Limit(1).Find(&p)
	if r.Error == nil && r.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &p, r.Error
}

// return product and corresponsing styles; price and stock will not included
func (h *dataHandler) GetProductAndStylesByProductCode(productCode string) (*Product, error) {
	p := Product{}
	r := h.dbClient.Where("product_code = ?", productCode).Limit(1).Find(&p)
	if r.Error != nil {
		return nil, r.Error
	}
	if r.RowsAffected != 1 {
		return nil, gorm.ErrRecordNotFound
	}
	s := []Style{}
	if r := h.dbClient.Where("product_id = ?", p.ID).Find(&s); r.Error != nil {
		return nil, r.Error
	}
	p.Styles = s
	return &p, nil
}

var EMPTY_PRODUCT = errors.New("no product info for creation")

// CreateProduct accepts *scraper.Result and returns *Product
func (h *dataHandler) CreateProduct(result *scraper.Result) (*Product, error) {
	if result.Product == nil {
		return nil, EMPTY_PRODUCT
	}

	styles := make([]Style, len(result.Product.Styles))
	for i, style := range result.Product.Styles {
		styles[i] = Style{
			StyleCode: style.StyleCode,
			ImageUrl:  style.ImageUrl,
			Colour:    style.Colour,
			Size:      style.Size,
			PriceHistories: []Price{
				{
					Price: style.Price,
					Stock: style.Stock,
				},
			},
		}
	}

	p := &Product{
		Name:        result.Product.Name,
		ProductCode: result.ProductCode,
		Styles:      styles,
	}

	err := h.addProduct(p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (h *dataHandler) addProduct(p *Product) error {
	r := h.dbClient.Create(p)
	if r.Error != nil {
		return r.Error
	}

	return nil
}

func (h *dataHandler) UpdateProduct(result *scraper.Result) error {
	product, err := h.GetProductByProductCode(result.ProductCode)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		product = &Product{
			Name:        result.Product.Name,
			ProductCode: result.ProductCode,
		}

		err = h.addProduct(product)
		if err != nil {
			return err
		} else {
			return nil
		}
	}

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// product has been removed, set price == 0 and stock == 0 for all styles
	if result.Err == scraper.PRODUCT_NOT_FOUND {
		for idx := range product.Styles {
			product.Styles[idx].PriceHistories = append(product.Styles[idx].PriceHistories, Price{Price: 0, Stock: 0})
		}
		h.dbClient.Save(product)
		return nil
	}

	// update product name
	if product.Name != result.Product.Name {
		product.Name = result.Product.Name
	}

	// get all styles of current product
	storedStyles, err := h.GetStylesByProductId(product.ID)
	if err != nil {
		return err
	}

	// add price history
	batchPrice := []Price{}
	batchStyle := []Style{}
	for _, style := range result.Product.Styles {
		key := style.Colour + "-" + style.Size
		if dbStyle, ok := storedStyles[key]; ok {
			newPrice := Price{
				StyleID: dbStyle.ID,
				Price:   style.Price,
				Stock:   style.Stock,
			}
			batchPrice = append(batchPrice, newPrice)
		} else {
			// create new style
			newStyle := Style{
				StyleCode: style.StyleCode,
				ImageUrl:  style.ImageUrl,
				ProductID: product.ID,
				Colour:    style.Colour,
				Size:      style.Size,
				PriceHistories: []Price{{
					Price: style.Price,
					Stock: style.Stock,
				}},
			}
			batchStyle = append(batchStyle, newStyle)
		}
	}

	h.addStyles(batchStyle)
	h.addPrices(batchPrice)
	return nil
}

func (h *dataHandler) deleteProduct(p *Product) error {
	styles, err := h.GetStylesByProductId(p.ID)
	if err != nil {
		return err
	}

	return h.dbClient.Transaction(func(tx *gorm.DB) error {
		for _, style := range styles {

			if err := tx.Delete(&Price{}, "style_id = ?", style.ID).Error; err != nil {
				return err
			}
			if err := tx.Delete(style).Error; err != nil {
				return err
			}

		}

		if err := tx.Delete(p).Error; err != nil {
			return err
		}

		return nil
	})
}

// Get styles
// key = colour-size
func (h *dataHandler) GetStylesByProductId(productId uint) (map[string]*Style, error) {
	styles := []Style{}
	r := h.dbClient.Where("product_id = ?", productId).Find(&styles)

	if r.Error == nil && r.RowsAffected > 0 {
		// create mapping
		styleMap := make(map[string]*Style)
		for idx := range styles {
			key := styles[idx].Colour + "-" + styles[idx].Size
			styleMap[key] = &styles[idx]
		}
		return styleMap, nil
	}
	return nil, r.Error
}

func (h *dataHandler) addStyles(s []Style) {
	h.dbClient.Create(&s)
}

// func (h *dataHandler) getPrices(styleId uint) ([]Price, error) {
// 	prices := []Price{}
// 	r := h.dbClient.Where("style_id = ?", styleId).Find(&prices)

// 	if r.Error == nil && r.RowsAffected > 0 {
// 		return prices, nil
// 	}
// 	return nil, r.Error
// }

func (h *dataHandler) addPrices(p []Price) {
	h.dbClient.Create(&p)
}
