package middleware

import (
	"crypto/subtle"
	"encoding/base64"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

func KeyAuthMiddleware(apiKey string) fiber.Handler {
	return keyauth.New(keyauth.Config{
		Validator: func(c *fiber.Ctx, key string) (bool, error) {
			providedKey, err := base64.StdEncoding.DecodeString(key)
			if err != nil {
				return false, keyauth.ErrMissingOrMalformedAPIKey
			}

			if subtle.ConstantTimeCompare(providedKey, []byte(apiKey)) == 1 {
				return true, nil
			}
			return false, keyauth.ErrMissingOrMalformedAPIKey
		},
	})
}
