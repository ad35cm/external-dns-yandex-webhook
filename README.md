# ExternalDNS - Yandex Cloud DNS Webhook

This is an [ExternalDNS provider](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/webhook-provider.md) for [Yandex Cloud DNS](https://yandex.cloud/en/services/dns).
It externalizes the Yandex Cloud DNS provider and offers a path for independent fixes and releases.

## Installation

The webhook is designed to run as a sidecar in the `external-dns` pod. The recommended authentication method on Yandex Managed Service for Kubernetes is [workload identity federation](https://yandex.cloud/en/docs/iam/concepts/workload-identity), which does not require a JSON key or Kubernetes Secret.

Configure the official ExternalDNS Helm chart with the `webhook` provider and annotate its Kubernetes service account with the IAM service account ID:

```yaml
serviceAccount:
  create: true
  name: external-dns
  annotations:
    yandex.cloud/federated-yc-service-account-id: YOUR_IAM_SERVICE_ACCOUNT_ID

provider:
  name: webhook
  webhook:
    image:
      repository: ghcr.io/ad35cm/external-dns-yandex-webhook
      tag: 1.1.0
    args:
      - --folder-id=YOUR_FOLDER_ID
    resources: {}
    securityContext:
      runAsUser: 1000
```

Do not set `--auth-key-file` when using workload identity. The webhook will obtain short-lived IAM tokens from the metadata endpoint provided to the annotated pod by Yandex Managed Service for Kubernetes.

## Workload identity setup

Before installing the chart:

1. Enable workload identity federation for both the Managed Service for Kubernetes cluster and its node group.
2. Create an IAM service account and grant it the `dns.editor` role for the folder containing the DNS zones.
3. Create a workload identity federation using the cluster's issuer URL, JWKS URL, and issuer URL as the audience.
4. Link the IAM service account to the federation. The external subject must match the Kubernetes service account:

   ```text
   system:serviceaccount:<namespace>:<kubernetes-service-account-name>
   ```

5. Set `serviceAccount.annotations.yandex.cloud/federated-yc-service-account-id` in the Helm values to the IAM service account ID.

For the complete platform setup, see Yandex Cloud's guide to [accessing the Yandex Cloud API from Managed Service for Kubernetes using workload identity federation](https://yandex.cloud/en/docs/iam/tutorials/wlif-managed-k8s-integration).

## Authentication behavior

Authentication is selected from the webhook configuration:

- If `--auth-key-file` is omitted or empty, the webhook uses Yandex Cloud workload identity through the metadata service.
- If `--auth-key-file` is set, the webhook reads the service account JSON key from that path. This remains available for backwards compatibility and environments without workload identity.

### JSON key fallback

To use the legacy key-file method, create an authorized key and store it in a Kubernetes Secret:

```shell
yc iam key create \
  --service-account-id=YOUR_IAM_SERVICE_ACCOUNT_ID \
  --format=json \
  --output=key.json

kubectl create secret generic yandexconfig \
  --namespace external-dns \
  --from-file=key.json
```

Then add the key argument and Secret volume to the Helm values:

```yaml
provider:
  name: webhook
  webhook:
    args:
      - --folder-id=YOUR_FOLDER_ID
      - --auth-key-file=/etc/kubernetes/key.json
    extraVolumeMounts:
      - name: yandexconfig
        mountPath: /etc/kubernetes
        readOnly: true

extraVolumes:
  - name: yandexconfig
    secret:
      secretName: yandexconfig
```

## Command-line arguments

- `--folder-id`: Required. Yandex Cloud folder ID containing the DNS zones.
- `--auth-key-file`: Optional. Path to a Yandex Cloud service account JSON key. If omitted, workload identity is used.
- `--webhook-port`: Webhook server port. Defaults to `8888`.
- `--health-port`: Health check server port. Defaults to `8080`.
