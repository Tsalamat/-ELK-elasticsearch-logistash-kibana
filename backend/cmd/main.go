package main

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	appinternal "todo/internal"
)

func main() {
	port := env("PORT", "8080")
	logPath := env("LOG_FILE", "/app/logs/todo-api.log")

	logger, err := appinternal.NewLogger(logPath, "todo-api")
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	store := appinternal.NewTodoStore()
	app := fiber.New(fiber.Config{
		AppName:      "todo-api",
		ServerHeader: "todo-api",
	})

	app.Use(corsMiddleware)
	app.Use(requestLogger(logger))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/todos", func(c *fiber.Ctx) error {
		todos := store.List()
		logger.Info("todos listed", fiber.Map{"count": len(todos)})
		return c.JSON(todos)
	})

	app.Post("/todos", func(c *fiber.Ctx) error {
		var input appinternal.TodoCreate
		if err := c.BodyParser(&input); err != nil {
			logger.Error("invalid create todo payload", fiber.Map{"error": err.Error()})
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
		}

		input.Title = strings.TrimSpace(input.Title)
		if input.Title == "" {
			logger.Error("empty todo title", nil)
			return fiber.NewError(fiber.StatusBadRequest, "title is required")
		}

		todo := store.Create(input)
		logger.Info("todo created", todo)
		return c.Status(fiber.StatusCreated).JSON(todo)
	})

	app.Patch("/todos/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid todo id")
		}

		var input appinternal.TodoUpdate
		if err := c.BodyParser(&input); err != nil {
			logger.Error("invalid update todo payload", fiber.Map{"error": err.Error()})
			return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body")
		}

		if input.Title != nil {
			trimmed := strings.TrimSpace(*input.Title)
			input.Title = &trimmed
		}

		todo, err := store.Update(id, input)
		if errors.Is(err, appinternal.ErrTodoNotFound) {
			logger.Error("todo update failed", fiber.Map{"id": id, "reason": "not found"})
			return fiber.NewError(fiber.StatusNotFound, "todo not found")
		}
		if err != nil {
			logger.Error("todo update failed", fiber.Map{"id": id, "error": err.Error()})
			return fiber.NewError(fiber.StatusInternalServerError, "failed to update todo")
		}

		logger.Info("todo updated", todo)
		return c.JSON(todo)
	})

	app.Delete("/todos/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid todo id")
		}

		if err := store.Delete(id); errors.Is(err, appinternal.ErrTodoNotFound) {
			logger.Error("todo delete failed", fiber.Map{"id": id, "reason": "not found"})
			return fiber.NewError(fiber.StatusNotFound, "todo not found")
		}

		logger.Info("todo deleted", fiber.Map{"id": id})
		return c.SendStatus(fiber.StatusNoContent)
	})

	logger.Info("todo API started", fiber.Map{"port": port, "log_file": logPath})
	if err := app.Listen(":" + port); err != nil {
		logger.Error("todo API stopped", fiber.Map{"error": err.Error()})
		panic(err)
	}
}

func requestLogger(logger *appinternal.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		startedAt := time.Now()
		err := c.Next()
		status := c.Response().StatusCode()
		if err != nil {
			if fiberErr, ok := err.(*fiber.Error); ok {
				status = fiberErr.Code
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		logger.Write(appinternal.LogEntry{
			Level:   levelForStatus(status),
			Message: "http request",
			Method:  c.Method(),
			Path:    c.Path(),
			Status:  status,
			Data: fiber.Map{
				"duration_ms": time.Since(startedAt).Milliseconds(),
				"ip":          c.IP(),
			},
		})

		return err
	}
}

func corsMiddleware(c *fiber.Ctx) error {
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
	c.Set("Access-Control-Allow-Headers", "Origin,Content-Type,Accept")

	if c.Method() == fiber.MethodOptions {
		return c.SendStatus(fiber.StatusNoContent)
	}

	return c.Next()
}

func env(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func levelForStatus(status int) string {
	if status >= 500 {
		return "error"
	}
	if status >= 400 {
		return "warn"
	}
	return "info"
}
