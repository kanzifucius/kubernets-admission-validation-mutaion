# Kubernetes Admission Webhook


## Prerequisites

Kubernetes 1.9.0 or above with the `admissionregistration.k8s.io/v1beta1` API enabled. Verify that by the following command:
```
kubectl api-versions | grep admissionregistration.k8s.io/v1beta1
```
The result should be:
```
admissionregistration.k8s.io/v1beta1
```

In addition, the `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook` admission controllers should be added and listed in the correct order in the admission-control flag of kube-apiserver.

## Build

1. Setup dep

   The repo uses [dep](https://github.com/golang/dep) as the dependency management tool for its Go codebase. Install `dep` by the following command:
```
go get -u github.com/golang/dep/cmd/dep
```

2. Build and push docker image
   
```
./build.sh
```

3. Install with Helm

```
helm  install --name=myhook .
```


## How does it work?

 - Will exectuire validation on all deployments and services on a namespace label with "admission-webhook: enabled"

eg:
```
kubectl label namespace temptest admission-webhook=enabled
kubectl apply -n temptest -f ./testdeployments/sleep.yaml
Error from server: error when creating "../deployment/sleep.yaml": admission webhook "test-admissionwebhook.vodacom.co.za" denied the request: required labels  are not set za.co.vodacom/team
```
