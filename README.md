# Litmus-Admission-Webhook

Litmus Admission Webhook is an extension of LitmusChaos Chaos-Operator. This Webhook helps in validating the Chaos target specified in ChaosEngines.
It will be enhanced for validating the secondary resources such as Application related details, RBAC's, ChaosExperiments, Annotations and much more.

In the current PWD, the admission.yaml could be deployed to see it work. Rightnow, it scope is just to add a log, and validate if the ChaosEngine's AppInfo,AppNamespace is `litmus` or not. But the scope will increase drastically.

The ValidatingWebhookConfiguration of this webhook would look something like:

```
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  generation: 1
  labels:
    app: admission-webhook
    litmuschaos.io/component-name: admission-webhook
    litmuschaos.io/version: v1.3.0
  name: litmuschaos-validation-webhook-cfg
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: Deployment
    name: litmus-admission-server
  resourceVersion: "3389002"
  selfLink: /apis/admissionregistration.k8s.io/v1/validatingwebhookconfigurations/litmuschaos-validation-webhook-cfg
webhooks:
- admissionReviewVersions:
  - v1beta1
  clientConfig:
    caBundle: ...
    service:
      name: admission-server-svc
      namespace: litmus
      path: /validate
      port: 443
  failurePolicy: Ignore
  matchPolicy: Exact
  name: admission-webhook.litmuschaos.io
  namespaceSelector: {}
  objectSelector: {}
  rules:
  - apiGroups:
    - '*'
    apiVersions:
    - '*'
    operations:
    - CREATE
    - DELETE
    resources:
    - chaosengines
    scope: '*'
  sideEffects: Unknown
  timeoutSeconds: 5

```
