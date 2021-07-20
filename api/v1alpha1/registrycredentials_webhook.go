/*
Copyright 2021 AstroKube.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var registrycredentialslog = logf.Log.WithName("registrycredentials-resource")

func (r *RegistryCredentials) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-registry-astrokube-com-v1alpha1-registrycredentials,mutating=true,failurePolicy=fail,sideEffects=None,groups=registry.astrokube.com,resources=registrycredentials,verbs=create;update,versions=v1alpha1,name=mregistrycredentials.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &RegistryCredentials{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *RegistryCredentials) Default() {
	registrycredentialslog.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-registry-astrokube-com-v1alpha1-registrycredentials,mutating=false,failurePolicy=fail,sideEffects=None,groups=registry.astrokube.com,resources=registrycredentials,verbs=create;update,versions=v1alpha1,name=vregistrycredentials.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &RegistryCredentials{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *RegistryCredentials) ValidateCreate() error {
	registrycredentialslog.Info("validate create", "name", r.Name)

	if r.Spec.Provider.AWSElasticContainerRegistry == nil {
		return errors.New("You must set a provider")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *RegistryCredentials) ValidateUpdate(old runtime.Object) error {
	registrycredentialslog.Info("validate update", "name", r.Name)

	if r.Spec.Provider.AWSElasticContainerRegistry == nil {
		return ErrProviderNotSet
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *RegistryCredentials) ValidateDelete() error {
	registrycredentialslog.Info("validate delete", "name", r.Name)

	return nil
}
