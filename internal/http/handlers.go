package http

import (
	"github.com/gofiber/fiber/v2"
	"smart/internal/service"
)

func Register(app *fiber.App, svcs *service.Services) {
	g := app.Group("/")
	g.Get("facilities", func(c *fiber.Ctx) error {
		items, err := svcs.Repos.ListFacilities()
		if err != nil { return c.Status(500).JSON(fiber.Map{"error": err.Error()}) }
		return c.JSON(items)
	})
	g.Get("meters", func(c *fiber.Ctx) error {
		items, err := svcs.Repos.ListMeters()
		if err != nil { return c.Status(500).JSON(fiber.Map{"error": err.Error()}) }
		return c.JSON(items)
	})
}
