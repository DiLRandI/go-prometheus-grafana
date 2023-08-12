package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	// Echo instance
	e := echo.New()

	// Custom metric
	customCounter := prometheus.NewCounter( // create new counter metric. This is replacement for `prometheus.Metric` struct
		prometheus.CounterOpts{
			Name: "custom_requests_total",
			Help: "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
	)
	if err := prometheus.Register(customCounter); err != nil { // register your new counter metric with default metrics registry
		log.Fatal(err)
	}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(echoprometheus.NewMiddleware("myapp")) // adds middleware to gather metrics
	e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		AfterNext: func(c echo.Context, err error) {
			customCounter.Inc() // use our custom metric in middleware. after every request increment the counter
		},
	}))

	// Routes
	e.GET("/", hello)
	e.GET("/test", test)
	e.GET("/randomerr", randomErr)
	e.GET("/metrics", echoprometheus.NewHandler())

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

// Handlers
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func test(c echo.Context) error {
	//Add random delay to simulate a slow request
	time.Sleep(time.Duration(100+rand.Intn(2000)) * time.Millisecond)

	return c.String(http.StatusOK, "Test is ok!")
}

func randomErr(c echo.Context) error {
	time.Sleep(time.Duration(100+rand.Intn(900)) * time.Millisecond)

	shouldFail := rand.Intn(100)
	if shouldFail < 20 {
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	} else if shouldFail < 50 {
		return c.String(http.StatusBadGateway, "Bad Gateway")
	} else if shouldFail < 70 {
		return c.String(http.StatusServiceUnavailable, "Service Unavailable")
	}

	return c.String(http.StatusOK, "Random error?? no error!")
}
