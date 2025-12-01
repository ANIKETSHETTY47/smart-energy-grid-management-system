package main

import (
	"os"

	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/config"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/database"
	httpHandlers "github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/http"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/service"
	dashboardAPI "github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/dashboard/api"
	dashboardServer "github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/dashboard/server"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if err := config.Load(); err != nil {
		log.Fatal().Err(err).Msg("config load failed")
	}

	// Only connect to database if not using cloud services
	var db *sqlx.DB
	useCloud := viper.GetBool("USE_CLOUD_SERVICES")
	if !useCloud {
		var err error
		db, err = database.Connect()
		if err != nil {
			log.Fatal().Err(err).Msg("db connect failed")
		}
		defer db.Close()
	} else {
		log.Info().Msg("Using cloud services (DynamoDB) - skipping PostgreSQL connection")
	}

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

	// Create dashboard handler
	dashClient := dashboardAPI.New()
	dashHandler := dashboardServer.New(
		"./web/dashboard/static",
		"./web/dashboard/templates",
		dashClient,
	)

	// Mount static assets FIRST
	app.Static("/static", "./web/dashboard/static")

	// Register API routes BEFORE dashboard (they have priority)
	httpHandlers.Register(app, svcs)

	// Dashboard WebSocket and API stats
	app.All("/healthz", adaptor.HTTPHandler(dashHandler))
	app.All("/ws", adaptor.HTTPHandler(dashHandler))
	app.Get("/api/stats", adaptor.HTTPHandler(dashHandler))

	// Dashboard HTML routes (these come AFTER API routes)
	app.All("/dashboard", adaptor.HTTPHandler(dashHandler))
	app.All("/equipment", adaptor.HTTPHandler(dashHandler))
	app.All("/analytics", adaptor.HTTPHandler(dashHandler))
	app.All("/alerts", adaptor.HTTPHandler(dashHandler))
	app.Post("/alerts/acknowledge", adaptor.HTTPHandler(dashHandler))
	
	// Root route serves dashboard
	app.All("/", adaptor.HTTPHandler(dashHandler))

	// Support both API_ADDR and PORT for Elastic Beanstalk
	addr := viper.GetString("API_ADDR")
	if addr == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		addr = ":" + port
	}

	log.Info().Str("addr", addr).Msg("unified api + dashboard listening")
	log.Fatal().Err(app.Listen(addr)).Msg("server exit")
}
