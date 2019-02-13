package routes

import (
	"github.com/stunndard/gopaycoin/config"
	"gopkg.in/kataras/iris.v6"
)

func Headers(ctx *iris.Context) {
	ctx.SetHeader("Server", "kozler")
	ctx.Next()
}

func CheckSecret(ctx *iris.Context) {
	if ctx.PostValue("secret") != config.Cfg.Secret {
		ctx.JSON(iris.StatusOK, &iris.Map{"status": "invalid secret"})
		return
	}
	ctx.Next()
}
