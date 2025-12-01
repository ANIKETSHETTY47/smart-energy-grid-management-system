package main

import (
	"os"

	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/config"
	dashboardAPI "github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/dashboard/api"
	dashboardServer "github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/dashboard/server"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/database"
	httpHandlers "github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/http"
	"github.com/ANIKETSHETTY47/smart-energy-grid-management-system/internal/service"
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info.IsDir()
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Load configuration
	if err := config.Load(); err != nil {
		log.Fatal().Err(err).Msg("config load failed")
	}

	// Connect to database if not using cloud
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

	// Initialize services
	svcs, err := service.New(db)
	if err != nil {
		log.Fatal().Err(err).Msg("service initialization failed")
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Smart Energy Grid API v1.0",
	})

	// Add middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Register API routes
	httpHandlers.Register(app, svcs)

	// Check if dashboard exists
	dashboardExists := fileExists("./web/dashboard/static")

	if dashboardExists {
		// Setup dashboard
		dashClient := dashboardAPI.New()
		dashHandler := dashboardServer.New(
			"./web/dashboard/static",
			"./web/dashboard/templates",
			dashClient,
		)

		// Serve static files
		app.Static("/static", "./web/dashboard/static")

		// Dashboard routes
		app.All("/healthz", adaptor.HTTPHandler(dashHandler))
		app.All("/ws", adaptor.HTTPHandler(dashHandler))
		app.Get("/api/stats", adaptor.HTTPHandler(dashHandler))
		app.All("/dashboard", adaptor.HTTPHandler(dashHandler))
		app.All("/equipment", adaptor.HTTPHandler(dashHandler))
		app.All("/analytics", adaptor.HTTPHandler(dashHandler))
		app.All("/alerts", adaptor.HTTPHandler(dashHandler))
		app.Post("/alerts/acknowledge", adaptor.HTTPHandler(dashHandler))

		log.Info().Msg("Dashboard routes enabled")
	} else {
		// API-only health check
		app.Get("/healthz", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"status": "healthy", "service": "smart-energy-grid-api"})
		})

		log.Info().Msg("Running in API-only mode (dashboard not deployed)")
	}

	// Get server address
	addr := viper.GetString("API_ADDR")
	if addr == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		addr = ":" + port
	}

	// Start server
	log.Info().Str("addr", addr).Msg("unified api + dashboard listening")
	log.Fatal().Err(app.Listen(addr)).Msg("server exit")
}
