package providers

import (
	"github.com/astrokube/registry-controller/api/v1alpha1"
	"github.com/go-logr/logr"
)

type Authenticator interface {
	GetToken(log logr.Logger, registryCredentials *v1alpha1.RegistryCredentials) *AuthenticationIntent
}
