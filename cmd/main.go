package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/markhaur/trivia/database"
	"github.com/markhaur/trivia/handlers"
)

func main() {
	database.ConnectDb()

	app := fiber.New()

	setupRoutes(app)

	app.Listen(":3000")
}

func setupRoutes(app *fiber.App) {
	app.Get("/home", handlers.Home)
	app.Get("/", handlers.ListFacts)
	app.Post("/fact", handlers.CreateFact)
}
