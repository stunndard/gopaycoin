package config

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/go-ini/ini"
)

type cfgStruct struct {
	WaitUnconfirmedMinutes int
	WaitConfirmedMinutes   int
	MinConfirmations       int
	PaymentCheckInterval   int
	FeeBTC                 float64
	DBConnection           string
	BTCHost                string
	BTCHostUseTLS          bool
	BTCUser                string
	BTCPass                string
	Secret                 string
	ColdWallet             string
	EnableRefunds          bool
	EnableWithdrawals      bool
	Listen                 string
}

var Cfg cfgStruct

func getEnvValue(env string) string {
	if os.Getenv(env) == "" {
		log.Fatalln(env, "environment value is not set")
	}
	return os.Getenv(env)
}

func loadEnvConfig() {
	// check if config params are set
	Cfg.WaitUnconfirmedMinutes, _ = strconv.Atoi(getEnvValue("APP_WAIT_UNCONFIRMED"))
	Cfg.WaitConfirmedMinutes, _ = strconv.Atoi(getEnvValue("APP_WAIT_CONFIRMED"))
	Cfg.MinConfirmations, _ = strconv.Atoi(getEnvValue("APP_MIN_CONFIRMATIONS"))
	Cfg.PaymentCheckInterval, _ = strconv.Atoi(getEnvValue("APP_CHECK_INTERVAL"))
	Cfg.FeeBTC, _ = strconv.ParseFloat(getEnvValue("APP_FEE_BTC"), 64)
	Cfg.DBConnection = getEnvValue("APP_DB_CONNECTION")
	Cfg.BTCHost = getEnvValue("APP_BTC_HOST")
	Cfg.BTCHostUseTLS = getEnvValue("APP_BTC_HOST_USE_TLS") == "1"
	Cfg.BTCUser = getEnvValue("APP_BTC_USER")
	Cfg.BTCPass = getEnvValue("APP_BTC_PASS")
	Cfg.Secret = getEnvValue("APP_SECRET")
	Cfg.ColdWallet = getEnvValue("APP_COLD_WALLET")
	Cfg.EnableRefunds = getEnvValue("APP_ENABLE_REFUNDS") == "1"
	Cfg.EnableWithdrawals = getEnvValue("APP_ENABLE_WITHDRAWALS") == "1"

	Cfg.Listen = getEnvValue("APP_LISTEN")
}

func loadFileConfig(inifile string) error {
	iniObj, err := ini.Load(inifile)
	if err != nil {
		return err
	}

	section, err := iniObj.GetSection("")
	if err != nil {
		return err
	}

	Cfg.WaitUnconfirmedMinutes = section.Key("WAIT_UNCONFIRMED").MustInt(15)
	Cfg.WaitConfirmedMinutes = section.Key("WAIT_CONFIRMED").MustInt(180)
	Cfg.MinConfirmations = section.Key("MIN_CONFIRMATIONS").MustInt(3)
	Cfg.PaymentCheckInterval = section.Key("CHECK_INTERVAL").MustInt(30)
	Cfg.DBConnection = section.Key("DB_CONNECTION").MustString("btc:btc@tcp(127.0.0.1:3306)/gopaycoin?charset=utf8&parseTime=True&loc=Local")
	Cfg.BTCHost = section.Key("BTC_HOST").MustString("127.0.0.1:9482")
	Cfg.BTCHostUseTLS = section.Key("BTC_HOST_USE_TLS").MustBool(true)
	Cfg.BTCUser = section.Key("BTC_USER").MustString("bitcoinrpc")
	Cfg.BTCPass = section.Key("BTC_PASS").MustString("hackme")
	Cfg.Secret = section.Key("SECRET").MustString("hackme")
	if Cfg.Secret == "hackme" {
		log.Fatalln("You MUST change the default SECRET config value!")
	}
	Cfg.Listen = section.Key("LISTEN").MustString(":8080")
	Cfg.FeeBTC = section.Key("FEE_BTC").MustFloat64(0.000428)
	Cfg.ColdWallet = section.Key("COLD_WALLET").MustString("")
	Cfg.EnableRefunds = section.Key("ENABLE_REFUNDS").MustBool(true)
	Cfg.EnableWithdrawals = section.Key("ENABLE_WITHDRAWALS").MustBool(true)

	return nil
}

func LoadConfig() error {
	inifile := flag.String("config", "/etc/gopaycoin.ini", "config file")
	envconfig := flag.Bool("env", false, "use ENV variables as config")
	flag.Parse()

	if !*envconfig {
		if err := loadFileConfig(*inifile); err != nil {
			return err
		}
	} else {
		loadEnvConfig()
	}

	return nil
}
