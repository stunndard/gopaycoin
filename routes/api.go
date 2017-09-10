package routes

import (
	"log"
	"time"

	"gopkg.in/kataras/iris.v6"

	"github.com/btcsuite/btcutil"
	"github.com/stunndard/gopaycoin/bitcoin"
	"github.com/stunndard/gopaycoin/config"
)

var (
	Started time.Time
	Version string
)

// show health
func ExtGetHealth(ctx *iris.Context) {
	ctx.JSON(iris.StatusOK, &iris.Map{"health": "OK"})
}

// show version
func ExtGetVersion(ctx *iris.Context) {
	ctx.JSON(iris.StatusOK, &iris.Map{
		"success": true,
		"status":  200,
		"version": Version,
		"started": Started,
	})
}

// callback test listener
func ExtPostCallbackTest(ctx *iris.Context) {
	ctx.JSON(iris.StatusOK, &iris.Map{"callback": "OK"})
}

// fee
func ExtGetFee(ctx *iris.Context) {
	feeamount, err := bitcoin.EstimateFee(6)
	if err != nil {
		log.Println("FEE: Error getting RPC estimated fee", err)
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error getting estimate fee",
			"message": err.Error(),
		})
		return
	}
	log.Println("FEE: Estimated fee fetched:", feeamount)
	ctx.JSON(iris.StatusOK, &iris.Map{
		"status": "OK",
		"fee":    feeamount.ToBTC(),
	})
}

// new BTC address
func ExtPostNewAddress(ctx *iris.Context) {
	if ctx.PostValue("secret") != config.Cfg.Secret {
		log.Println("NEW: Invalid secret")
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "invalid secret"})
		return
	}
	log.Println("NEW: New address request.")

	address, err := bitcoin.CreateNewAddress()
	if err != nil {
		log.Println("RPC: ERROR can't create new BTC address", err)
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "RPC error create BTC address"})
		return
	}
	log.Println("NEW: New address created. Address:", address.String())
	ctx.JSON(iris.StatusOK, &iris.Map{
		"status":  "OK",
		"address": address.String(),
	})
}

func ExtPostBalance(ctx *iris.Context) {
	postAddress := ctx.PostValue("address")
	log.Println("BAL: New send request. Address:",
		postAddress)

	if ctx.PostValue("secret") != config.Cfg.Secret {
		log.Println("BAL: Invalid secret")
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "invalid secret"})
		return
	}

	// check and convert addresses
	address, err := btcutil.DecodeAddress(postAddress, nil)
	if err != nil {
		log.Println("BAL: Error invalid address")
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error in address",
			"message": err.Error(),
		})
		return
	}

	unconfBalance, confBalance, err := bitcoin.GetBalance(address)
	if err != nil {
		log.Println("BAL: Error getting balance")
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error getting balance",
			"message": err.Error(),
		})
		return
	}

	log.Println("BAL: Success getting balance. Address:",
		address.String(),
		"Unconfirmed:", unconfBalance.ToBTC(),
		"Confirmed:", confBalance.ToBTC())
	ctx.JSON(iris.StatusOK, &iris.Map{
		"status":      "OK",
		"unconfirmed": unconfBalance.ToBTC(),
		"confirmed":   confBalance.ToBTC(),
	})
}
