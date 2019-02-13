package bitcoin

import (
	"errors"
	"log"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcrpcclient"
	"github.com/btcsuite/btcutil"
	"github.com/stunndard/gopaycoin/config"
)

var (
	btcrpc *btcrpcclient.Client
)

func SendRawTx(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error) {
	return btcrpc.SendRawTransaction(tx, allowHighFees)
}

// sendtoaddress ALWAYS subsctracts the fee from the amount being send
func SendToAddress(address btcutil.Address, amount btcutil.Amount) (*chainhash.Hash, error) {
	return btcrpc.SendToAddressComment(address, amount, "", "", true)
}

func SetTxFee(amount btcutil.Amount) error {
	return btcrpc.SetTxFee(amount)
}

func EstimateFee(numBlocks int64) (btcutil.Amount, error) {
	return btcrpc.EstimateFee(numBlocks)
}

func GetTotalBalance() (btcutil.Amount, error) {
	return btcrpc.GetBalanceMinConf("", config.Cfg.MinConfirmations)
}

func CreateNewAddress() (btcutil.Address, error) {
	return btcrpc.GetNewAddress("")
}

func SetSmartFee() (btcutil.Amount, error) {
	// get the estimated smart fee
	fee, err := btcrpc.EstimateSmartFee(20)
	if err != nil {
		return 0, err
	}
	if fee.FeeRate < 0 {
		return 0, errors.New("Estimate fee returned negative value")
	}

	feeamount, err := btcutil.NewAmount(fee.FeeRate)
	if err != nil {
		return 0, err
	}

	// set the smart fee
	if err := btcrpc.SetTxFee(feeamount); err != nil {
		return 0, err
	}
	return feeamount, nil
}

func GetReceived(address btcutil.Address, minconf int) (btcutil.Amount, btcutil.Amount, error) {
	unconf, err := btcrpc.GetReceivedByAddressMinConf(address, 0)
	if err != nil {
		return 0, 0, err
	}
	conf, err := btcrpc.GetReceivedByAddressMinConf(address, minconf)
	if err != nil {
		return 0, 0, err
	}
	return unconf, conf, err
}

func GetUnspents(address btcutil.Address, minConf int) ([]btcjson.ListUnspentResult, error) {
	unspents, err := btcrpc.ListUnspentMinMaxAddresses(minConf, 999999999, []btcutil.Address{address})
	if err != nil {
		return nil, err
	}
	return unspents, nil
}

func GetBalance(address btcutil.Address) (btcutil.Amount, btcutil.Amount, error) {
	// get all unspents
	unspents, err := GetUnspents(address, 0)
	if err != nil {
		return 0, 0, err
	}

	unconfBalance, confBalance := btcutil.Amount(0), btcutil.Amount(0)
	for i := range unspents {
		if !unspents[i].Spendable {
			continue
		}
		amount, _ := btcutil.NewAmount(unspents[i].Amount)
		if unspents[i].Confirmations < int64(config.Cfg.MinConfirmations) {
			unconfBalance += amount
		} else {
			confBalance += amount
		}
	}
	//unconfBalance = unconfBalance + confBalance
	return unconfBalance, confBalance, nil
}

func CreateSignTransaction(fromaddress, toaddress btcutil.Address, sendamount, feeamount btcutil.Amount) (
	*wire.MsgTx, btcutil.Amount, error) {
	// check the balance first if we have enough
	_, balamount, err := GetBalance(fromaddress)
	if err != nil {
		return nil, 0, err
	}

	log.Println(fromaddress, sendamount, feeamount, balamount)

	if balamount < sendamount {
		return nil, 0, errors.New("confirmed balance is less then requested balance")
	}

	unspents, err := GetUnspents(fromaddress, config.Cfg.MinConfirmations)
	if err != nil {
		return nil, 0, err
	}

	// create transaction inputs for the transaction
	var tis []btcjson.TransactionInput
	spendamount, _ := btcutil.NewAmount(0.0)
	for i := range unspents {
		if spendamount < sendamount+feeamount {
			ti := btcjson.TransactionInput{unspents[i].TxID, unspents[i].Vout}
			tis = append(tis, ti)
			amount, _ := btcutil.NewAmount(unspents[i].Amount)
			spendamount = spendamount + amount
		}
	}

	// calculate the change
	changeamount := spendamount - sendamount - feeamount
	if changeamount < 0 && feeamount > 0 {
		return nil, changeamount, errors.New("Not enough balance to send with the fee specified.")
	}

	/*
	   // if this change is dust we don't send it back
	   if changeamount > 0 && changeamount < 100+1 {
	       feeamount = feeamount + changeamount
	       changeamount = 0
	   }*/

	// we send the requested amount and send the change back to us
	sendto := map[btcutil.Address]btcutil.Amount{
		toaddress: sendamount,
	}
	if changeamount > 0 || feeamount == 0 {
		sendto[fromaddress] = changeamount
	}

	// in tis we have all the transaction inputs to spend now
	MsgTx, err := btcrpc.CreateRawTransaction(tis, sendto, nil)
	if err != nil {
		return nil, 0, err
	}

	// unsigned transaction length
	//log.Println(MsgTx.SerializeSize())

	SigTx, ok, err := btcrpc.SignRawTransaction(MsgTx)
	if err != nil {
		return nil, 0, err
	}
	if !ok {
		return nil, 0, errors.New("not all tx inputs can be signed")
	}

	return SigTx, changeamount, nil
}

func InitBTCRPC() {
	// prepare bitcoind RPC connection
	// Connect to remote bitcoin core RPC server using HTTP POST mode.
	connCfg := &btcrpcclient.ConnConfig{
		Host:         config.Cfg.BTCHost,
		User:         config.Cfg.BTCUser,
		Pass:         config.Cfg.BTCPass,
		HTTPPostMode: true,                      // Bitcoin core only supports HTTP POST mode
		DisableTLS:   !config.Cfg.BTCHostUseTLS, // Bitcoin core can use TLS if it's behind a TLS proxy like nginx
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	var err error
	btcrpc, err = btcrpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
}
