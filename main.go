package main

import (
	"net/http"
	"time"

	"proxy-gateway/config"
	"proxy-gateway/middlewares"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	logger "cloudmt.co.kr/mateLogger"
)

var g errgroup.Group

func main() {
	if err := config.InitConfig(); err != nil {
		panic(err.Error())
	}

	conf := config.GetConfig()

	logger.SetupLog("logs", "proxy-gateway", false)
	logger.Start()
	logger.Custom(conf.PrintJson())

	route := makeRouter()

	for port, _ := range conf.Network_list {
		if port != "" && conf.Https_certfile != "" && conf.Https_keyfile != "" {
			https_server := &http.Server{
				Addr:         ":" + port,
				Handler:      route,
				ReadTimeout:  time.Duration(conf.ReadTimeout) * time.Second,
				WriteTimeout: time.Duration(conf.WriteTimeout) * time.Second,
			}

			g.Go(func() error {
				err := https_server.ListenAndServeTLS(conf.Https_certfile, conf.Https_keyfile)
				if err != nil && err != http.ErrServerClosed {
				}
				return err
			})
		}
	}

	if err := g.Wait(); err != nil {
		// error log
	}
}

func makeRouter() *gin.Engine {
	gin.DisableConsoleColor()

	route := gin.Default()
	route.Use(middlewares.AccessControlAllowOrigin())
	route.Use(middlewares.GinBodyLogMiddleware)

	route.NoRoute(middlewares.ReturnReverseProxy())
	return route
}
