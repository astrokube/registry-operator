# Integrate AWS ECR

You can integrate AWS ECR in a single Namespace by creating a RegistryCredentials object.


## Prerequisites

- You have already created an AWS Access Key with Read access level to Elastic Container Registry. Check the [IAM Policy details](aws-ecr-policy.md).

### Procedure

1. Create the following YAML manifest. In this example the file si called _ecr-credentials.yaml_:

    ```yaml
    apiVersion: registry.astrokube.com/v1alpha1
    kind: RegistryCredentials
    metadata:
      name: sample
    spec:
      accessKeyId: XXXXXXXXXXXXXXXXXXXX
      secretAccessKey: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
      region: eu-central-1
    ```

2. Apply the configuration to create the RegistryCredentials object:

    ```sh
    kubectl apply -f ecr-credentials.yaml
    ```

3. Verify the RegistryCredentials is authenticated:

    ```sh
    kubectl get registrycredentials
    ```

    _Example output_

    ```sh
    NAME     STATUS
    sample   Authenticated
    ```

    > NOTE: 
    > The status will be set as Unauthorized if your put the wrong AWS Access Key or if it hasn't the correct IAM Policy attached.

4. Once the RegistryCredentials object is Authenticated, a new secret will be created and can be used for the ImagePullPolicy.

    You can verify the secret has been created with the following command:

    ```sh
    kubectl get secrets sample
    ```

    _Example output_

    ```sh
    NAME     TYPE                             DATA   AGE
    sample   kubernetes.io/dockerconfigjson   1      5s
    ```
