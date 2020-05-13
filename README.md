# Litmus-Admission-Controllers

Litmus Admission Webhook is an extension of LitmusChaos Chaos-Operator. This Webhook helps in validating the Chaos target specified in ChaosEngines.
It will be enhanced for validating the secondary resources such as Application related details, RBAC's, ChaosExperiments, Annotations and much more.


## Validation performed

Currently this Webhook validates the application information provided in chaoengine(`.spec.appInfo`).
    - It validates if the resource searched for exists in the k8s cluster or not.


## Installation Steps:

- Using current `litmus` ServiceAccount provided for chaos-operator (https://docs.litmuschaos.io/docs/getstarted/#install-litmus) will work for this deployment, or else a ServiceAccount linked with a ClusterRole with the following permission will work.
Prefer this deployment to run in parallel with chaos-operator.

```
rules:
  - apiGroups: ["","apps","litmuschaos.io"]
    resources: ["pods","jobs","deployments","replicationcontrollers","daemonsets","replicasets","statefulsets","chaosengines"]
    verbs: ["get","create","update","patch","delete","list","watch","deletecollection"]
  - apiGroups: ["admissionregistration.k8s.io"]
    resources: ["validatingwebhookconfigurations"]
    verbs: ["get","create","list","delete","update"]
```
 
- Deploy the `./litmus-admission-controller.yaml` or copy it for changing the ServiceAccount and Namespace.


## Sample ValidatingWebhookConfigration created 
The ValidatingWebhookConfiguration of this webhook would look something like:

```
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  creationTimestamp: "2020-05-05T10:00:52Z"
  generation: 1
  labels:
    app: admission-controller
    litmuschaos.io/component-name: admission-controller
    litmuschaos.io/version: 1.3.0
  name: litmuschaos-validation-webhook-cfg
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: Deployment
    name: litmus-admission-controllers
    uid: 71efe590-1432-4dfc-8f61-628d5bd22b44
  resourceVersion: "4219155"
  selfLink: /apis/admissionregistration.k8s.io/v1/validatingwebhookconfigurations/litmuschaos-validation-webhook-cfg
  uid: c685bc76-ff36-44f4-8e16-6815ae3a7186
webhooks:
- admissionReviewVersions:
  - v1beta1
  clientConfig:
    caBundle: ...
    service:
      name: admission-controller-svc
      namespace: litmus
      path: /validate
      port: 443
  failurePolicy: Ignore
  matchPolicy: Exact
  name: admission-controller.litmuschaos.io
  namespaceSelector: {}
  objectSelector: {}
  rules:
  - apiGroups:
    - litmuschaos.io
    apiVersions:
    - '*'
    operations:
    - CREATE
    resources:
    - chaosengines
    scope: '*'
  sideEffects: Unknown
  timeoutSeconds: 5


```
