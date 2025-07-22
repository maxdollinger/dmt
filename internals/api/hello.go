package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type HelloService struct {
	db *pgx.Conn
}

func NewHelloService(db *pgx.Conn) *HelloService {
	return &HelloService{db: db}
}

func (s *HelloService) Hello(c *fiber.Ctx) error {
	var greeting string
	err := s.db.QueryRow(c.Context(), "select 'Hello, World!'").Scan(&greeting)
	if err != nil {
		return c.Status(500).SendString("Database error: " + err.Error())
	}
	return c.SendString(greeting)
}
