package main

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/stunndard/gopaycoin/bitcoin"
	"github.com/stunndard/gopaycoin/config"
	"github.com/stunndard/gopaycoin/model"
	"github.com/stunndard/gopaycoin/routes"
	"github.com/stunndard/gopaycoin/worker"
	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
    "gopkg.in/kataras/iris.v6/adaptors/view"
	"github.com/stunndard/gopaycoin/rates"
	//"github.com/carlescere/scheduler"
	"github.com/carlescere/scheduler"
)

func StartWebApp() {
	app := iris.New(iris.Configuration{Gzip: true})
	app.Adapt(iris.DevLogger())
	app.Adapt(httprouter.New())
    app.Adapt(view.HTML("./templates", ".html").Reload(true))

	app.OnError(iris.StatusNotFound, func(ctx *iris.Context) {
		log.Println("NOT FOUND " + ctx.Path())
		ctx.JSON(iris.StatusNotFound, &iris.Map{"route": "NOT FOUND"})
	})

	// headers
	app.UseFunc(routes.Headers)

	// register routers
	app.Get("/health", routes.ExtGetHealth)
	app.Get("/", routes.ExtGetVersion)
	app.Post("/pay", routes.CheckSecret, routes.ExtCreatePayment)
	app.Get("/status/:id", routes.ExtGetPaymentStatus)
	app.Post("/callbacktest", routes.ExtPostCallbackTest)
	app.Post("/test", routes.ExtPostCallbackTest)
	app.Post("/payout", routes.ExtPostPayout)
	app.Get("/fee", routes.ExtGetFee)
	app.Post("/newaddress", routes.ExtPostNewAddress)
	app.Post("/send", routes.ExtPostSend)
	app.Post("/balance", routes.ExtPostBalance)
	app.Get("/invoice/:id", routes.ExtGetInvoice)

    // static invoice assets
    //app.StaticWeb("/assets", "./assets")
	h := app.StaticHandler("/assets", "./assets", false, false)
	app.Get("/assets/*path", h)

	// start the web app
	app.Listen(config.Cfg.Listen)
}

func main() {

	if err := config.LoadConfig(); err != nil {
		log.Println(err)
		return
	}

	// app start time
	routes.Started = time.Now()

	// app version
	b, err := ioutil.ReadFile("/go/src/app/version.txt")
	if err != nil {
		log.Println("cannot read version.txt file", err)
		routes.Version = "0.0.0-local debug"
	} else {
		routes.Version = string(b)
	}

	model.InitDB()

	bitcoin.InitBTCRPC()
	//defer btcrpc.Shutdown()

	// start the payments PaymentWorker
	go worker.PaymentWorker()

	// start the exchange rates worker
	go rates.Worker()

	// start the expired worker
	go worker.ExpiredWorker()

	// start refunds worker
	go worker.RefundWorker()

	// schedule the withdraw worker
	scheduler.Every().Day().At("00:58").Run(worker.WithdrawWorker)

	StartWebApp()
}
