# ECRCredentials

## Description

ECRCredentials represents the AWS Access Key to autenticate against an ECR registries into an AWS Region.

## Specification

| Property | Type | Required | Description |
| --- | --- | --- | --- |
| `.apiVersion` | `string` | yes | Defines the versioned schema of this object. |
| `.kind` | `string` | yes | ECRCredentials |

### .spec

| Property | Type | Required | Description |
| --- | --- | --- | --- |
| `accessKeyID` | `string` | yes | AWS Access Key ID |
| `secretAccessKey` | `string` | yes | AWS Secret Access Key |
| `region` | `string` | yes | AWS Region |
| `imageSelector` | `array (string)` | no | List of regexp to match images |


### .status

| Property | Type | Required | Description |
| --- | --- | --- | --- |
| `phase` | `string` | no | The current phase of the object: Authenticating, Aunthenticated, Unauthenticated, Error |
| `errorMessage` | `string` | no | The message returned when in Error phase |
