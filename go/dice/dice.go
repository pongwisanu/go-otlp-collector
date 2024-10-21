package dice

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

const name = "rolldice.service"

var (
	tracer = otel.Tracer(name)
	meter  = otel.Meter(name)
	logger = otelslog.NewLogger(name)
)

func RollDice(c *fiber.Ctx) error {
	spanName := fmt.Sprintf("%s.RollDice", name)
	successCounter, err := meter.Int64Counter(fmt.Sprintf("%s.success", spanName), metric.WithDescription("Counts successful RollDice requests"))
	if err != nil {
		log.Fatal(err)
	}

	errCounter, err := meter.Int64Counter(fmt.Sprintf("%s.error", spanName), metric.WithDescription("Counts successful RollDice requests"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, span := tracer.Start(c.Context(), spanName)
	defer span.End()

	span.SetAttributes(attribute.String("http.method", c.Method()),
		attribute.String("http.uri", c.OriginalURL()),
		attribute.String("http.schema", c.Protocol()),
		attribute.String("http.domain", c.Hostname()),
		attribute.String("http.client", c.IP()),
	)

	number, err := c.ParamsInt("number")

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		logger.WarnContext(ctx, spanName, "result", err.Error())
		errCounter.Add(ctx, 1)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error",
			"error":   err.Error(),
		})
	}

	result, err := Roll(ctx, number)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		logger.WarnContext(ctx, spanName, "result", err.Error())
		errCounter.Add(ctx, 1)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error",
			"error":   err.Error(),
		})
	}

	successCounter.Add(ctx, 1)
	logger.InfoContext(ctx, spanName, "result", "success")

	return c.JSON(fiber.Map{
		"message": "Success",
		"value":   result,
	})
}

func Roll(ctx context.Context, numberOfDice int) (int, error) {
	spanName := fmt.Sprintf("%s.Roll", name)
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	if numberOfDice > 3 {
		err_msg := fmt.Errorf("number of dice must below or equal 3")
		logger.WarnContext(ctx, spanName, "numberOfDice", numberOfDice, "result", err_msg)
		return 0, err_msg
	}

	result := 0
	for i := 0; i < numberOfDice; i++ {
		roll := 1 + rand.Intn(6)
		result = result + roll
	}

	logger.InfoContext(ctx, spanName, "numberOfDice", numberOfDice, "result", result)
	return result, nil

}
