package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/markhaur/trivia/database"
)

func main() {
	database.ConnectDb()

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, markhaur here!")
	})

	app.Listen(":3000")
}
