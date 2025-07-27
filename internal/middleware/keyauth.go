package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

func KeyAuthMiddleware(apiKey string) fiber.Handler {
	return keyauth.New(keyauth.Config{
		Validator: func(c *fiber.Ctx, key string) (bool, error) {
			providedKey, err := base64.StdEncoding.DecodeString(key)
			if err != nil {
				return false, err
			}

			if subtle.ConstantTimeCompare(providedKey, []byte(apiKey)) == 1 {
				return true, nil
			}

			return false, errors.New("invalid API key")
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Warnf("Access denied from '%s' - %s", c.IP(), err.Error())
			return c.Status(fiber.StatusUnauthorized).Send(nil)
		},
	})
}
