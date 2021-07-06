package controllers

import (
	"context"
	"os"
	"time"

	registryv1alpha1 "github.com/astrokube/registry-controller/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("RegistryCredentials controller", func() {

	const (
		timeout   = time.Second * 5
		interval  = time.Second * 1
		namespace = "default"
	)

	Context("When creating RegistryCredentials", func() {
		It("Should set RegistryCredentials.Status to Unathorized when credentials are not valid", func() {
			By("By creating a new RegistryCredentials")
			ctx := context.Background()
			name := "invalid-credentials"
			r := &registryv1alpha1.RegistryCredentials{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: registryv1alpha1.RegistryCredentialsSpec{
					Provider: registryv1alpha1.RegistryProvider{
						AWSElasticContainerRegistry: &registryv1alpha1.AWSElasticContainerRegistry{
							AccessKeyID:     "test",
							SecretAccessKey: "test",
							Region:          "eu-central-1",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, r)).Should(Succeed())

			fetched := &registryv1alpha1.RegistryCredentials{}
			Eventually(func() registryv1alpha1.RegistryCredentialsState {
				k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      name,
					Namespace: namespace,
				}, fetched)
				return fetched.Status.State
			}, timeout, interval).Should(Equal(registryv1alpha1.RegistryCredentialsUnauthorized))
		})

		It("Should set RegistryCredentials.Status to Error when region is not valid", func() {
			By("By creating a new RegistryCredentials")
			ctx := context.Background()
			name := "invalid-region"
			r := &registryv1alpha1.RegistryCredentials{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: registryv1alpha1.RegistryCredentialsSpec{
					Provider: registryv1alpha1.RegistryProvider{
						AWSElasticContainerRegistry: &registryv1alpha1.AWSElasticContainerRegistry{
							AccessKeyID:     "test",
							SecretAccessKey: "test",
							Region:          "invalid",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, r)).Should(Succeed())

			fetched := &registryv1alpha1.RegistryCredentials{}
			Eventually(func() registryv1alpha1.RegistryCredentialsState {
				k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      name,
					Namespace: namespace,
				}, fetched)
				return fetched.Status.State
			}, timeout, interval).Should(Equal(registryv1alpha1.RegistryCredentialsErrored))
		})

		It("Should set RegistryCredentials.Status to Error when provider is not set", func() {
			By("By creating a new RegistryCredentials")
			ctx := context.Background()
			name := "provider-not-set"
			r := &registryv1alpha1.RegistryCredentials{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: registryv1alpha1.RegistryCredentialsSpec{},
			}
			Expect(k8sClient.Create(ctx, r)).Should(Succeed())

			fetched := &registryv1alpha1.RegistryCredentials{}
			Eventually(func() registryv1alpha1.RegistryCredentialsState {
				k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      name,
					Namespace: namespace,
				}, fetched)
				return fetched.Status.State
			}, timeout, interval).Should(Equal(registryv1alpha1.RegistryCredentialsErrored))

			Eventually(func() string {
				k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      name,
					Namespace: namespace,
				}, fetched)
				return fetched.Status.ErrorMessage
			}, timeout, interval).Should(Equal("Provider not implemented"))

		})

		if os.Getenv("ENABLE_ALL_TESTS") == "true" {
			It("Should set RegistryCredentials.Status to Authenticated with valid credentials", func() {
				By("By creating a new RegistryCredentials")
				ctx := context.Background()
				name := "valid"
				r := &registryv1alpha1.RegistryCredentials{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
					Spec: registryv1alpha1.RegistryCredentialsSpec{
						Provider: registryv1alpha1.RegistryProvider{
							AWSElasticContainerRegistry: &registryv1alpha1.AWSElasticContainerRegistry{
								AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
								SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
								Region:          os.Getenv("AWS_REGION"),
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, r)).Should(Succeed())

				fetched := &registryv1alpha1.RegistryCredentials{}
				Eventually(func() registryv1alpha1.RegistryCredentialsState {
					k8sClient.Get(context.Background(), types.NamespacedName{
						Name:      name,
						Namespace: namespace,
					}, fetched)
					return fetched.Status.State
				}, timeout, interval).Should(Equal(registryv1alpha1.RegistryCredentialsAuthenticated))
			})
		}

	})

})
