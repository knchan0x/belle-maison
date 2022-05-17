package target

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Target struct {
	gorm.Model
	ProductCode string
	ProductID   uint
	StyleID     uint
	TargetPrice uint
	dbClient    *gorm.DB `gorm:"-:all"`
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

func New(dbClient *gorm.DB, productCode string, productID uint, styleId uint, price uint) (*Target, error) {

	// check duplicate
	if t, err := getByStyleId(dbClient, styleId); !errors.Is(err, gorm.ErrRecordNotFound) {
		return t, fmt.Errorf("target exists")
	}

	// create target
	newTarget := Target{
		ProductCode: productCode,
		ProductID:   productID,
		StyleID:     styleId,
		TargetPrice: price,
		dbClient:    dbClient,
	}

	// save target
	err := newTarget.Save()
	if err != nil {
		return nil, err
	}

	return &newTarget, nil
}

// Get all targets' product info
func GetAll(dbClient *gorm.DB) (results []TargetInfo) {
	latestPrice := dbClient.Table("prices").
		Select("style_id, MAX(updated_at) as latest").
		Group("style_id")

	priceList := dbClient.Table("prices").
		Select("prices.style_id, prices.price, prices.stock, prices.updated_at").
		Joins("INNER JOIN (?) latestPrice ON prices.style_id = latestPrice.style_id AND prices.updated_at = latestPrice.latest", latestPrice).
		Group("prices.style_id, prices.price, prices.stock")

	styles := dbClient.Table("styles").
		Select("styles.id, styles.product_id, styles.colour, styles.size, styles.image_url, priceList.price, priceList.stock").
		Joins("LEFT JOIN (?) priceList ON styles.id = priceList.style_id", priceList)

	products := dbClient.Table("products").
		Select("products.id AS productId, products.name, styleList.id AS styleId, styleList.colour, styleList. `size`, styleList.image_url, styleList.price, styleList.stock").
		Joins("RIGHT JOIN (?) styleList ON styleList.product_id = products.id", styles)

	r := dbClient.Table("targets").
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
func GetList(dbClient *gorm.DB) []string {
	t := []Target{}
	r := dbClient.Select("productCode").Find(&t)
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

func getByProductCode(dbClient *gorm.DB, productCode string) (*Target, error) {
	t := Target{}
	r := dbClient.Where("product_code = ?", productCode).Limit(1).Find(&t)
	if r.Error == nil && r.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &t, r.Error
}

func getByStyleId(dbClient *gorm.DB, styleId uint) (*Target, error) {
	t := Target{}
	r := dbClient.Where("style_id = ?", styleId).Limit(1).Find(&t)
	if r.Error == nil && r.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &t, r.Error
}

func GetById(dbClient *gorm.DB, id uint) (*Target, error) {
	t := Target{}
	r := dbClient.Where("id = ?", id).Limit(1).Find(&t)
	if r.Error == nil && r.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &t, r.Error
}

// Save save instance to db
func (t *Target) Save() error {
	r := t.dbClient.Create(t)
	if r.Error != nil {
		return r.Error
	}

	return nil
}

// Update updates changes to db
func (t *Target) Update() {
	t.dbClient.Save(t)
}

// Delete deletes record from db
func (t *Target) Delete() {
	t.dbClient.Delete(t)
}
