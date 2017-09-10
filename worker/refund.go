package worker

import (
    "log"

    "github.com/btcsuite/btcutil"
    "github.com/stunndard/gopaycoin/bitcoin"
    _ "github.com/stunndard/gopaycoin/config"
    "github.com/stunndard/gopaycoin/model"
    "time"
)

func RefundWorker() {
    var refunds []model.Refund

    time.Sleep(time.Duration(time.Second) * 1)
    log.Println("REF: Started")

    for {
        time.Sleep(time.Duration(time.Second) * 5) //time.Duration(config.Cfg.CheckInterval))

        // get all expired payments
        var err error
        refunds, err = model.GetActiveRefunds()
        if err != nil {
            log.Println("REF: Error getting refunds", err)
            return
        }
        if len(refunds) > 0 {
            log.Println("REF: Processing", len(refunds), "refunds")
        }
        for i := range refunds {
            // convert address
            address, err := btcutil.DecodeAddress(refunds[i].Address, nil)
            if err != nil {
                log.Println("REF: Error decoding btc address", err)
                continue
            }
            // convert amount
            amount, err := btcutil.NewAmount(refunds[i].AmountBTC)
            if err != nil {
                log.Println("REF: Error converting amount", err)
                continue
            }

            fee, err := bitcoin.SetSmartFee()
            if err != nil {
                log.Println("REF: Cannot set smart fee", err)
                return
            }

            log.Println("REF: Smart fee set:", fee.ToBTC())

            // refund here
            log.Println("REF: Refunding:", refunds[i].AmountBTC, "to address:", refunds[i].Address)
            tx, err := bitcoin.SendToAddress(address, amount)
            if err != nil {
                log.Println("REF: Error sending refund", err)
                continue
            }

            // save refund as completed
            refunds[i].Tx = tx.String()
            refunds[i].Completed = true
            if err := refunds[i].Save(); err != nil {
                log.Println("REF: Error saving refund", err)
                continue
            }
        }
    }
}
