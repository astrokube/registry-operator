# RegistryCredentials

## Description

RegistryCredentials represents the credentials required to authenticate to a remote Container Registry.

## Specification

| Property | Type | Required | Description |
| --- | --- | --- | --- |
| `.apiVersion` | `string` | yes | Defines the versioned schema of this object. |
| `.kind` | `string` | yes | RegistryCredentials |

### .spec


| Property | Type | Required | Description |
| --- | --- | --- | --- |
| `provider` | `object` | yes | The provider object |
| `imageSelector` | `object` | no | List of regexp to match images |

## .spec.awsElasticContainerRegistry

| Property | Type | Required | Description |
| --- | --- | --- | --- |
| `accessKeyID` | `string` | yes | AWS Access Key ID |
| `secretAccessKey` | `string` | yes | AWS Secret Access Key |
| `region` | `string` | yes | AWS Region |

## .spec.imageSelector

| Property | Type | Required | Description |
| --- | --- | --- | --- |
| `matchRegexp` | `array (string)` | no | Comma separated list of regexp to match container images. |
| `matchEquals` | `array (string)` | no | Comma separated list of images to match container images. |


### .status

| Property | Type | Required | Description |
| --- | --- | --- | --- |
| `state` | `string` | no | The current state of the object: Authenticating, Aunthenticated, Unauthorized, Errored, Terminating. |
| `errorMessage` | `string` | no | The message returned when in Error phase |
| `expirationTime` | `time` | no | The expiration time. |
| `authenticatedTime` | `time` | no | The authenticated time. |
