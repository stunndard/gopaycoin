package model

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Withdraw struct {
	gorm.Model
	Address string
	Amount  float64
	Fee     float64
	Tx      string
}

func CreateWithdraw(withdraw *Withdraw) error {
	err := db.Create(&withdraw).Error
	if err != nil {
		log.Println("db: ERROR updating db", err)
		//ctx.JSON(iris.StatusOK, &iris.Map{"status": "db error"})
		//return err
	}
	return err
}
