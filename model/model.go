package model

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/stunndard/gopaycoin/config"
)

func InitDB() {
	// prepare db connection
	log.Println("Connecting to db", config.Cfg.DBConnection)

	var err error
	db, err = gorm.Open("mysql", "mysql", config.Cfg.DBConnection)
	if err != nil {
		log.Fatalln("db: failed to connect database")
	}
	//defer db.Close()
	// Migrate the schema
	log.Println("Migrating the database...")
	if db.AutoMigrate(&Payment{},
					  &Withdraw{},
		              &Rates{},
		              &Refund{},
		              &Callback{}).Error != nil {
		log.Fatalln("db: failed to migrate database")
	}
}
