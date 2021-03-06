# Helm Cache

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.16.0](https://img.shields.io/badge/AppVersion-1.16.0-informational?style=flat-square)

A service for caching charts from secrets in Kubernetes, save them localy and store on [Chartmuseum](https://github.com/helm/chartmuseum) (optionally).

## Prerequisites

- Helm >= 3
- Kubernetes >= 1.16

## Installing the chart

Without Chartmuseum cache:

```shell
$ helm upgrade --install --create-namespace \
    -n helm-cache \
    helm-cache \
    ./charts/helm-cache
```

With Chartmuseum cache:

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
| affinity | object | `{}` | Affinity for pod assignment. |
| chartmuseum.password | string | `""` | Chartmuseum password. |
| chartmuseum.url | string | `""` | Chartmuseum URL. |
| chartmuseum.username | string | `""` | Chartmuseum username. |
| fullnameOverride | string | `""` | String to fully override helm-cache.fullname template. |
| image.pullPolicy | string | `"IfNotPresent"` | helm-cache image pull policy. |
| image.repository | string | `"turboazot/helm-cache"` | helm-cache image repository. |
| image.tag | string | `"0.0.8"` | helm-cache image tag. |
| imagePullSecrets | list | `[]` | helm-cache image pull secrets. |
| nameOverride | string | `""` | String to partially override helm-cache.fullname template (will maintain the release name). |
| nodeSelector | object | `{}` | Node labels for pod assignment. Evaluated as a template. |
| podAnnotations | object | `{}` | Annotations for helm-cache pods. |
| podSecurityContext | object | `{}` | helm-cache pods' Security Context. |
| rbac.create | bool | `true` | Create RBAC resources. |
| resources | object | `{}` | The resources requests and limits for the helm-cache container. |
| scanningInterval | string | `"10s"` | An interval between scanning release secrets. |
| securityContext | object | `{}` | helm-cache security context. |
| serviceAccount.annotations | object | `{}` | Annotations for service account. |
| tolerations | list | `[]` | Tolerations for pod assignment. |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.7.0](https://github.com/norwoodj/helm-docs/releases/v1.7.0)
