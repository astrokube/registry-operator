package providers

import (
	"time"

	"github.com/astrokube/registry-controller/api/v1alpha1"
)

type AuthenticationIntent struct {
	ExpiresAt *time.Time
	Registry  string
	State     v1alpha1.RegistryCredentialsState
	Token     string
	Error     error
}
