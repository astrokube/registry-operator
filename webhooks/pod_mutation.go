package webhooks

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	registryv1alpha1 "github.com/astrokube/registry-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
)

type MutatePodWebhook struct {
	Client   client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	decoder  *admission.Decoder
}

//+kubebuilder:webhook:path=/mutate-pod,mutating=true,failurePolicy=ignore,sideEffects=None,admissionReviewVersions=v1,groups=core,resources=pods,verbs=create;update,versions=v1,name=mutate-pod.registry.astrokube.io

func (w *MutatePodWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := w.Log.WithValues("pod", req.Name)

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

	// Wait 10 seconds for RegistryCredentials authentication process
	for i := 0; i < 10; i++ {
		if !w.isReadyForInjection(log, pod) {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	// Get secrets to inject in the pod
	secretsToAdd := []string{}
	for _, image := range images {
		ecrSecrets, err := w.getSecretNamesForRegistryCredentials(image, pod.ObjectMeta.Namespace)
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

func (w *MutatePodWebhook) getSecretNamesForRegistryCredentials(image, namespace string) ([]string, error) {
	registryCredentialsList, err := w.getRegistryCredentialsList(namespace)
	if err != nil {
		return nil, err
	}

	secretNames := []string{}

	for _, registryCredentials := range registryCredentialsList.Items {
		for _, imageSelector := range registryCredentials.Spec.ImageSelector.MatchHostRegexp {
			match, err := regexp.Match(imageSelector, []byte(image))
			if err != nil {
				return nil, err
			}
			if match {
				secretNames = append(secretNames, registryCredentials.ObjectMeta.Name)
			}
		}
	}

	return secretNames, nil
}

func (w *MutatePodWebhook) getRegistryCredentialsList(namespace string) (*registryv1alpha1.RegistryCredentialsList, error) {
	list := &registryv1alpha1.RegistryCredentialsList{}
	err := w.Client.List(context.TODO(), list, &client.ListOptions{Namespace: namespace})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	return list, nil
}

func (w *MutatePodWebhook) isReadyForInjection(log logr.Logger, pod *corev1.Pod) bool {
	registryCredentialsList, err := w.getRegistryCredentialsList(pod.ObjectMeta.Namespace)
	if err != nil {
		log.Error(err, "Unable to get RegistryCredentials list")
		return false
	}

	for _, registryCredentials := range registryCredentialsList.Items {
		if registryCredentials.Status.State == registryv1alpha1.RegistryCredentialsAuthenticating || registryCredentials.Status.State == "" {
			w.Recorder.Eventf(pod, corev1.EventTypeWarning, "Creating pod", "Waiting to create Pod will beacuse the authentication process is not finished for the RegistryCredentials \"%v\"", registryCredentials.ObjectMeta.Name)
			return false
		}
	}

	return true
}
