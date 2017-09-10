package worker

import (
    "log"
    "github.com/stunndard/gopaycoin/model"
    "time"
    "github.com/stunndard/gopaycoin/bitcoin"
    "github.com/stunndard/gopaycoin/config"
    "github.com/btcsuite/btcutil"
)

var withdrawpending bool

func WithdrawWorker() {

    log.Println("WDR: Started")
    defer withdrawEnd()

    if withdrawpending {
        log.Println("WDR: another withdraw still active, exiting...")
        return
    }

    withdrawpending = true
    for {
        time.Sleep(time.Duration(time.Second) * time.Duration(3))

        refunds, err := model.GetActiveRefunds()
        if err != nil {
            log.Println("WDR: Cannot get active refunds, cannot withdraw now", err)
            return
        }
        if len(refunds) > 0 {
            log.Println("WDR: Found active refunds, postponing withdrawal")
            time.Sleep(time.Duration(time.Second) * 30)
            continue
        }

        // check and convert address
        address, err := btcutil.DecodeAddress(config.Cfg.ColdWallet, nil)
        if err != nil {
            log.Println("WDR: Error invalid address", config.Cfg.ColdWallet, err)
            return
        }

        balance, err := bitcoin.GetTotalBalance()
        if err != nil {
            log.Println("WDR: Cannot get total balance", err)
            return
        }

        // can withdraw now
        log.Println("WDR: Withdrawing, total balance:", balance.ToBTC())


        fee, err := bitcoin.SetSmartFee()
        if err != nil {
            log.Println("WDR: Cannot set smart fee", err)
            return
        }

        log.Println("WDR: Smart fee set:", fee)

        amount, _ := btcutil.NewAmount(0.5)

        // send
        tx, err := bitcoin.SendToAddress(address, amount)
        if err != nil {
            log.Println("WDR: Error withdrawing address", err)
            return
        }

        log.Println("WDR: Success. Address:", address.String(), "amount:", amount.ToBTC())

        if err := model.CreateWithdraw(&model.Withdraw{
            Address: config.Cfg.ColdWallet,
            Amount: amount.ToBTC(),
            Fee:    fee.ToBTC(),
            Tx:     tx.String(),
        }); err != nil {
            log.Println("db: ERROR updating db", err)
        }

        break
    }
}

func withdrawEnd() {
    withdrawpending = false
    log.Println("WDR: Ended")
}
