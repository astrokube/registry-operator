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

package controllers

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/astrokube/registry-controller/api/v1alpha1"
	registryv1alpha1 "github.com/astrokube/registry-controller/api/v1alpha1"
	"github.com/astrokube/registry-controller/pkg/providers"
	"github.com/go-logr/logr"
)

// RegistryCredentialsReconciler reconciles a RegistryCredentials object
type RegistryCredentialsReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=registry.astrokube.com,resources=registrycredentials,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=registry.astrokube.com,resources=registrycredentials/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=registry.astrokube.com,resources=registrycredentials/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RegistryCredentials object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *RegistryCredentialsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	registryCredentials := &registryv1alpha1.RegistryCredentials{}

	// Skip if registryCredentials doesn't exists
	if err := r.Get(ctx, req.NamespacedName, registryCredentials); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		l.Error(err, "Unable to get RegistryCredentials")
		return ctrl.Result{}, err
	}

	// registryCredentials is not going to be deleted
	if registryCredentials.ObjectMeta.DeletionTimestamp.IsZero() {

		switch registryCredentials.Status.State {
		case "":
			if err := r.authenticate(l, registryCredentials); err != nil {
				return ctrl.Result{
					RequeueAfter: time.Minute * 5,
				}, err
			}
			return ctrl.Result{}, nil

		case registryv1alpha1.RegistryCredentialsAuthenticated:
			return ctrl.Result{}, nil

		case registryv1alpha1.RegistryCredentialsErrored:
			if err := r.authenticate(l, registryCredentials); err != nil {
				return ctrl.Result{
					RequeueAfter: time.Minute * 5,
				}, err
			}
			return ctrl.Result{}, nil

		case registryv1alpha1.RegistryCredentialsAuthenticating:
			if err := r.authenticate(l, registryCredentials); err != nil {
				return ctrl.Result{
					RequeueAfter: time.Minute * 5,
				}, err
			}
			return ctrl.Result{}, nil

		case registryv1alpha1.RegistryCredentialsUnauthorized:
			if err := r.authenticate(l, registryCredentials); err != nil {
				return ctrl.Result{
					RequeueAfter: time.Minute * 5,
				}, err
			}

			return ctrl.Result{}, nil
		}

	} else {
		// Set Terminating status
		if err := r.setStatus(l, registryCredentials, registryv1alpha1.RegistryCredentialsTerminating); err != nil {
			return ctrl.Result{
				RequeueAfter: time.Minute * 5,
			}, err
		}

		return ctrl.Result{}, nil

	}

	// your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RegistryCredentialsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&registryv1alpha1.RegistryCredentials{}).
		Complete(r)
}

func (r *RegistryCredentialsReconciler) setStatus(log logr.Logger, registryCredentials *registryv1alpha1.RegistryCredentials, state registryv1alpha1.RegistryCredentialsState) error {
	ctx := context.Background()

	if state != v1alpha1.RegistryCredentialsErrored {
		registryCredentials.Status.ErrorMessage = ""
	}
	registryCredentials.Status.State = state
	if err := r.Status().Update(ctx, registryCredentials); err != nil {
		log.Error(err, "Unable to set status")
		return err
	}

	return nil
}

func (r *RegistryCredentialsReconciler) authenticate(log logr.Logger, registryCredentials *registryv1alpha1.RegistryCredentials) error {
	authenticator, err := r.getAuthenticator(log, registryCredentials)
	if err != nil {
		log.Error(err, "Unable to get authenticator")
		if err := r.setError(log, registryCredentials, err); err != nil {
			return err
		}
		return nil
	}

	intent := authenticator.GetToken(log, registryCredentials)

	switch intent.State {
	case v1alpha1.RegistryCredentialsErrored:
		if err := r.setError(log, registryCredentials, intent.Error); err != nil {
			log.Error(err, "Unable to set error")
			return err
		}

		return nil
	case v1alpha1.RegistryCredentialsAuthenticated:
		secret := r.getSecret(registryCredentials, *intent)

		err = r.createOrUpdateSecret(log, &secret)
		if err != nil {
			if err := r.setError(log, registryCredentials, err); err != nil {
				log.Error(err, "Unable to set error")
				return err
			}

			return nil
		}

		// Set Authenticated status
		if err := r.setStatus(log, registryCredentials, registryv1alpha1.RegistryCredentialsAuthenticated); err != nil {
			log.Error(err, "Unable to set status")
			return err
		}
		return nil
	default:
		if err := r.setStatus(log, registryCredentials, intent.State); err != nil {
			log.Error(err, "Unable to set status")
			return err
		}

		return nil
	}
}

func (r *RegistryCredentialsReconciler) getAuthenticator(log logr.Logger, registryCredentials *registryv1alpha1.RegistryCredentials) (providers.Authenticator, error) {
	if registryCredentials.Spec.Provider.AWSElasticContainerRegistry != nil {
		return providers.NewAWSElasticContainerRegistryAuthenticator(registryCredentials.Spec.Provider.AWSElasticContainerRegistry), nil
	}

	return nil, fmt.Errorf("Provider not implemented")
}

func (r *RegistryCredentialsReconciler) getSecret(registryCredentials *v1alpha1.RegistryCredentials, intent providers.AuthenticationIntent) corev1.Secret {
	dockerConfig := fmt.Sprintf("{\"auths\":{\"%v\":{\"auth\":\"%v\"}}}", intent.Registry, intent.Token)

	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      registryCredentials.ObjectMeta.Name,
			Namespace: registryCredentials.ObjectMeta.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(registryCredentials, v1alpha1.GroupVersion.WithKind("RegistryCredentials")),
			},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: []byte(dockerConfig),
		},
	}
}

func (r *RegistryCredentialsReconciler) createOrUpdateSecret(log logr.Logger, object *corev1.Secret) error {
	ctx := context.Background()

	err := r.Client.Get(ctx, client.ObjectKey{
		Name:      object.ObjectMeta.Name,
		Namespace: object.ObjectMeta.Namespace,
	}, &corev1.Secret{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if err != nil && errors.IsNotFound(err) {
		if err := r.Client.Create(ctx, object); err != nil {
			log.Error(err, "Unable to create object")
			return client.IgnoreNotFound(err)
		}
		r.Recorder.Eventf(object, corev1.EventTypeNormal, "Created", "Created secret %q", object.ObjectMeta.Name)
	} else {
		if err := r.Client.Update(ctx, object); err != nil {
			log.Error(err, "Unable to update object")
			return client.IgnoreNotFound(err)
		}
		r.Recorder.Eventf(object, corev1.EventTypeNormal, "Updated", "Updated secret %q", object.ObjectMeta.Name)
	}

	return nil
}

func (r *RegistryCredentialsReconciler) setError(log logr.Logger, registryCredentials *registryv1alpha1.RegistryCredentials, err error) error {
	// Set Error status
	if err := r.setStatus(log, registryCredentials, registryv1alpha1.RegistryCredentialsErrored); err != nil {
		log.Error(err, "Unable to set status")
		return err
	}

	// Set ErrorMessage
	if err := r.setErrorMessage(log, registryCredentials, err.Error()); err != nil {
		log.Error(err, "Unable to set ErrorMessage")
		return err
	}

	return nil
}

func (r *RegistryCredentialsReconciler) setErrorMessage(log logr.Logger, registryCredentials *registryv1alpha1.RegistryCredentials, message string) error {
	ctx := context.Background()

	registryCredentials.Status.ErrorMessage = message
	if err := r.Status().Update(ctx, registryCredentials); err != nil {
		log.Error(err, "Unable to set status")
		return err
	}

	return nil
}
