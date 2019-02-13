package model

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Callback struct {
	gorm.Model
	Payment          uint
	Url              string
	Event            string
	LastResponseCode uint
	Attempts         uint
	Success          bool
	Failed           bool
}

func CreateCallback(callback *Callback, id uint) (*Callback, error) {
	err := db.FirstOrCreate(&callback, "(payment = ?) AND (event = ?)", id, callback.Event).Error
	if err != nil {
		log.Println("db: ERROR updating db", err)
		//ctx.JSON(iris.StatusOK, &iris.Map{"status": "db error"})
		//return err
	}
	return callback, err
}

func GetCallbackByPaymentID(id uint) (Callback, error) {
	var callback Callback
	if err := db.Where("payment = ?", id).Find(&callback).Error; err != nil {
		log.Println("db: ERROR querying db", err)
		return callback, err
	}
	return callback, nil
}

func (c *Callback) Save() error {
	if err := db.Save(c).Error; err != nil {
		log.Println("DB: ERROR saving Callback", err)
		return err
	}
	return nil
}

func CallbackCompleted(id uint, event string) bool {
	var c Callback
	if err := db.Where("(payment = ?) AND (event = ?)", id, event).Find(&c).Error; err != nil {
		//log.Println("db: ERROR querying db", err)
		return false
	}
	return (id == c.Payment) && (c.Failed || c.Success)
}

func GetNotCompleteCallbacks() ([]Callback, error) {
	var cs []Callback
	if err := db.Where("(success = ?) AND (failed = ?)",
		false, false).Find(&cs).Error; err != nil {
		log.Println("db: ERROR querying db", err)
		return nil, err
	}
	return cs, nil
}
