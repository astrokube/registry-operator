package webhooks

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	registryv1alpha1 "github.com/astrokube/registry-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

type MutatePodWebhook struct {
	Client  client.Client
	Log     logr.Logger
	decoder *admission.Decoder
}

//+kubebuilder:webhook:path=/mutate-pod,mutating=true,failurePolicy=ignore,sideEffects=None,admissionReviewVersions=v1,groups=core,resources=pods,verbs=create;update,versions=v1,name=mutate-pod.registry.astrokube.io

func (w *MutatePodWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := w.Log.WithValues("route", req.Name)

	// Get Pod
	pod := &corev1.Pod{}
	err := w.decoder.Decode(req, pod)
	if err != nil {
		log.Error(err, "Unable to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Get Pod images
	images := []string{}
	for _, container := range pod.Spec.InitContainers {
		images = append(images, container.Image)
	}
	for _, container := range pod.Spec.Containers {
		images = append(images, container.Image)
	}

	// Get secrets to inject in the pod
	secretsToAdd := []string{}
	for _, image := range images {
		ecrSecrets, err := w.getSecretNamesForECRCredentials(image, pod.ObjectMeta.Namespace)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		secretsToAdd = append(secretsToAdd, ecrSecrets...)
	}

	// Inject secrets
	for _, secret := range secretsToAdd {
		pod.Spec.ImagePullSecrets = append(pod.Spec.ImagePullSecrets, corev1.LocalObjectReference{
			Name: secret,
		})
	}

	// Return the injected pod
	marshalledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshalledPod)
}

func (w *MutatePodWebhook) InjectDecoder(d *admission.Decoder) error {
	w.decoder = d
	return nil
}

func (w *MutatePodWebhook) getSecretNamesForECRCredentials(image, namespace string) ([]string, error) {
	ecrCredentialsList, err := w.getECRCredentialsList(namespace)
	if err != nil {
		return nil, err
	}

	secretNames := []string{}

	for _, ecrCredentials := range ecrCredentialsList.Items {
		for _, imageSelector := range ecrCredentials.Spec.ImageSelector {
			match, err := regexp.Match(imageSelector, []byte(image))
			if err != nil {
				return nil, err
			}
			if match {
				secretNames = append(secretNames, ecrCredentials.ObjectMeta.Name)
			}
		}
	}

	return secretNames, nil
}

func (w *MutatePodWebhook) getECRCredentialsList(namespace string) (*registryv1alpha1.ECRCredentialsList, error) {
	list := &registryv1alpha1.ECRCredentialsList{}
	err := w.Client.List(context.TODO(), list, &client.ListOptions{Namespace: namespace})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	return list, nil
}
