# Registry Controller

## Overview

**Registry Controller** lets you integrate external container registries in your Kubernetes cluster.

## CustomResourceDefinitions

The core feature of the Registry Controller is to manage the DockerConfig credentials in your Namespaces with the usage of CustomResourcesDefinitions (CRD).

Those are the implemented CRD:

* ECRCredentials: an object to store the DockerConfig credentials for AWS ECR.
