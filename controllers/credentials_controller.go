package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CredentialsReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

func (r *CredentialsReconciler) getSecret(credentials RegistryCredentials) corev1.Secret {
	dockerConfig := fmt.Sprintf("{\"auths\":{\"%v\":{\"auth\":\"%v\"}}}", credentials.Host, credentials.AuthorizationToken)

	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            credentials.Name,
			Namespace:       credentials.Namespace,
			OwnerReferences: credentials.OwnerReferences,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: []byte(dockerConfig),
		},
	}
}

func (r *CredentialsReconciler) createOrUpdateSecret(log logr.Logger, object *corev1.Secret) error {
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
