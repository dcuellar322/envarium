//go:build !darwin && !linux && !windows

package process

import (
	"context"
	"errors"

	"envarium.dev/envarium/internal/runtime/domain"
)

func resolveKeychain(context.Context, domain.SecretReference) (string, error) {
	return "", errors.New("operating-system keychain lookup is unavailable")
}
