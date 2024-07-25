package main

import (
	"fmt"
	"log"

	"github.com/caitlinelfring/go-env-default"
	"github.com/gofiber/fiber/v2"
)

var httpPort = env.GetIntDefault("HTTP_PORT", 3000)

func main() {
	app := fiber.New()

	app.Post("/:name", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "TODO: Deploy " + c.Params("name")})
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%d", httpPort)))
}
