package routes

import (
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/btcsuite/btcutil"
	"github.com/stunndard/gopaycoin/bitcoin"
	"github.com/stunndard/gopaycoin/config"
	"github.com/stunndard/gopaycoin/model"
	"github.com/stunndard/gopaycoin/rates"
	"gopkg.in/kataras/iris.v6"
)

// new payment
func ExtCreatePayment(ctx *iris.Context) {
	// check if callback url is correct
	_, err := url.ParseRequestURI(ctx.PostValue("callback"))
	if err != nil {
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "invalid callback url"})
		return
	}

	// check the amount
	amount, err := strconv.ParseFloat(ctx.PostValue("amount"), 64)
	if err != nil {
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "amount error"})
		return
	}

	// check the currency name
	if ctx.PostValue("currency") != "USD" {
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "wrong currency name"})
		return
	}

	// check the item name
	item := ctx.PostValue("item")
	if item == "" {
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "item is empty"})
		return
	}

	// check if the refund address is valid
	refundAddress := ctx.PostValue("refundaddress")
	if refundAddress != "" {
		if _, err := btcutil.DecodeAddress(refundAddress, nil); err != nil {
			ctx.JSON(iris.StatusOK, &iris.Map{"status": "refundaddress is invalid"})
			return
		}
	}

	// our rates not updated
	// we cannot accept payments
	if rates.R.USD == 0 {
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "rates error, try again later"})
		return
	}

	// get new AmountBTC address
	address, err := bitcoin.CreateNewAddress()
	if err != nil {
		log.Println("RPC: ERROR can't create new AmountBTC address", err)
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "RPC error create AmountBTC address"})
		return
	}

	// get our AmountBTC price
	btcfloat := amount * (1 / rates.R.USD)

	fee := rates.RoundBTC(config.Cfg.FeeBTC)

	// add fee
	btcfloat = btcfloat + fee

	// round btc amount to 6 digits
	btc := rates.RoundBTC(btcfloat)

	// create payment
	payment := &model.Payment{
		Item:          item,
		Merchant:      ctx.PostValue("merchant"),
		Address:       address.String(),
		RefundAddress: refundAddress,
		Amount:        amount,
		Currency:      ctx.PostValue("currency"),
		AmountBTC:     btc,
		FeeBTC:        fee,
		Callback:      ctx.PostValue("callback"),
		Custom:        ctx.PostValue("custom"),
		ReturnUrl:     ctx.PostValue("returnurl"),
	}

	if err := model.CreatePayment(payment); err != nil {
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "db error"})
	}

	log.Println("REQ: new payment create successful", "Ref:", payment.Reference, "AmountBTC:", payment.AmountBTC,
		"Address:", payment.Address, "Callback:", payment.Callback)
	ctx.JSON(iris.StatusOK, &iris.Map{"status": "OK",
		"item":          payment.Item,
		"amount":        payment.Amount,
		"currency":      payment.Currency,
		"amountbtc":     payment.AmountBTC,
		"address":       payment.Address,
		"refundaddress": payment.RefundAddress,
		"reference":     payment.Reference,
		"paybefore":     time.Now().Add(time.Minute * time.Duration(config.Cfg.WaitUnconfirmedMinutes)),
		"statusurl":     "http://" + ctx.ServerHost() + "/status/" + payment.Reference,
		"invoice":       "http://" + ctx.ServerHost() + "/invoice/" + payment.Reference,
	})
}

// payment status
func ExtGetPaymentStatus(ctx *iris.Context) {
	id, err := ctx.ParamInt("id")
	if err != nil {
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "param error"})
		return
	}

	payment, err := model.GetPaymentByReference(id)
	if err != nil {
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "No such payment"})
		return
	}

	status, callbackpending, callbackcompleted := payment.GetStatus()

	ctx.JSON(iris.StatusOK, &iris.Map{
		"status":             status,
		"amount":             payment.AmountBTC,
		"address":            payment.Address,
		"paybefore":          payment.CreatedAt.Add(time.Minute * time.Duration(config.Cfg.WaitUnconfirmedMinutes)),
		"confirmbefore":      payment.UpdatedAt.Add(time.Minute * time.Duration(config.Cfg.WaitConfirmedMinutes)),
		"callback_pending":   callbackpending,
		"callback_completed": callbackcompleted,
	})

}
