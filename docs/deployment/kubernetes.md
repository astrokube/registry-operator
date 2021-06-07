# Kubernetes

The Registry Operator can be deployed in any Kubernetes distribution.

## Kubectl

### Prerequisites

- A Kubernetes cluster
- kubectl installed installed in your machine

### Procedure

1. Apply the YAML manifests from the Github repository:

    ```sh
    kubectl apply -f https://...
    ```

## Helm

### Prerequisites

- A Kubernetes cluster
- kubectl installed installed in your machine
- Helm client installed in your machine

### Procedure

1. Add the Helm Repository:

    ```sh
    helm repo add astrokube https://astrokube.github.io/charts
    ```

2. Install the Chart:

    ```sh
    helm install registry-operator astrokube/registry-operator
    ```
