package main

import (
	"fmt"
	"github.com/gin-contrib/static"
	_ "github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v9"
	"log"
	"reserve/reserve"
	"reserve/reserve/allocator"
	"reserve/reserve/concurrency"
	"time"
)

func main() {
	router := buildRouter()
	err := router.Run(":8080")
	if err != nil {
		log.Panic(err)
	}
}

func buildRouter() *gin.Engine {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterStructValidation(reserve.BodyStructValidation, reserve.Body{})
	}

	config := reserve.NewConfig()

	concurrencyService := concurrency.NewService(config.Concurrency)
	allocatorService := allocator.NewService(config.Allocator)

	reserveService := reserve.NewService(
		concurrencyService.CheckConcurrency,
		allocatorService.AllocateReserve,
		allocatorService.ListFromDB,
		allocatorService.ListFromRegistry,
	)

	router := gin.New()
	router.Use(loggerMiddleware)
	router.Use(concurrencyService.RegisterEntryMiddleware)
	router.Use(allocatorService.RegisterBucketExpirationMiddleware)
	router.Use(gin.Recovery())

	router.Use(static.Serve("/", static.LocalFile("./static", true)))
	router.POST("/api/users/:user_id/reserve", reserveService.HandleCreation)
	router.GET("/db/:user_id", reserveService.HandleDBRequest)
	router.GET("/registry/:user_id", reserveService.HandleRegistryRequest)

	return router
}

var loggerMiddleware = gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
	return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
		param.ClientIP,
		param.TimeStamp.Format(time.RFC1123),
		param.Method,
		param.Path,
		param.Request.Proto,
		param.StatusCode,
		param.Latency,
		param.Request.UserAgent(),
		param.ErrorMessage,
	)
})
