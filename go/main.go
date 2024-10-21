package main

import (
	"context"
	"errors"
	"go-otlp-collector/dice"
	"os"
	"os/signal"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panicf("Error loading .env file: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		log.Panicf("Error setup otlp")
	}
	defer func() {
		log.Warn("Shutting down Open-Telemetry")
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	PORT := os.Getenv("PORT")

	app := fiber.New()

	RegisterRoutes(ctx, app)

	app.Use(cors.New())
	app.Use(logger.New())

	app.Listen(":" + PORT)

}

func RegisterRoutes(ctx context.Context, api *fiber.App) {

	api.Get("/roll/:number", dice.RollDice)
}
