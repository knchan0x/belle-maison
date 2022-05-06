package db

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name        string
	ProductCode string
	Styles      []Style
}

type Style struct {
	gorm.Model
	ProductID      uint
	StyleCode      string
	Colour         string
	Size           string
	ImageUrl       string
	PriceHistories []Price
}

type Price struct {
	gorm.Model
	StyleID uint
	Price   uint
	Stock   uint
}

type Target struct {
	gorm.Model
	ProductCode string
	ProductID   uint
	StyleID     uint
	TargetPrice uint
}
