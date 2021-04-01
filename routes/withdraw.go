package routes

import (
	"log"
	"strconv"

	"github.com/btcsuite/btcutil"
	"github.com/stunndard/gopaycoin/bitcoin"
	"github.com/stunndard/gopaycoin/config"
	"github.com/stunndard/gopaycoin/model"
	"gopkg.in/kataras/iris.v6"
)

// payout
func ExtPostPayout(ctx *iris.Context) {

	postAddress := ctx.PostValue("address")
	postAmount := ctx.PostValue("amount")
	postFee := ctx.PostValue("fee")

	log.Println("PYO: New payout request. Address:",
		postAddress,
		"Amount:", postAmount,
		"Fee:", postFee)

	if ctx.PostValue("secret") != config.Cfg.Secret {
		log.Println("PYO: Invalid secret")
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "invalid secret"})
		return
	}

	// check and convert fee
	fee, err := strconv.ParseFloat(postFee, 64)
	if err != nil {
		log.Println("PYO: Incorrect fee", err)
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "fee incorrect"})
		return
	}
	feeamount, err := btcutil.NewAmount(fee)
	if err != nil {
		log.Println("PYO: Error converting fee value", err)
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error in converting fee",
			"message": err.Error(),
		})
		return
	}

	// check and convert amount
	amount, err := strconv.ParseFloat(postAmount, 64)
	if err != nil {
		log.Println("PYO: Incorrect amount", err)
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "amount incorrect"})
		return
	}
	payoutamount, err := btcutil.NewAmount(amount)
	if err != nil {
		log.Println("PYO: Error converting amount value")
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error in converting amount",
			"message": err.Error(),
		})
		return
	}

	// check and convert address
	address, err := btcutil.DecodeAddress(postAddress, nil)
	if err != nil {
		log.Println("PYO: Error invalid address")
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error in address",
			"message": err.Error(),
		})
		return
	}

	// check balance if we have enough
	balance, err := bitcoin.GetTotalBalance()
	if err != nil {
		log.Println("PYO: Error getting RPC balance", err)
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error getting RPC balance",
			"message": err.Error(),
		})
		return
	}

	if balance < payoutamount {
		log.Println("PYO: Error requested amount is bigger than actual balance. Balance:", balance, "Requested:", payoutamount)
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status": "error not enough balance",
		})
		return
	}

	if fee == 0 {
		// set the estimated smart fee
		feeamount, err = bitcoin.EstimateFee(6)
		if err != nil {
			log.Println("PYO: Error getting RPC estimated fee", err)
			ctx.JSON(iris.StatusOK, &iris.Map{
				"status":  "error getting estimate fee",
				"message": err.Error(),
			})
			return
		}
		log.Println("PYO: Estimated fee fetched:", feeamount)
	}

	// set the fee given by user
	if err := bitcoin.SetTxFee(feeamount); err != nil {
		log.Println("PYO: Error setting RPC fee", err)
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error setting fee",
			"message": err.Error(),
		})
		return
	}

	tx, err := bitcoin.SendToAddress(address, payoutamount)
	if err != nil {
		log.Println("PYO: Error sending RPC to address", err)
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error sending",
			"message": err.Error(),
		})
		return
	}

	payout := &model.Withdraw{
		Address: address.String(),
		Amount:  payoutamount.ToBTC(),
		Fee:     feeamount.ToBTC(),
		Tx:      tx.String(),
	}

	if err := model.CreateWithdraw(payout); err != nil {
		log.Println("db: ERROR updating db", err)
	}

	log.Println("PYO: Withdraw sent successfully. TX:", tx)
	ctx.JSON(iris.StatusOK, &iris.Map{
		"status":      "OK",
		"transaction": tx.String(),
		"fee":         feeamount.ToBTC(),
	})

}

