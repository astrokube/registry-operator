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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RegistryCredentialsSpec defines the desired state of RegistryCredentials
type RegistryCredentialsSpec struct {
	//+kubebuilder:validation:Required
	Provider RegistryProvider `json:"provider"`

	// Foo is an example field of RegistryCredentials. Edit registrycredentials_types.go to remove/update
	ImageSelector ImageSelector `json:"imageSelector,omitempty"`
}

type ImageSelector struct {
	//+kubebuilder:validation:Optional
	MatchRegexp []string `json:"matchRegexp,omitempty"`

	//+kubebuilder:validation:Optional
	MatchExact []string `json:"matchExact,omitempty"`

	//+kubebuilder:validation:Optional
	MatchHostRegexp []string `json:"matchHostRegexp,omitempty"`

	//+kubebuilder:validation:Optional
	MatchHostExact []string `json:"matchHostExact,omitempty"`
}

type RegistryProvider struct {
	//+kubebuilder:validation:Optional
	AWSElasticContainerRegistry *AWSElasticContainerRegistry `json:"awsElasticContainerRegistry,omitempty"`
}

type AWSElasticContainerRegistry struct {
	//+kubebuilder:validation:Optional
	AccessKeyID string `json:"accessKeyId,omitempty"`

	//+kubebuilder:validation:Optional
	Region string `json:"region,omitempty"`

	//+kubebuilder:validation:Optional
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
}

// RegistryCredentialsStatus defines the observed state of RegistryCredentials
type RegistryCredentialsStatus struct {
	//+kubebuilder:validation:Optional
	State RegistryCredentialsState `json:"state,omitempty"`

	//+kubebuilder:validation:Optional
	ErrorMessage string `json:"errorMessage,omitempty"`

	//+kubebuilder:validation:Optional
	ExpirationTime *metav1.Time `json:"expirationTime,omitempty"`

	//+kubebuilder:validation:Optional
	AuthenticatedTime *metav1.Time `json:"authenticatedTime,omitempty"`
}

type RegistryCredentialsState string

var (
	RegistryCredentialsAuthenticated  RegistryCredentialsState = "Authenticated"
	RegistryCredentialsAuthenticating RegistryCredentialsState = "Authenticating"
	RegistryCredentialsErrored        RegistryCredentialsState = "Errored"
	RegistryCredentialsTerminating    RegistryCredentialsState = "Terminating"
	RegistryCredentialsUnauthorized   RegistryCredentialsState = "Unauthorized"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.state`

// RegistryCredentials is the Schema for the registrycredentials API
type RegistryCredentials struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegistryCredentialsSpec   `json:"spec,omitempty"`
	Status RegistryCredentialsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RegistryCredentialsList contains a list of RegistryCredentials
type RegistryCredentialsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RegistryCredentials `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RegistryCredentials{}, &RegistryCredentialsList{})
}
