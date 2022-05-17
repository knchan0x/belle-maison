package product

import (
	"errors"

	"github.com/knchan0x/belle-maison/internal/scraper"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name        string
	ProductCode string
	Styles      []Style
	dbClient    *gorm.DB `gorm:"-:all"`
}

type Style struct {
	gorm.Model
	ProductID      uint
	StyleCode      string
	Colour         string
	Size           string
	ImageUrl       string
	PriceHistories []Price
	dbClient       *gorm.DB `gorm:"-:all"`
}

type Price struct {
	gorm.Model
	StyleID uint
	Price   uint
	Stock   uint
}

var EMPTY_PRODUCT = errors.New("no product info for creation")

// New accepts *scraper.Result and returns *Product
func New(dbClient *gorm.DB, result *scraper.Result) (*Product, error) {
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
		dbClient:    dbClient,
	}

	err := p.Save()
	if err != nil {
		return nil, err
	}

	return p, nil
}

// return product only, corresponsing styles, price and stock will not included
func GetProductByCode(dbClient *gorm.DB, productCode string) (*Product, error) {
	p := Product{}
	r := dbClient.Where("product_code = ?", productCode).Limit(1).Find(&p)
	if r.Error == nil && r.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &p, r.Error
}

// return product only, corresponsing styles, price and stock will not included
func GetProductById(dbClient *gorm.DB, pid uint) (*Product, error) {
	p := Product{}
	r := dbClient.Where("product_id = ?", pid).Limit(1).Find(&p)
	if r.Error == nil && r.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &p, r.Error
}

// Get styles
// key = colour-size
func (p *Product) AllStyles() (map[string]*Style, error) {
	styles := []Style{}
	r := p.dbClient.Where("product_id = ?", p.ID).Find(&styles)

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

func (p *Product) Style(colour, size string) (*Style, error) {
	s := Style{}
	r := p.dbClient.Where("product_id = ? AND colour = ? AND size = ?", p.ID, colour, size).Limit(1).Find(&s)
	if r.Error == nil && r.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &s, r.Error
}

func (s *Style) PriceHistory() ([]Price, error) {
	prices := []Price{}
	r := s.dbClient.Where("style_id = ?", s.ID).Find(&prices)

	if r.Error == nil && r.RowsAffected > 0 {
		return prices, nil
	}
	return nil, r.Error
}

func (p *Product) Delete() error {
	styles, err := p.AllStyles()
	if err != nil {
		return err
	}

	return p.dbClient.Transaction(func(tx *gorm.DB) error {
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

func (p *Product) Update(result *scraper.Result) error {

	// product has been removed, set price == 0 and stock == 0 for all styles
	if result.Err == scraper.PRODUCT_NOT_FOUND {
		for idx := range p.Styles {
			p.Styles[idx].PriceHistories = append(p.Styles[idx].PriceHistories, Price{Price: 0, Stock: 0})
		}
		p.dbClient.Save(p)
		return nil
	}

	// update product name
	if p.Name != result.Product.Name {
		p.Name = result.Product.Name
	}

	// get all styles of current product
	storedStyles, err := p.AllStyles()
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
				ProductID: p.ID,
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

	p.dbClient.Create(&batchStyle)
	p.dbClient.Create(&batchPrice)
	return nil
}

func (p *Product) Save() error {
	r := p.dbClient.Create(p)
	if r.Error != nil {
		return r.Error
	}

	return nil
}
