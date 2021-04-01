package model

import (
	"log"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pjebs/optimus-go"
	"github.com/stunndard/gopaycoin/config"
)

type Payment struct {
	gorm.Model
	Item          string
	Merchant      string
	Reference     string
	Address       string
	Amount        float64
	Currency      string
	AmountBTC     float64
	FeeBTC        float64
	Pending       bool
	Paid          bool
	Refunded      bool
	RefundAddress string
	Expired       bool
	Callback      string
	Custom        string
	ReturnUrl     string
}

// primes for optimus encryption
const Prime1 uint64 = 66441343
const Prime2 uint64 = 1360654207
const Prime3 uint64 = 146570057

// global objects
var db *gorm.DB

func CreatePayment(payment *Payment) error {
	if err := db.Create(&payment).Error; err != nil {
		log.Println("db: ERROR updating db", err)
		//ctx.JSON(iris.StatusOK, &iris.Map{"status": "db error"})
		return err
	}

	// generate the obfuscated ID
	o := optimus.New(Prime1, Prime2, Prime3)
	payment.Reference = strconv.Itoa(int(o.Encode(uint64(payment.ID))))

	// save the ref number
	if err := payment.Save(); err != nil {
		//log.Println("db: ERROR updating db", err)
		//ctx.JSON(iris.StatusOK, &iris.Map{"status": "db error"})
		return err
	}

	return nil
}

func GetPaymentByReference(id int) (Payment, error) {
	var payment Payment
	err := db.First(&payment, "reference = ?", id).Error
	//ctx.JSON(iris.StatusOK, &iris.Map{"status": "No such payment"})
	return payment, err
}

func GetPaymentByID(id uint) (Payment, error) {
	var payment Payment
	err := db.First(&payment, "id = ?", id).Error
	//ctx.JSON(iris.StatusOK, &iris.Map{"status": "No such payment"})
	return payment, err
}

func GetActivePayments() ([]Payment, error) {
	var payments []Payment

	// get all new and pending payments
	if err := db.Where("(created_at > ? AND paid + expired + refunded + pending = ?) OR "+
		"(created_at > ? AND pending = ? AND paid + expired + refunded = ?)",
		time.Now().Add(-time.Minute*time.Duration(config.Cfg.WaitUnconfirmedMinutes)),
		false,
		time.Now().Add(-time.Minute*time.Duration(config.Cfg.WaitConfirmedMinutes+config.Cfg.WaitConfirmedMinutes)),
		true,
		false).Find(&payments).Error; err != nil {

		log.Println("db: ERROR querying db", err)
		return nil, err
	}
	return payments, nil
}

func GetActiveCallbacks() ([]Payment, error) {
	var payments []Payment

	/*
		if err := db.Where("(pending = ? AND callback_pending = ?) OR (paid = ? AND callback_completed = ?)",
			true, false, true, false).Find(&payments).Error; err != nil {
			log.Println("db: ERROR querying db", err)
			return nil, err
		}
	*/

	if err := db.Where("(pending = ? AND callback_pending = ?) OR (paid = ? AND callback_completed = ?)",
		true, false, true, false).Find(&payments).Error; err != nil {
		log.Println("db: ERROR querying db", err)
		return nil, err
	}
	return payments, nil
}

func GetExpiredPayments() ([]Payment, error) {
	var payments []Payment

	if err := db.Where("(created_at < ? AND paid = ? AND refunded = ? AND expired = ?) OR "+
		"(created_at < ? AND pending = ? AND expired = ? AND refunded = ? AND paid = ?)",
		time.Now().Add(-time.Minute*time.Duration(config.Cfg.WaitUnconfirmedMinutes+config.Cfg.WaitConfirmedMinutes)),
		false, false, false,
		time.Now().Add(-time.Minute*time.Duration(config.Cfg.WaitUnconfirmedMinutes)),
		false, false, false, false).Find(&payments).Error; err != nil {
		log.Println("db: ERROR querying db", err)
		return nil, err
	}
	return payments, nil
}

func (p *Payment) Save() error {
	if err := db.Save(p).Error; err != nil {
		log.Println("DB: ERROR saving Payment", err)
		return err
	}
	return nil
}

func (p *Payment) IsActive() bool {
	if p.Paid || p.Pending {
		return false
	}
	return p.CreatedAt.Add(time.Minute * time.Duration(config.Cfg.WaitUnconfirmedMinutes)).After(time.Now())
}

func (p *Payment) IsPending() bool {
	if !p.Pending {
		return false
	}
	return p.UpdatedAt.Add(time.Minute * time.Duration(config.Cfg.WaitConfirmedMinutes)).After(time.Now())
}

func (p *Payment) IsPaid() bool {
	return p.Paid && !p.Refunded
}

func (p *Payment) IsOverPaid() bool {
	return p.Paid && p.Refunded
}

func (p *Payment) IsRefunded() bool {
	return p.Refunded && !p.Paid
}

func (p *Payment) IsUnknown() bool {
	return !p.Pending && !p.Paid && !p.Refunded && !p.Expired
}

func (p *Payment) IsRefundAllowed() bool {
	return (p.RefundAddress != "") && (config.Cfg.EnableRefunds)
}

func (p *Payment) GetStatus() (string, string, string) {
	status := ""
	if p.Paid {
		status = "Paid"
	} else if (p.CreatedAt.Add(time.Minute * time.Duration(config.Cfg.WaitUnconfirmedMinutes))).Before(time.Now()) &&
		!p.Pending {
		status = "Expired, never paid"
	} else if (p.UpdatedAt.Add(time.Minute * time.Duration(config.Cfg.WaitConfirmedMinutes))).Before(time.Now()) {
		status = "Expired, not enough confirmations"
	} else if p.Pending {
		status = "Awaiting more confirmations"
	} else {
		status = "Awaiting payment"
	}

	// check callback status
	callbackpending := "delivering"
	callbackcompleted := "delivering"
	callback, err := GetCallbackByPaymentID(p.ID)
	if err != nil {
		callbackpending = "error getting callback status"
		callbackcompleted = "error getting callback status"
	} else {
		if callback.Event == "pending" {
			if callback.Success {
				callbackpending = "delivered"
			} else if callback.Failed {
				callbackpending = "not delivered"
			}
		} else if callback.Event == "completed" {
			if callback.Success {
				callbackcompleted = "delivered"
			} else if callback.Failed {
				callbackcompleted = "not delivered"
			}

		}
	}
	return status, callbackpending, callbackcompleted
}