// send raw tx
func postCalcFee(ctx *iris.Context) {
	postFromAddress := ctx.PostValue("fromaddress")
	postToAddress := ctx.PostValue("toaddress")

	postAmount := ctx.PostValue("amount")
	postFeePerKB := ctx.PostValue("feeperkb")

	log.Println("CLC: New send calculate request. FromAddress:",
		postFromAddress,
		"ToAdress:", postToAddress,
		"Amount:", postAmount,
		"Fee:", postFeePerKB)

	if ctx.PostValue("secret") != config.Cfg.Secret {
		log.Println("CLC: Invalid secret")
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "invalid secret"})
		return
	}

	// check and convert feePerKB
	feePerKB, err := strconv.ParseFloat(postFeePerKB, 64)
	if err != nil {
		log.Println("CLC: Incorrect feePerKB", err)
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "feePerKB incorrect"})
		return
	}
	feePerKBamount, err := btcutil.NewAmount(feePerKB)
	if err != nil {
		log.Println("CLC: Error converting feePerKB value", err)
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error in converting feePerKB",
			"message": err.Error(),
		})
		return
	}

	// check and convert amount
	samount, err := strconv.ParseFloat(postAmount, 64)
	if err != nil {
		log.Println("CLC: Incorrect amount", err)
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "amount incorrect"})
		return
	}
	sendamount, err := btcutil.NewAmount(samount)
	if err != nil {
		log.Println("CLC: Error converting amount value")
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error in converting amount",
			"message": err.Error(),
		})
		return
	}

	// check and convert addresses
	fromaddress, err := btcutil.DecodeAddress(postFromAddress, nil)
	if err != nil {
		log.Println("CLC: Error invalid address")
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error in fromaddress",
			"message": err.Error(),
		})
		return
	}
	toaddress, err := btcutil.DecodeAddress(postToAddress, nil)
	if err != nil {
		log.Println("CLC: Error invalid address")
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error in toaddress",
			"message": err.Error(),
		})
		return
	}

	// prepare and get the signed transaction
	sigTx, changeamount, err := bitcoin.CreateSignTransaction(fromaddress, toaddress, sendamount, 0)
	if err != nil {
		log.Println("CLC: Error creating/signing transaction", err)
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error creating/signing transaction",
			"message": err.Error(),
		})
		return
	}

	// calculate the fee for this transaction
	feefloat := float64(sigTx.SerializeSize()) / 1000.0 * float64(feePerKBamount)
	feeamount := btcutil.Amount(feefloat)
	changeamount = changeamount - feeamount

	log.Println("CLC: Calc send successful. From:",
		fromaddress.String(),
		"To:", toaddress.String(),
		"Amount:", sendamount.ToBTC(),
		//"Change:", changeamount.ToBTC(),
		"Fee per KB:", feePerKBamount.ToBTC(),
		"Fee:", feeamount.ToBTC(),
		"Tx Size:", sigTx.SerializeSize())

	ctx.JSON(iris.StatusOK, &iris.Map{
		"status": "OK",
		"from":   fromaddress.String(),
		"to":     toaddress.String(),
		"amount": sendamount.ToBTC(),
		//"change": changeamount.ToBTC(),
		"txsize":   sigTx.SerializeSize(),
		"fee":      feeamount.ToBTC(),
		"feeperkb": feePerKBamount.ToBTC(),
	})
}

// send raw tx
func ExtPostSend(ctx *iris.Context) {
	postFromAddress := ctx.PostValue("fromaddress")
	postToAddress := ctx.PostValue("toaddress")

	postAmount := ctx.PostValue("amount")
	postFee := ctx.PostValue("fee")

	log.Println("SND: New send request. FromAddress:",
		postFromAddress,
		"ToAdress:", postToAddress,
		"Amount:", postAmount,
		"Fee:", postFee)

	if ctx.PostValue("secret") != config.Cfg.Secret {
		log.Println("SND: Invalid secret")
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "invalid secret"})
		return
	}

	// check and convert fee
	fee, err := strconv.ParseFloat(postFee, 64)
	if err != nil {
		log.Println("SND: Incorrect fee", err)
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "fee incorrect"})
		return
	}
	feeamount, err := btcutil.NewAmount(fee)
	if err != nil {
		log.Println("SND: Error converting fee value", err)
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error in converting fee",
			"message": err.Error(),
		})
		return
	}

	// check and convert amount
	samount, err := strconv.ParseFloat(postAmount, 64)
	if err != nil {
		log.Println("SND: Incorrect amount", err)
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "amount incorrect"})
		return
	}
	sendamount, err := btcutil.NewAmount(samount)
	if err != nil {
		log.Println("SND: Error converting amount value")
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error in converting amount",
			"message": err.Error(),
		})
		return
	}

	// check and convert addresses
	fromaddress, err := btcutil.DecodeAddress(postFromAddress, nil)
	if err != nil {
		log.Println("SND: Error invalid address")
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error in fromaddress",
			"message": err.Error(),
		})
		return
	}
	toaddress, err := btcutil.DecodeAddress(postToAddress, nil)
	if err != nil {
		log.Println("SND: Error invalid address")
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error in toaddress",
			"message": err.Error(),
		})
		return
	}

	// prepare and get the signed transaction
	sigTx, changeamount, err := bitcoin.CreateSignTransaction(fromaddress, toaddress, sendamount, feeamount)
	if err != nil {
		log.Println("SND: Error creating/signing transaction", err)
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error creating/signing transaction",
			"message": err.Error(),
			"change":  changeamount.ToBTC(),
		})
		return
	}

	txid, err := bitcoin.SendRawTx(sigTx, false)
	if err != nil {
		log.Println("SND: Error sending transaction")
		ctx.JSON(iris.StatusOK, &iris.Map{
			"status":  "error sending transaction",
			"message": err.Error(),
		})
		return
	}

	log.Println("SND: Send successful. From:",
		fromaddress.String(),
		"To:", toaddress.String(),
		"Amount:", sendamount.ToBTC(),
		"Change:", changeamount.ToBTC(),
		"Fee:", feeamount.ToBTC(),
		"Tx Size:", sigTx.SerializeSize(),
		"Tx:", txid.String(),
	)

	ctx.JSON(iris.StatusOK, &iris.Map{
		"status": "OK",
		"from":   fromaddress.String(),
		"to":     toaddress.String(),
		"amount": sendamount.ToBTC(),
		"change": changeamount.ToBTC(),
		"txsize": sigTx.SerializeSize(),
		"fee":    feeamount.ToBTC(),
		"tx":     txid.String(),
	})
}
