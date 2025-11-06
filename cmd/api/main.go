package main

import (
	"os"

	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/config"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/database"
	httpHandlers "github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/http"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
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

	svcs, err := service.New(db)
	if err != nil {
		log.Fatal().Err(err).Msg("service initialization failed")
	}

	app := fiber.New(fiber.Config{
		AppName: "Smart Energy Grid API v1.0",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "smart-energy-grid-api",
		})
	})

	httpHandlers.Register(app, svcs)

	// Support both API_ADDR and PORT for Elastic Beanstalk
	addr := viper.GetString("API_ADDR")
	if addr == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		addr = ":" + port
	}

	log.Info().Str("addr", addr).Msg("api listening")
	log.Fatal().Err(app.Listen(addr)).Msg("server exit")
}
