package worker

import (
	"log"
	"time"

	"github.com/btcsuite/btcutil"
	"github.com/stunndard/gopaycoin/bitcoin"
	"github.com/stunndard/gopaycoin/callback"
	"github.com/stunndard/gopaycoin/config"
	"github.com/stunndard/gopaycoin/model"
	"github.com/stunndard/gopaycoin/rates"
)

// payment processing PaymentWorker
func PaymentWorker() {

	time.Sleep(time.Duration(time.Second) * 1)
	log.Println("WRK: Started")

	var payments []model.Payment

	for {
		time.Sleep(time.Duration(time.Second) * time.Duration(config.Cfg.PaymentCheckInterval))
		log.Println("PAY: Started")

		// process payments
		// get all new and pending payments
		var err error
		if payments, err = model.GetActivePayments(); err != nil {
			continue
		}

		if len(payments) > 0 {
			log.Println("PAY: Processing", len(payments), "payments")
		}
		for i := range payments {
			log.Println("PAY:",
				"ID:", payments[i].ID, "Required  balance:", payments[i].AmountBTC,
				"AmountBTC, Address:", payments[i].Address, "Pending:", payments[i].Pending)
			// convert the address
			btcAddress, err := btcutil.DecodeAddress(payments[i].Address, nil)
			if err != nil {
				log.Println("PAY: ERROR decoding address", payments[i].Address)
				continue
			}

			// get the unconfirmed and confirmed received amounts
			unconfBalance, confBalance, err := bitcoin.GetReceived(btcAddress, config.Cfg.MinConfirmations)

			if !payments[i].Pending {
				log.Println("PAY:", "ID:", payments[i].ID, "Unconfirmed balance:", unconfBalance.String(),
					", Address:", btcAddress.String())

				if unconfBalance > 0 {
					// we have some unconfirmed payment
					// mark the payment as pending
					payments[i].Pending = true
					if err := payments[i].Save(); err != nil {
						//log.Println("db: ERROR updating db", err)
						continue
					}
					log.Println("PAY:", "ID:", payments[i].ID, "Pending successful.",
						"Reference:", payments[i].Reference,
						"AmountBTC:", payments[i].AmountBTC,
						"Address:", payments[i].Address)

					// add the callback
					if !model.CallbackCompleted(payments[i].ID, "pending") {
						callback.AddCallbackWorker(payments[i])
					}
				} else {
					continue
				}
			}

			// Get the address CONFIRMED balance
			log.Println("PAY:", "ID:", payments[i].ID, "Confirmed balance:", confBalance.String(),
				", Address:", btcAddress.String())

			// get the expected balance
			expectedBalance, err := btcutil.NewAmount(payments[i].AmountBTC)
			if err != nil {
				log.Println("PAY: ERROR converting to BTCAmount", payments[i].AmountBTC)
				continue
			}
			// balance arrived
			if confBalance >= expectedBalance {
				payments[i].Pending = false
				payments[i].Paid = true
				// if overpaid more than processing fee*2
				overpaid := confBalance - expectedBalance
				if overpaid >= (rates.FeeToBTC() * 2) {
					if (!payments[i].Refunded) && (payments[i].IsRefundAllowed()) {
						log.Println("PAY: Payment overpaid by:", overpaid.ToBTC(), "refunding..")
						// create refund by amount of overpaid
						if err := model.CreateRefund(&model.Refund{
							Payment:   payments[i].ID,
							Address:   payments[i].RefundAddress,
							AmountBTC: overpaid.ToBTC(),
						}); err != nil {
							log.Println("PAY: Error creating refund", err)
							continue
						}
						payments[i].Refunded = true
					}
				}
				if err := payments[i].Save(); err != nil {
					continue
				}

				log.Println("PAY:", "ID:", payments[i].ID, "Payment successfull.",
					"Reference:", payments[i].Reference,
					"AmountBTC:", payments[i].AmountBTC,
					"Address:", payments[i].Address)

				if !model.CallbackCompleted(payments[i].ID, "completed") {
					callback.AddCallbackWorker(payments[i])
				}
			}
		}

		// process callbacks
		// get all not delivered callbacks

		callbacks, err := model.GetNotCompleteCallbacks()
		if err != nil {
			continue
		}
		for i := range callbacks {
			//if !model.CallbackCompleted(callbacks[i].Payment) {
			payment, err := model.GetPaymentByID(callbacks[i].Payment)
			if err != nil {
				log.Println("PAY: Error db getting payments from callback", err)
				continue
			}
			callback.AddCallbackWorker(payment)
			//}
		}

		log.Println("PAY: Ended")
	}
}
