package controllers

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RegistryCredentials struct {
	Name               string
	Namespace          string
	Host               string
	AuthorizationToken string
	ExpiresAt          *time.Time
	OwnerReferences    []metav1.OwnerReference
}
