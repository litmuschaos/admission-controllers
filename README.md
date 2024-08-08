# Litmus-Admission-Controller

## Overview

The Litmus Admission Controller is designed to validate and restrict the usage of the TargetServiceAccount (typically  `litmus-admin`) within a Kubernetes cluster.
It ensures that only specific pods can use the target service account based on predefined criteria. 
This is crucial for maintaining the security and proper governance of the cluster.

## Features

- **Service Account Validation**: Only pods originating from specific service accounts can use the target service account.
- **Origin Pod Image Validation**: Only pods with specific images can launch using the target service account.
- **Target Pod Image Validation**: Restrict the images that can be used by pods utilizing the target service account.
- **TLS Certificate Management**: The admission controller supports both self-managed and user-provided TLS certificates for secure communication.

## Configuration Options

### 1. AllowedOriginServiceAccounts

A list of service accounts that are allowed to launch pods with the target service account. This restricts the usage of the target service account to a controlled set of origins.

```yaml
- name: ALLOWED_ORIGIN_SERVICE_ACCOUNTS
  value: '["litmus-admin","replicaset-controller"]'
```

### 2. AllowedOriginImages

A list of container images from which the origin pods are allowed to use the target service account. This ensures that only pods with trusted and verified images can use this privileged account.

```yaml
- name: ALLOWED_ORIGIN_IMAGES
  value: '["^.*/litmuschaos/go-runner:.*$","^.*/litmuschaos/chaos-operator:.*$"]'
```

### 3. AllowedTargetImages

A list of container images that can be used by pods running with the target service account. This helps in ensuring that only vetted images are run with elevated privileges.

```yaml
- name: ALLOWED_TARGET_IMAGES
  value: '["^.*/litmuschaos/go-runner:.*$"]'
```

### 4. Self-Managed Dependencies

The `SELF_MANAGED_DEPENDENCIES` environment variable controls how the TLS certificates and `ValidatingWebhookConfiguration` are managed:

- **Default (`false`)**: The user is expected to provide and manage the TLS certificates and create the `ValidatingWebhookConfiguration`.
- **Set to `true`**: The admission controller will manage its own TLS certificates and create the recommended `ValidatingWebhookConfiguration`.

```yaml
- name: SELF_MANAGED_DEPENDENCIES
  value: "false"
```

### TLS Certificate Management

If `SELF_MANAGED_DEPENDENCIES` is set to `false`, you must manually provide the TLS certificates:

1. **Create a Kubernetes Secret**: Create a secret containing the TLS certificate and key.

    ```bash
    kubectl create secret tls admission-server-tls \
      --cert=/path/to/tls.crt \
      --key=/path/to/tls.key \
      -n <namespace>
    ```

2. **Mount the Secret**: Mount the secret into the admission controller at the path `/etc/certs`.

## ValidatingWebhookConfiguration

To enforce these restrictions, a `ValidatingWebhookConfiguration` is created. This webhook intercepts pod creation requests and validates them against the configured rules.

### Sample ValidatingWebhookConfiguration

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: pod-validation
webhooks:
  - name: pod-validation.default.svc
    namespaceSelector:
      matchExpressions:
        - key: kubernetes.io/metadata.name
          operator: In
          values: [ "controller" ]
    clientConfig:
      service:
        name: admission-server
        namespace: default
        path: "/validate/pods"
      caBundle: "${CA_BUNDLE}"
    rules:
      - operations: ["CREATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
        scope: "Namespaced"
    admissionReviewVersions: ["v1"]
    sideEffects: None
```

### Explanation:

- **Namespace Selector**: The webhook is configured to operate only within specific namespaces, as defined by the `matchExpressions`.
- **Client Config**: Specifies the admission server service and the path to be used for validation.
- **Rules**: The webhook only triggers on `CREATE` operations for pod resources within a namespace.
- **Admission Review Versions**: Specifies the admission review versions supported by the webhook.
- **Side Effects**: Indicates that the webhook does not have side effects.

## Deployment

### For Self-Managed Dependencies

1. **Set `SELF_MANAGED_DEPENDENCIES` to `true`**: In your admission controller deployment, set the environment variable to `true`.
2. **Deploy the Admission Controller**: Deploy the admission controller and associated service in the desired namespace. It will automatically manage TLS certificates and create the `ValidatingWebhookConfiguration`.

### For User-Provided Dependencies

1. **Set `SELF_MANAGED_DEPENDENCIES` to `false`**: In your admission controller deployment, ensure that the environment variable is set to `false`.
2. **Create and Mount TLS Secret**: Create a Kubernetes secret with your TLS certificates and mount it at `/etc/certs` in the admission controller deployment.
3. **Deploy the Admission Controller**: Deploy the admission controller and associated service in the desired namespace.
4. **Create the ValidatingWebhookConfiguration**: Manually apply the `ValidatingWebhookConfiguration` YAML after the admission controller is running.

## Conclusion

The Litmus Admission Controller provides a robust mechanism to control and secure the usage of the target service account, ensuring that only authorized pods can utilize it. With flexible TLS certificate management options, it offers a secure and customizable solution for your Kubernetes cluster.