package routes

import (
	"html/template"
	"strconv"
	"time"

	"github.com/stunndard/gopaycoin/config"
	"github.com/stunndard/gopaycoin/model"
	"github.com/stunndard/gopaycoin/rates"
	"gopkg.in/kataras/iris.v6"
)

const (
	tpl     = "invoice.html"
	refresh = "<meta http-equiv=\"refresh\" content=\"30\">"
)

func ExtGetInvoice(ctx *iris.Context) {
	id, err := ctx.ParamInt("id")
	if err != nil {
		// bad invoice ref
		ctx.MustRender(
			tpl,
			iris.Map{
				"invoice_status": "archived",
				"archived":       "active",
			},
		)
		return
	}

	payment, err := model.GetPaymentByReference(id)
	if err != nil {
		// payment not found
		ctx.MustRender(
			tpl,
			iris.Map{
				"invoice_status": "archived",
				"archived":       "active",
			},
		)
		return
	}

	status, _, _ := payment.GetStatus()

	// if active or pending or partially paid
	if payment.IsActive() || payment.IsPending() {
		var dur time.Duration
		active := ""
		paidPartial := ""
		if payment.IsActive() {
			active = "active"
			dur = time.Duration(time.Minute * time.Duration(config.Cfg.WaitUnconfirmedMinutes))
		} else {
			paidPartial = "paid-partial"
			dur = time.Duration(time.Minute * time.Duration(config.Cfg.WaitConfirmedMinutes))
		}
		timedur := dur - time.Now().Sub(payment.CreatedAt)

		expiring := ""
		min := int(timedur.Seconds()) / 60
		if min < 5 {
			expiring = "expiring-soon"
		}
		sec := int(timedur.Seconds()) % 60
		s := ""
		if sec < 10 {
			s = "0"
		}
		left := strconv.Itoa(min) + ":" + s + strconv.Itoa(sec)

		ctx.MustRender(
			tpl,
			iris.Map{
				"invoice_status":   paidPartial,
				"status":           status,
				"time":             left,
				"expiring":         expiring,
				"item":             payment.Item,
				"merchant":         payment.Merchant,
				"payment":          active,
				"refresh":          template.HTML(refresh),
				"pay_amount":       rates.RoundBTC(payment.AmountBTC - payment.FeeBTC),
				"fee_amount":       payment.FeeBTC,
				"btc_rate":         rates.RoundBTC(rates.R.USD),
				"pay_total_amount": payment.AmountBTC,
				"pay_address":      payment.Address,
				"return_url":       payment.ReturnUrl,
			},
		)
		return
	}

	// invoice paid
	if payment.IsPaid() {
		ctx.MustRender(
			tpl,
			iris.Map{
				"invoice_status": "paid",
				"paid":           "active",
				"item":           payment.Item,
				"merchant":       payment.Merchant,
				"pay_amount":     rates.RoundBTC(payment.AmountBTC - payment.FeeBTC),
				"fee_amount":     payment.FeeBTC,
				//"btc_rate":    rates.RoundBTC(rates.R.USD),
				"pay_total_amount": payment.AmountBTC,
			},
		)
		return
	}

	// invoice paid and overpaid
	if payment.IsOverPaid() {
		refund, _ := model.GetRefundByPaymentID(payment.ID)
		ctx.MustRender(
			tpl,
			iris.Map{
				"invoice_status": "paid",
				//"paid":        "active",
				"refunded":      "active",
				"refund_status": "Payment and Refund Complete",
				"refund_label":  "Overpaid by",
				"item":          payment.Item,
				"merchant":      payment.Merchant,
				"pay_amount":    rates.RoundBTC(payment.AmountBTC - payment.FeeBTC),
				"fee_amount":    payment.FeeBTC,
				//"btc_rate":    rates.RoundBTC(rates.R.USD),
				"pay_total_amount": payment.AmountBTC,
				"refunded_amount":  refund.AmountBTC,
				"refund_address":   refund.Address,
			},
		)
		return
	}

	// invoice refunded
	if payment.IsRefunded() {
		refund, _ := model.GetRefundByPaymentID(payment.ID)
		ctx.MustRender(
			tpl,
			iris.Map{
				"invoice_status":  "paid archived",
				"refunded":        "active",
				"refund_status":   "Refund Complete",
				"refund_label":    "Amount Refunded",
				"refunded_amount": refund.AmountBTC,
				"refund_address":  refund.Address,
				"merchant":        payment.Merchant,
			},
		)
		return
	}

	// invoice is in unknown state, i.e. not yet picked by expiredworker
	// we show invoice without any counters
	if payment.IsUnknown() {
		ctx.MustRender(
			tpl,
			iris.Map{
				"invoice_status": "paid",
				"refresh":        template.HTML(refresh),
				"item":           payment.Item,
				"merchant":       payment.Merchant,
				"pay_amount":     rates.RoundBTC(payment.AmountBTC - payment.FeeBTC),
				"fee_amount":     payment.FeeBTC,
				//"btc_rate":    rates.RoundBTC(rates.R.USD),
				"pay_total_amount": payment.AmountBTC,
			},
		)
		return
	}

	// invoice expired
	ctx.MustRender(
		tpl,
		iris.Map{
			"invoice_status": "archived",
			"valid_time":     config.Cfg.WaitUnconfirmedMinutes,
			"merchant":       payment.Merchant,
			"invoice_id":     payment.Reference,
			"expired":        "active",
			"return_url":     payment.ReturnUrl,
		},
	)
}
