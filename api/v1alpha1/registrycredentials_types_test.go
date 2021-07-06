package v1alpha1

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("RegistryCredentials type", func() {

	const (
		timeout   = time.Second * 10
		interval  = time.Second * 1
		namespace = "default"
	)

	Context("When creating RegistryCredentials", func() {
		It("Should fails", func() {
			By("By creating without Provider")

			ctx := context.Background()
			name := "invalid"
			r := &RegistryCredentials{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: RegistryCredentialsSpec{},
			}
			fmt.Fprintf(GinkgoWriter, "Creating: %v\n", r)
			Expect(k8sClient.Create(ctx, r)).ShouldNot(Succeed())
		})

		if os.Getenv("ENABLE_ALL_TESTS") == "true" {
			It("Should create an object successfully", func() {
				By("By creating a new RegistryCredentials")

				ctx := context.Background()
				name := "valid"
				r := &RegistryCredentials{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
					Spec: RegistryCredentialsSpec{
						Provider: RegistryProvider{
							AWSElasticContainerRegistry: &AWSElasticContainerRegistry{
								AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
								SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
								Region:          os.Getenv("AWS_REGION"),
							},
						},
					},
				}
				fmt.Fprintf(GinkgoWriter, "Creating: %v\n", r)
				Expect(k8sClient.Create(ctx, r)).Should(Succeed())
			})
		}
	})
})
