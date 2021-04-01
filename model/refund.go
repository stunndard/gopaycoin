package model

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Refund struct {
	gorm.Model
	Payment   uint
	Address   string
	AmountBTC float64
	Tx        string
	Completed bool
}

func CreateRefund(refund *Refund) error {
	if err := db.Create(&refund).Error; err != nil {
		log.Println("db: ERROR updating db", err)
		return err
	}
	return nil
}

func GetActiveRefunds() ([]Refund, error) {
	var refunds []Refund

	if err := db.Where("completed = ?", false).Find(&refunds).Error; err != nil {
		log.Println("db: ERROR querying db", err)
		return nil, err
	}
	return refunds, nil
}

func GetRefundByPaymentID(id uint) (Refund, error) {
	var refund Refund
	if err := db.Where("payment = ?", id).Find(&refund).Error; err != nil {
		log.Println("db: ERROR querying db", err)
		return refund, err
	}
	return refund, nil
}

func (r *Refund) Save() error {
	if err := db.Save(r).Error; err != nil {
		log.Println("DB: ERROR saving Refund", err)
		return err
	}
	return nil
}
