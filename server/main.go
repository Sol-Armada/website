package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
	attndnc "github.com/sol-armada/sol-bot/attendance"
	"github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/stores"
	tkns "github.com/sol-armada/sol-bot/tokens"
	"github.com/spf13/viper"
)

var version = "local"
var hash = "local"

var ctx = context.Background()

func main() {
	viper.AddConfigPath(".")
	viper.AddConfigPath("./server")
	viper.AddConfigPath("/etc/solarmada")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		slog.Error("failed to read config", "error", err)
		return
	}

	opts := log.Options{
		Fields:          []interface{}{"version", version, "hash", hash},
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.DateTime,
	}
	if viper.GetBool("LOG.DEBUG") {
		opts.Level = log.DebugLevel
	}

	handler := log.NewWithOptions(os.Stdout, opts)

	logger := slog.New(handler)

	logger.Debug("starting server")

	slog.SetDefault(logger)

	echoLogger := slog.Default()

	// set slog output to file if not local
	if !viper.GetBool("LOG.CLI") {
		opts.Formatter = log.JSONFormatter
		opts.TimeFormat = time.RFC3339

		path := strings.TrimSuffix(viper.GetString("LOG.FILE"), "/")
		f, err := os.OpenFile(path+"/server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			slog.Error("failed to open log file", "error", err, "path", path+"/server.log")
			return
		}
		slog.SetDefault(slog.New(log.NewWithOptions(f, opts)))
		logger = slog.Default()

		f, err = os.OpenFile(path+"/server-http.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			slog.Error("failed to open log file", "error", err, "path", path+"/server-http.log")
			return
		}
		echoLogger = slog.New(log.NewWithOptions(f, opts))
	}

	slog.Info("connecting to mongo and setting up stores")

	host := viper.GetString("MONGO.HOST")
	port := viper.GetInt("MONGO.PORT")
	replicaSetName := viper.GetString("MONGO.REPLICA_SET_NAME")
	database := viper.GetString("MONGO.DATABASE")

	_, err := stores.New(ctx, host, port, "", "", database, replicaSetName)
	if err != nil {
		slog.Error("failed to connect to mongo", "error", err)
		return
	}

	if err := members.Setup(); err != nil {
		slog.Error("failed to setup members", "error", err)
		return
	}

	if err := attndnc.Setup(); err != nil {
		slog.Error("failed to setup attendance", "error", err)
		return
	}

	if err := tkns.Setup(); err != nil {
		slog.Error("failed to setup tokens", "error", err)
		return
	}

	logger.Info("starting websocket server")

	hub := newHub(ctx)
	go hub.run()

	go watchForTokens(ctx, hub)
	go watchForAttendance(ctx, hub)

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

func encrypt(in string) (string, error) {
	key := make([]byte, 32)
	copy(key, []byte(version+hash))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, aes.BlockSize+len(in))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], []byte(in))

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func decrypt(in string) (string, error) {
	key := make([]byte, 32)
	copy(key, []byte(version+hash))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	cipherText, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return "", err
	}

	iv := cipherText[:aes.BlockSize]
	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(cipherText[aes.BlockSize:], cipherText[aes.BlockSize:])

	return string(cipherText[aes.BlockSize:]), nil
}
