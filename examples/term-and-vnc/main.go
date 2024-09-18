package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"term-and-vnc/impl"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	log.Logger = log.
		With().
		Caller().
		Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})
	godotenv.Load()

	impl.Init()

	log.Info().Msg("starting server")
	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	})
	r.GET("/term", impl.Term)
	r.GET("/vnc", impl.Vnc)
	r.GET("/vnc-ticket", impl.VncTicket)

	r.RunTLS(":8523", "server.crt", "server.key")
}
