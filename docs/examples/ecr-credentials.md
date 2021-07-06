# RegistryCredentials

## Simple

```yaml
apiVersion: registry.astrokube.com/v1alpha1
kind: RegistryCredentials
metadata:
  name: sample
spec:
  provider:
    awsElasticContainerRegistry:
      accessKeyId: XXXXXXXXXXXXXXXXXXXX
      secretAccessKey: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
      region: eu-central-1
```

## With imageSelector

```yaml
apiVersion: registry.astrokube.com/v1alpha1
kind: RegistryCredentials
metadata:
  name: sample
spec:
  provider:
    awsElasticContainerRegistry:
      accessKeyId: XXXXXXXXXXXXXXXXXXXX
      secretAccessKey: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
      region: eu-central-1
  imageSelector:
    - 111111111111.dkr.ecr.eu-central-1.amazonaws.com/myimage:.*
```
