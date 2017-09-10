package rates

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/stunndard/gopaycoin/config"
	"github.com/stunndard/gopaycoin/model"
	"github.com/btcsuite/btcutil"
)

type exchangeRates struct {
	USD         float64
	LastUpdated time.Time
}

var R exchangeRates

func RoundBTC(amount float64) float64 {
	n := fmt.Sprintf("%.6f", amount)
	rounded, _ := strconv.ParseFloat(n, 64)
	return rounded
}

func FeeToBTC() btcutil.Amount {
	// convert amount
	amount, _ := btcutil.NewAmount(config.Cfg.FeeBTC)
	return amount
}

func Worker() {

	errors := 0
	firsttime := true
	sleepfor := time.Duration(0)

	time.Sleep(time.Duration(time.Second) * 1)
	log.Println("RAT: Started")
	for {

		if errors == 0 {
			sleepfor = time.Duration(time.Minute) * time.Duration(config.Cfg.WaitUnconfirmedMinutes)
		} else {
			sleepfor = time.Duration(time.Second) * 15
		}

		if errors > 15 && R.USD == 0 {
			break
		}

		if !firsttime {
			log.Println("RAT: Sleeping for", sleepfor)
			time.Sleep(sleepfor)
		}

		if R.LastUpdated.Add(time.Duration(time.Minute) * time.Duration(config.Cfg.WaitUnconfirmedMinutes)).Before(time.Now()) {
			log.Println("RAT: Too old rates value, update is required")
		}

		log.Println("RAT: Updating rates")

		request := gorequest.New()
		resp, body, errs := request.Get("https://blockchain.info/ticker").
			Timeout(60 * time.Second).
			End()

		if errs != nil {
			//return nil, errs
			// err request
			log.Println("RAT: Error updating rates", errs)
			errors++
			continue
		}

		if resp.StatusCode != 200 {
			// error
			log.Println("RAT: Error, http code:", resp.StatusCode)
			errors++
			continue
		}

		var res map[string]interface{}
		err := json.Unmarshal([]byte(body), &res)
		if err != nil {
			log.Println("RAT: Error unmarshal json", err)
			errors++
			continue
		}

		usd, ok := res["USD"].(map[string]interface{})["last"].(float64)

		if !ok {
			log.Println("RAT: Error converting rates values")
			errors++
			continue
		}

		R.LastUpdated = time.Now()
		R.USD = usd

		if err := model.CreateRates(&model.Rates{Currency: "USD", Value: usd}); err != nil {
			log.Println("RAT: Error saving to DB")
		}

		errors = 0
		firsttime = false
		log.Println("RAT: Success updating rates. USD:", usd)
	}
}
