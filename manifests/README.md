# Kustomize Configuration

This directory uses Kustomize to manage Kubernetes manifests with configurable variables. All hardcoded values have been moved to `vars.yaml` for easy configuration.

## Structure

```
manifests/
├── kustomization.yaml      # Main Kustomize configuration
├── vars.yaml              # All configurable variables (EDIT THIS FILE)
├── base/                  # Base manifests
│   ├── url-shortener-demo-project.yaml
│   ├── postgres-component.yaml
│   ├── redis-component.yaml
│   ├── api-service-component.yaml
│   ├── analytics-service-component.yaml
│   └── frontend-component.yaml
└── README.md             # This file
```

## Quick Start

### 1. Customize Variables

Edit `vars.yaml` to change any configuration values:

```yaml
data:
  # Change image tags
  IMAGE_TAG: "v1.0.0"

  # Change database credentials
  DB_PASSWORD: "your-secure-password"

  # Change resource limits
  API_MEMORY_LIMIT: "512Mi"

  # ... and more
```

### 2. Build Manifests

Preview the generated manifests:

```bash
kubectl kustomize manifests/
```

Or using standalone kustomize:

```bash
kustomize build manifests/
```

### 3. Deploy to Kubernetes

Deploy directly:

```bash
kubectl apply -k manifests/
```

Or build and save to a file:

```bash
kubectl kustomize manifests/ > deployment.yaml
kubectl apply -f deployment.yaml
```

## What Can You Configure?

All variables in `vars.yaml` can be customized:

### Project Configuration
- `PROJECT_NAME`: Project name used across all components
- `NAMESPACE`: Kubernetes namespace
- `ENVIRONMENT`: Environment name (development, production, etc.)

### Image Configuration
- `IMAGE_REGISTRY`: Docker registry for custom images
- `IMAGE_TAG`: Tag for all custom images (api, analytics, frontend)
- `POSTGRES_IMAGE`: PostgreSQL image and tag
- `REDIS_IMAGE`: Redis image and tag

### Service Configuration
- `API_SERVICE_PORT`: API service container port
- `ANALYTICS_SERVICE_PORT`: Analytics service container port
- `FRONTEND_PORT`: Frontend container port
- `DB_PORT`: PostgreSQL port
- `REDIS_PORT`: Redis port
- `*_REPLICAS`: Number of replicas for each service

### Database Configuration
- `DB_USER`: PostgreSQL username
- `DB_PASSWORD`: PostgreSQL password
- `DB_NAME`: PostgreSQL database name

### Application Configuration
- `RATE_LIMIT_REQUESTS`: API rate limit requests per window
- `RATE_LIMIT_WINDOW`: API rate limit window in seconds

### Resource Limits
For each service (API, Analytics, Frontend, PostgreSQL, Redis):
- `*_CPU_REQUEST`: CPU request
- `*_MEMORY_REQUEST`: Memory request
- `*_CPU_LIMIT`: CPU limit
- `*_MEMORY_LIMIT`: Memory limit

### OpenChoreo Configuration
- `SERVICE_EXPOSE_PORT`: Port on which services are exposed (default: 80)

## Common Use Cases

### Change Image Tag for All Services

Edit `vars.yaml`:
```yaml
data:
  IMAGE_TAG: "v2.0.0"
```

Then apply:
```bash
kubectl apply -k manifests/
```

### Update Database Password

Edit `vars.yaml`:
```yaml
data:
  DB_PASSWORD: "new-secure-password-123"
```

**Note**: This also updates the DATABASE_URL in api-service and analytics-service automatically.

### Change Namespace

Edit `vars.yaml`:
```yaml
data:
  NAMESPACE: "production"
```

This updates the namespace in all components.

### Increase Resources for API Service

Edit `vars.yaml`:
```yaml
data:
  API_CPU_LIMIT: "500m"
  API_MEMORY_LIMIT: "512Mi"
```

### Scale Services

Edit `vars.yaml`:
```yaml
data:
  API_SERVICE_REPLICAS: "3"
  ANALYTICS_SERVICE_REPLICAS: "2"
```

## Advanced Usage

### Using Different Image Registries

Edit `vars.yaml` to change the registry:
```yaml
data:
  IMAGE_REGISTRY: "myregistry.azurecr.io"
```

Then update the `images` section in `kustomization.yaml` to use this variable:
```yaml
images:
  - name: rashadxyz/url-shortener-api
    newName: myregistry.azurecr.io/url-shortener-api
    newTag: demo
```

### Customizing Database Connection Strings

The database connection URLs are constructed in `kustomization.yaml` patches section:

```yaml
patches:
  - target:
      kind: Workload
      name: api-service
    patch: |-
      - op: replace
        path: /spec/containers/main/env/1/value
        value: "postgres://urlshortener:password123@postgres:80/urlshortener?sslmode=disable"
```

Edit these patches to customize the connection strings with your vars.yaml values.

## Troubleshooting

### Verify Generated Manifests

Always preview before applying:
```bash
kubectl kustomize manifests/ | less
```

### Check Specific Component

```bash
kubectl kustomize manifests/ | grep -A 20 "kind: Workload" | grep -A 20 "name: api-service"
```

### Validate Syntax

```bash
kubectl kustomize manifests/ > /tmp/test.yaml
kubectl apply --dry-run=client -f /tmp/test.yaml
```

## Migration from Old Manifests

The old manifests in `base/` directory are now templates. All customization should be done in `vars.yaml` instead of editing the base manifests directly.

Before Kustomize:
- Edit multiple files to change image tags
- Search/replace for hardcoded values
- Risk of missing updates in some files

After Kustomize:
- Edit one file (`vars.yaml`)
- All manifests updated automatically
- Consistent configuration across all components
