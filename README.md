# Litmus-Admission-Controllers

Litmus Admission Webhook is an extension of LitmusChaos Chaos-Operator. It helps in validating the chaos intent specified in the chaos custom resources (ChaosExperiments, ChaosEngines & ChaosSchedules).

As of now, this controller helps in validating existence of the application under test (AUT). In subsequent releases, it will be enhanced to perform increased validation of chaos inputs and environmental dependencies, thereby offloading these functions from the chaos-operator/runner. 

## Installation Steps:

- Using current `litmus` ServiceAccount provided for chaos-operator (https://docs.litmuschaos.io/docs/getstarted/#install-litmus) will work for this deployment, or else a ServiceAccount linked with a ClusterRole with the following permission will work.
Prefer this deployment to run in parallel with chaos-operator.
 
- Deploy the `./litmus-admission-controller.yaml` or use the raw link of that YAML.


### Example Validation

- Reference for a Dummy ChaosEngine:
```
apiVersion: litmuschaos.io/v1alpha1
kind: ChaosEngine
metadata:
  name: engine
  namespace: litmus
spec:
  monitoring: true
  appinfo:
    appkind: deployment
    applabel: app=nginx
    appns: litmus
  chaosServiceAccount: litmus
  experiments:
  - name: pod-delete

```

- And, the fields in `.spec.appInfo` specify that a deployment labelled `app=nginx` should exists in `litmus` namespace

- For a failure case, lets assume that this type of deployment does'nt exist. So the response of admission controller, would be something like:
```
rahul@rahul-ThinkPad-E490:~$ kubectl apply -f chaos-engine.yaml 
Error from server (BadRequest): error when creating "chaos-engine.yaml": admission webhook "admission-controller.litmuschaos.io" denied the request: unable to find deployment specified in ChaosEngine
```

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
