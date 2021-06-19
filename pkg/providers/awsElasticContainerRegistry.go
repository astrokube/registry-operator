package providers

import (
	"fmt"

	"github.com/astrokube/registry-controller/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/go-logr/logr"
)

func NewAWSElasticContainerRegistryAuthenticator(provider *v1alpha1.AWSElasticContainerRegistry) Authenticator {
	return &awsElasticContainerRegistryAuthenticator{
		AccessKeyID:     provider.AccessKeyID,
		Region:          provider.Region,
		SecretAccessKey: provider.SecretAccessKey,
	}
}

type awsElasticContainerRegistryAuthenticator struct {
	AccessKeyID     string
	Region          string
	SecretAccessKey string
}

func (c *awsElasticContainerRegistryAuthenticator) GetToken(log logr.Logger, registryCredentials *v1alpha1.RegistryCredentials) *AuthenticationIntent {
	awsSession, err := c.getAwsSession(log, registryCredentials)
	if err != nil {
		return &AuthenticationIntent{
			State: c.getState(err),
			Error: err,
		}
	}

	svc := ecr.New(awsSession)
	input := &ecr.GetAuthorizationTokenInput{}

	result, err := svc.GetAuthorizationToken(input)
	if err != nil {
		log.Info("Unable to get authorization token")
		return &AuthenticationIntent{
			State: c.getState(err),
			Error: err,
		}
	}

	stsSvc := sts.New(awsSession)
	identity, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Info("Unable to get CallerIdentity")
		return &AuthenticationIntent{
			State: c.getState(err),
			Error: err,
		}
	}

	registry := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", *identity.Account, registryCredentials.Spec.Provider.AWSElasticContainerRegistry.Region)

	return &AuthenticationIntent{
		Registry:  registry,
		State:     v1alpha1.RegistryCredentialsAuthenticated,
		Token:     *result.AuthorizationData[0].AuthorizationToken,
		ExpiresAt: result.AuthorizationData[0].ExpiresAt,
	}
}

func (r *awsElasticContainerRegistryAuthenticator) getAwsSession(log logr.Logger, registryCredentials *v1alpha1.RegistryCredentials) (*session.Session, error) {
	credentials := credentials.NewStaticCredentialsFromCreds(credentials.Value{
		AccessKeyID:     registryCredentials.Spec.Provider.AWSElasticContainerRegistry.AccessKeyID,
		SecretAccessKey: registryCredentials.Spec.Provider.AWSElasticContainerRegistry.SecretAccessKey,
	})
	awsConfig := &aws.Config{
		Credentials: credentials,
		Region:      aws.String(registryCredentials.Spec.Provider.AWSElasticContainerRegistry.Region),
	}
	return session.NewSession(awsConfig)
}

func (r *awsElasticContainerRegistryAuthenticator) getState(err error) v1alpha1.RegistryCredentialsState {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case "UnrecognizedClientException":
			return v1alpha1.RegistryCredentialsUnauthorized
		case "InvalidSignatureException":
			return v1alpha1.RegistryCredentialsUnauthorized
		default:
			return v1alpha1.RegistryCredentialsErrored
		}
	} else {
		return v1alpha1.RegistryCredentialsErrored
	}
}
