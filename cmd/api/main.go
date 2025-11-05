package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"smart/internal/config"
	httpHandlers "smart/internal/http"
	"smart/internal/database"
	"smart/internal/service"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if err := config.Load(); err != nil {
		log.Fatal().Err(err).Msg("config load failed")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatal().Err(err).Msg("db connect failed")
	}
	defer db.Close()

	svcs := service.New(db)
	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error { return c.SendString("ok") })

	httpHandlers.Register(app, svcs)

	addr := viper.GetString("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Info().Str("addr", addr).Msg("api listening")
	log.Fatal().Err(app.Listen(addr)).Msg("server exit")
}
