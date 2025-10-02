package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		// サービスBのURLを環境変数から取得
		serviceBURL := os.Getenv("SERVICE_B_URL")
		if serviceBURL == "" {
			serviceBURL = "http://localhost:8081"
		}

		// サービスBにリクエスト
		resp, err := http.Get(serviceBURL + "/")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("failed to call service B: %v", err),
			})
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("failed to read response: %v", err),
			})
		}

		// サービスBの結果を加工して返す
		return c.JSON(http.StatusOK, map[string]string{
			"service":       "A",
			"message":       "Response from service A",
			"service_b_msg": string(body),
		})
	})

	e.Logger.Fatal(e.Start(":8080"))
}