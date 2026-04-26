package middleware

import (
	"errors"
	"strings"
)

var (
	errAuthorizationHeaderMissing = errors.New("authorization header missing")
	errAuthorizationHeaderInvalid = errors.New("authorization header invalid")
)

func parseBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errAuthorizationHeaderMissing
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errAuthorizationHeaderInvalid
	}

	return parts[1], nil
}
