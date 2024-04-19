package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
	"github.com/spf13/viper"
)

func main() {
	viper.AddConfigPath(".")
	viper.AddConfigPath("./server")
	viper.AddConfigPath("/etc/solarmada")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		slog.Error("failed to read config", "error", err)
		return
	}

	opts := &slog.HandlerOptions{
		AddSource: true,
	}
	if viper.GetString("LOG.LEVEL") == "DEBUG" {
		opts.Level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, opts)))

	echoLogger := slog.Default()

	// set slog output to file if not local
	if !viper.GetBool("LOCAL") {
		path := strings.TrimSuffix(viper.GetString("LOG.FILE"), "/")
		f, err := os.OpenFile(path+"/server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			slog.Error("failed to open log file", "error", err, "path", path+"/server.log")
			return
		}
		slog.SetDefault(slog.New(slog.NewJSONHandler(f, opts)))

		f, err = os.OpenFile(path+"/server-http.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			slog.Error("failed to open log file", "error", err, "path", path+"/server-http.log")
			return
		}
		echoLogger = slog.New(slog.NewJSONHandler(f, opts))
	}

	hub := newHub()
	go hub.run()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(slogecho.New(echoLogger))
	e.Use(middleware.Recover())

	// websocket
	e.GET("/ws", func(c echo.Context) error {
		serveWs(hub, c.Response(), c.Request())
		return nil
	})
	e.Logger.Fatal(e.Start(":" + viper.GetString("PORT")))
}
