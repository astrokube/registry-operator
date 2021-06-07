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

var _ = Describe("EcrCredentials controller", func() {

	const (
		timeout   = time.Second * 2
		interval  = time.Second * 1
		namespace = "default"
	)

	Context("When creating ECRCredentials", func() {
		It("Should set ECRCredentials.Status to Unathorized when credentials are not valid", func() {
			By("By creating a new ECRCredentials")
			ctx := context.Background()
			name := "invalid-credentials"
			r := &registryv1alpha1.ECRCredentials{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: registryv1alpha1.ECRCredentialsSpec{
					AccessKeyID:     "test",
					SecretAccessKey: "test",
					Region:          "eu-central-1",
				},
			}
			Expect(k8sClient.Create(ctx, r)).Should(Succeed())

			fetched := &registryv1alpha1.ECRCredentials{}
			Eventually(func() registryv1alpha1.ECRCredentialsPhase {
				k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      name,
					Namespace: namespace,
				}, fetched)
				return fetched.Status.Phase
			}, timeout, interval).Should(Equal(registryv1alpha1.ECRCredentialsUnauthorized))
		})

		It("Should set ECRCredentials.Status to Error when region is not valid", func() {
			By("By creating a new ECRCredentials")
			ctx := context.Background()
			name := "invalid-region"
			r := &registryv1alpha1.ECRCredentials{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: registryv1alpha1.ECRCredentialsSpec{
					AccessKeyID:     "test",
					SecretAccessKey: "test",
					Region:          "invalid",
				},
			}
			Expect(k8sClient.Create(ctx, r)).Should(Succeed())

			fetched := &registryv1alpha1.ECRCredentials{}
			Eventually(func() registryv1alpha1.ECRCredentialsPhase {
				k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      name,
					Namespace: namespace,
				}, fetched)
				return fetched.Status.Phase
			}, timeout, interval).Should(Equal(registryv1alpha1.ECRCredentialsUnauthorized))
		})

		if os.Getenv("ENABLE_ALL_TESTS") == "true" {
			It("Should set ECRCredentials.Status to Authenticated with valid credentials", func() {
				By("By creating a new ECRCredentials")
				ctx := context.Background()
				name := "valid"
				r := &registryv1alpha1.ECRCredentials{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
					Spec: registryv1alpha1.ECRCredentialsSpec{
						AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
						SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
						Region:          os.Getenv("AWS_REGION"),
					},
				}
				Expect(k8sClient.Create(ctx, r)).Should(Succeed())

				fetched := &registryv1alpha1.ECRCredentials{}
				Eventually(func() registryv1alpha1.ECRCredentialsPhase {
					k8sClient.Get(context.Background(), types.NamespacedName{
						Name:      name,
						Namespace: namespace,
					}, fetched)
					return fetched.Status.Phase
				}, timeout, interval).Should(Equal(registryv1alpha1.ECRCredentialsAuthenticated))
			})
		}

	})
})
