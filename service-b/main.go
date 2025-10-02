package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"service": "B",
			"message": "Hello from service B",
		})
	})

	e.Logger.Fatal(e.Start(":8080"))
}