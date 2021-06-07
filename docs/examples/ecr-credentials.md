# ECRCredentials

## Simple

```yaml
apiVersion: registry.astrokube.com/v1alpha1
kind: ECRCredentials
metadata:
  name: sample
spec:
  accessKeyId: XXXXXXXXXXXXXXXXXXXX
  secretAccessKey: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
  region: eu-central-1
```

## With imageSelector

```yaml
apiVersion: registry.astrokube.com/v1alpha1
kind: ECRCredentials
metadata:
  name: sample
spec:
  accessKeyId: XXXXXXXXXXXXXXXXXXXX
  secretAccessKey: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
  region: eu-central-1
  imageSelector:
    - 921780870478.dkr.ecr.eu-central-1.amazonaws.com/myimage:.*
```
