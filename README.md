# Prerequisites

Kubernetes 1.9.0 or above with the admissionregistration.k8s.io/v1beta1 API enabled. Verify that by the following command:

```
kubectl api-versions | grep admissionregistration.k8s.io/v1beta1
```
The result should be:
```
admissionregistration.k8s.io/v1beta1
```


# Deploy
0. Create namespace and rbad with kube-webhook
```ecma script level 4
kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/rbac.yaml
```


1. Create a signed cert/key pair and store it in a Kubernetes secret that will be consumed by deployment

```ecma script level 4
./deploy/webhook-create-cert.sh \
    --service kube-webhook-svc \
    --secret kube-webhook-certs \
    --namespace kube-webhook
```

2. Install deployment and service
```ecma script level 4
kubectl apply -f deploy/deployment.yaml
kubectl apply -f deploy/service.yaml
``` 

3. Patch the MutatingWebhookConfigurations by set caBundle with correct value from Kubernetes cluster

```ecma script level 4
cat deploy/mutating-webhook.yaml | \
    deploy/webhook-patch-ca-bundle.sh | \
    kubectl apply -f -
```
`kubectl get MutatingWebhookConfiguration`
