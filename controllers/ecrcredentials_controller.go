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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	registryv1alpha1 "github.com/astrokube/registry-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ECRCredentialsReconciler reconciles a ECRCredentials object
type ECRCredentialsReconciler struct {
	CredentialsReconciler
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=registry.astrokube.com,resources=ecrcredentials,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=registry.astrokube.com,resources=ecrcredentials/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=registry.astrokube.com,resources=ecrcredentials/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ECRCredentials object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *ECRCredentialsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("ecrcredentials", req.NamespacedName)

	ecrCredentials := &registryv1alpha1.ECRCredentials{}

	// Skip if ecrCredentials doesn't exists
	if err := r.Get(ctx, req.NamespacedName, ecrCredentials); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to get ECRCredentials")
		return ctrl.Result{}, err
	}

	// ecrCredentials is not going to be deleted
	if ecrCredentials.ObjectMeta.DeletionTimestamp.IsZero() {

		// If Authenticating status if is not set
		if ecrCredentials.Status.Phase == "" {
			if err := r.setStatus(log, ecrCredentials, registryv1alpha1.ECRCredentialsAuthenticating); err != nil {
				return ctrl.Result{}, err
			}
		}

		return r.authenticate(log, ecrCredentials)
	} else {
		// Set Terminating status
		if err := r.setStatus(log, ecrCredentials, registryv1alpha1.ECRCredentialsTerminating); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ECRCredentialsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&registryv1alpha1.ECRCredentials{}).
		Complete(r)
}

func (r *ECRCredentialsReconciler) authenticate(log logr.Logger, ecrCredentials *registryv1alpha1.ECRCredentials) (ctrl.Result, error) {
	awsSession, err := r.getAwsSession(log, ecrCredentials)
	if err != nil {
		if err := r.setError(log, ecrCredentials, err); err != nil {
			return ctrl.Result{}, err
		}
	}

	credentials, err := r.getToken(log, ecrCredentials, awsSession)
	if err != nil {
		if err := r.setError(log, ecrCredentials, err); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	secret := r.getSecret(*credentials)

	err = r.createOrUpdateSecret(log, &secret)
	if err != nil {
		if err := r.setError(log, ecrCredentials, err); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Set Authenticated status
	if err := r.setStatus(log, ecrCredentials, registryv1alpha1.ECRCredentialsAuthenticated); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ECRCredentialsReconciler) setError(log logr.Logger, ecrCredentials *registryv1alpha1.ECRCredentials, err error) error {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case "UnrecognizedClientException":
			// Set Error status
			if err := r.setStatus(log, ecrCredentials, registryv1alpha1.ECRCredentialsUnauthorized); err != nil {
				return err
			}
		default:
			// Set Error status
			if err := r.setStatus(log, ecrCredentials, registryv1alpha1.ECRCredentialsError); err != nil {
				return err
			}
		}
	}

	// Set ErrorMessage
	if err := r.setErrorMessage(log, ecrCredentials, err.Error()); err != nil {
		return err
	}

	return nil
}

func (r *ECRCredentialsReconciler) setErrorMessage(log logr.Logger, ecrCredentials *registryv1alpha1.ECRCredentials, message string) error {
	ctx := context.Background()

	ecrCredentials.Status.ErrorMessage = message
	if err := r.Status().Update(ctx, ecrCredentials); err != nil {
		log.Error(err, "Unable to set status")
		return err
	}

	return nil
}

func (r *ECRCredentialsReconciler) setStatus(log logr.Logger, ecrCredentials *registryv1alpha1.ECRCredentials, phase registryv1alpha1.ECRCredentialsPhase) error {
	ctx := context.Background()

	ecrCredentials.Status.Phase = phase
	if err := r.Status().Update(ctx, ecrCredentials); err != nil {
		log.Error(err, "Unable to set status")
		return err
	}

	return nil
}

func (r *ECRCredentialsReconciler) getAwsSession(log logr.Logger, ecrCredentials *registryv1alpha1.ECRCredentials) (*session.Session, error) {
	credentials := credentials.NewStaticCredentialsFromCreds(credentials.Value{
		AccessKeyID:     ecrCredentials.Spec.AccessKeyID,
		SecretAccessKey: ecrCredentials.Spec.SecretAccessKey,
	})
	awsConfig := &aws.Config{
		Credentials: credentials,
		Region:      aws.String(ecrCredentials.Spec.Region),
	}
	return session.NewSession(awsConfig)
}

func (r *ECRCredentialsReconciler) getToken(log logr.Logger, ecrCredentials *registryv1alpha1.ECRCredentials, awsSession *session.Session) (*RegistryCredentials, error) {
	svc := ecr.New(awsSession)
	input := &ecr.GetAuthorizationTokenInput{}

	result, err := svc.GetAuthorizationToken(input)
	if err != nil {
		log.Info("Unable to get authorization token")
		return nil, err
	}

	stsSvc := sts.New(awsSession)
	identity, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Info("Unable to get CallerIdentity")
		return nil, err
	}

	host := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", *identity.Account, ecrCredentials.Spec.Region)
	return &RegistryCredentials{
		Name:               ecrCredentials.ObjectMeta.Name,
		Namespace:          ecrCredentials.ObjectMeta.Namespace,
		Host:               host,
		AuthorizationToken: *result.AuthorizationData[0].AuthorizationToken,
		ExpiresAt:          result.AuthorizationData[0].ExpiresAt,
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(ecrCredentials, registryv1alpha1.GroupVersion.WithKind("ECRCredentials")),
		},
	}, nil
}
