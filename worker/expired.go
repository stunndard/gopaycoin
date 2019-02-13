package worker

import (
	"log"
	"time"

	"github.com/btcsuite/btcutil"
	"github.com/stunndard/gopaycoin/bitcoin"
	"github.com/stunndard/gopaycoin/config"
	"github.com/stunndard/gopaycoin/model"
)

func ExpiredWorker() {
	var payments []model.Payment

	time.Sleep(time.Duration(time.Second) * 1)
	log.Println("EXP: Started")

	for {
		time.Sleep(time.Duration(time.Second) * time.Duration(config.Cfg.PaymentCheckInterval))

		// get all expired payments
		var err error
		payments, err = model.GetExpiredPayments()
		if err != nil {
			log.Println("EXP: Error getting expired payments", err)
			continue
		}

		for i := range payments {
			// get address
			address, err := btcutil.DecodeAddress(payments[i].Address, nil)
			if err != nil {
				log.Println("EXP: Error decoding address", err)
				continue
			}

			// had this address ever received anything?
			unconfBalance, confBalance, err := bitcoin.GetReceived(address, config.Cfg.MinConfirmations)
			if err != nil {
				log.Println("EXP: Error getting received by address", err)
				continue
			}
			if unconfBalance != confBalance {
				log.Println("EXP: Expired payment not fully confirmed. Cannot refund", unconfBalance, confBalance)
				continue
			}

			if (confBalance > 0) && (payments[i].IsRefundAllowed()) {
				// we need to refund this payment
				if err := model.CreateRefund(&model.Refund{
					Payment:   payments[i].ID,
					Address:   payments[i].RefundAddress,
					AmountBTC: confBalance.ToBTC(),
				}); err != nil {
					log.Println("EXP: Error creating refund", err)
					continue
				}
				payments[i].Refunded = true
			} else {
				payments[i].Expired = true
			}
			log.Println("EXP: Payment expired:", payments[i].ID)
			payments[i].Pending = false
			payments[i].Paid = false
			payments[i].Save()
		}
	}
}
