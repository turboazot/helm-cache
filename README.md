# Helm Cache

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.16.0](https://img.shields.io/badge/AppVersion-1.16.0-informational?style=flat-square)

A Helm chart for caching charts from secrets in Kubernetes and save them in your chartmuseum

## Prerequisites

- Helm >= 3
- Kubernetes >= 1.16

## Installing the chart

```shell
$ helm upgrade --install --create-namespace \
    -n helm-cache \
    helm-cache \
    --set chartmuseum.url=http://chartmuseum-url:8080 \
    --set chartmuseum.username=chartmuseum \
    --set chartmuseum.password=chartmuseum \
    ./charts/helm-cache
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| chartmuseum.password | string | `""` |  |
| chartmuseum.url | string | `""` |  |
| chartmuseum.username | string | `""` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"turboazot/helm-cache"` |  |
| image.tag | string | `"0.0.2"` |  |
| imagePullSecrets | list | `[]` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| rbac.create | bool | `true` |  |
| resources | object | `{}` |  |
| securityContext | object | `{}` |  |
| serviceAccount.annotations | object | `{}` |  |
| tolerations | list | `[]` |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.7.0](https://github.com/norwoodj/helm-docs/releases/v1.7.0)
