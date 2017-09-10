package model

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Rates struct {
	gorm.Model
	Currency string
	Value    float64
}

func CreateRates(rates *Rates) error {
	if err := db.Create(&rates).Error; err != nil {
		log.Println("db: ERROR updating db", err)
		return err
	}
	return nil
}
