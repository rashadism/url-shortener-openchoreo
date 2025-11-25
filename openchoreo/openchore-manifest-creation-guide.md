# OpenChoreo Manifest Creation Guide

## Resource Creation Hierarchy
*Create resources in this order due to dependencies:*

1. **Platform Setup** (Platform Engineers)
   - Organization → DataPlane → BuildPlane → Environment → Classes → DeploymentPipeline

2. **Application Setup** (Developers)
   - Project → Component → Workload → Service/WebApplication/ScheduledTask

## Core Resources to Create

### Organization Resources

#### **Organization**
```yaml
apiVersion: core.openchoreo.io/v1alpha1
kind: Organization
metadata:
  name: <org-name>
```
**When**: First resource, one per tenant/company

#### **DataPlane**
```yaml
apiVersion: core.openchoreo.io/v1alpha1
kind: DataPlane
metadata:
  name: <cluster-name>
  namespace: openchoreo-system
spec:
  organizationRef: <org-name>
  connection:
    kubeconfig: <base64-encoded-kubeconfig>
```
**When**: For each Kubernetes cluster to deploy to

#### **BuildPlane**
```yaml
apiVersion: core.openchoreo.io/v1alpha1
kind: BuildPlane
metadata:
  name: <build-cluster-name>
  namespace: openchoreo-system
spec:
  organizationRef: <org-name>
  connection:
    kubeconfig: <base64-encoded-kubeconfig>
```
**When**: For CI/CD execution (optional if using external CI)

#### **Environment**
```yaml
apiVersion: core.openchoreo.io/v1alpha1
kind: Environment
metadata:
  name: <env-name>  # e.g., dev, staging, prod
  namespace: openchoreo-<org-name>
spec:
  organizationRef: <org-name>
  dataPlaneRef: <dataplane-name>
```
**When**: For each deployment stage

### Platform Classes

#### **ServiceClass**
```yaml
apiVersion: core.openchoreo.io/v1alpha1
kind: ServiceClass
metadata:
  name: <class-name>  # e.g., standard, high-performance
  namespace: openchoreo-<org-name>
spec:
  organizationRef: <org-name>
  resources:
    requests:
      memory: "256Mi"
      cpu: "100m"
    limits:
      memory: "512Mi"
      cpu: "500m"
  scaling:
    minReplicas: 2
    maxReplicas: 10
```
**When**: Define standards for backend services

#### **WebApplicationClass**
```yaml
apiVersion: core.openchoreo.io/v1alpha1
kind: WebApplicationClass
metadata:
  name: <class-name>  # e.g., static-site, spa
  namespace: openchoreo-<org-name>
spec:
  organizationRef: <org-name>
  resources:
    requests:
      memory: "128Mi"
      cpu: "50m"
```
**When**: Define standards for frontend applications

#### **ScheduledTaskClass**
```yaml
apiVersion: core.openchoreo.io/v1alpha1
kind: ScheduledTaskClass
metadata:
  name: <class-name>  # e.g., batch-job
  namespace: openchoreo-<org-name>
spec:
  organizationRef: <org-name>
  resources:
    limits:
      memory: "1Gi"
      cpu: "1000m"
```
**When**: Define standards for cron jobs

### Application Resources

#### **Project**
```yaml
apiVersion: apps.openchoreo.io/v1alpha1
kind: Project
metadata:
  name: <project-name>
  namespace: openchoreo-<org-name>
spec:
  organizationRef: <org-name>
  description: "Business domain/team boundary"
```
**When**: Group related components (per team/domain)

#### **Component**
```yaml
apiVersion: apps.openchoreo.io/v1alpha1
kind: Component
metadata:
  name: <component-name>
  namespace: openchoreo-<org-name>-<project-name>
spec:
  projectRef: <project-name>
  organizationRef: <org-name>
  type: Service  # or WebApplication, ScheduledTask
  source:
    repository: https://github.com/org/repo
    branch: main
    path: /services/api  # optional
  build:
    strategy: buildpack  # or dockerfile
    dockerfilePath: Dockerfile  # if strategy is dockerfile
```
**When**: For each deployable unit

#### **Workload**
```yaml
apiVersion: apps.openchoreo.io/v1alpha1
kind: Workload
metadata:
  name: <component-name>-workload
  namespace: openchoreo-<org-name>-<project-name>
spec:
  componentRef: <component-name>
  containers:
    - name: main
      image: <will-be-set-by-build>
      env:
        - name: ENV_VAR
          value: "value"
  endpoints:
    - name: http
      port: 8080
      protocol: HTTP
      visibility: Project  # or Organization, External
  connections:
    - name: database
      service: postgres-service
      namespace: openchoreo-<org-name>-<project-name>
      env:
        - name: DB_HOST
          valueFrom: host
        - name: DB_PORT
          valueFrom: port
```
**When**: Define runtime requirements for component

#### **Service** (The Claim)
```yaml
apiVersion: apps.openchoreo.io/v1alpha1
kind: Service
metadata:
  name: <service-name>
  namespace: openchoreo-<org-name>-<project-name>
spec:
  componentRef: <component-name>
  classRef: <service-class-name>
  workloadRef: <workload-name>
```
**When**: Deploy a backend service

#### **WebApplication** (The Claim)
```yaml
apiVersion: apps.openchoreo.io/v1alpha1
kind: WebApplication
metadata:
  name: <webapp-name>
  namespace: openchoreo-<org-name>-<project-name>
spec:
  componentRef: <component-name>
  classRef: <webapp-class-name>
  workloadRef: <workload-name>
  route:
    host: app.example.com
    path: /
```
**When**: Deploy a frontend application

#### **ScheduledTask** (The Claim)
```yaml
apiVersion: apps.openchoreo.io/v1alpha1
kind: ScheduledTask
metadata:
  name: <task-name>
  namespace: openchoreo-<org-name>-<project-name>
spec:
  componentRef: <component-name>
  classRef: <task-class-name>
  workloadRef: <workload-name>
  schedule: "0 2 * * *"  # Cron expression
```
**When**: Deploy a scheduled job

### Deployment Configuration

#### **DeploymentPipeline**
```yaml
apiVersion: core.openchoreo.io/v1alpha1
kind: DeploymentPipeline
metadata:
  name: <pipeline-name>
  namespace: openchoreo-<org-name>
spec:
  organizationRef: <org-name>
  stages:
    - name: dev
      environmentRef: dev
      autoPromote: true
    - name: staging
      environmentRef: staging
      approvals:
        required: true
    - name: prod
      environmentRef: prod
      approvals:
        required: true
        approvers: ["platform-team"]
```
**When**: Define promotion flow between environments

## Manifest Creation Rules

### Naming Conventions
- **Namespaces**: `openchoreo-<org>-<project>-<env>`
- **Resources**: Use kebab-case
- **References**: Must match exactly

### Required Fields by Resource
- **All resources**: `metadata.name`, `metadata.namespace`
- **Organization resources**: `spec.organizationRef`
- **Project resources**: `spec.projectRef`, `spec.organizationRef`
- **Component resources**: `spec.componentRef`, `spec.projectRef`

### Component Type Determines Resources
- **Service** → Create Service + ServiceClass
- **WebApplication** → Create WebApplication + WebApplicationClass (add `spec.route`)
- **ScheduledTask** → Create ScheduledTask + ScheduledTaskClass (add `spec.schedule`)

### Connection Patterns
```yaml
# Internal (same project)
connections:
  - name: auth-service
    service: auth
    namespace: openchoreo-<org>-<project>

# Cross-project (explicit)
connections:
  - name: shared-db
    service: postgres
    namespace: openchoreo-<org>-shared-services
    crossProject: true

# External
connections:
  - name: external-api
    url: https://api.external.com
    type: external
```

## Creation Workflow

1. **Platform Setup** (once per organization):
   - Create Organization
   - Register DataPlanes/BuildPlanes
   - Create Environments (dev, staging, prod)
   - Define Classes (ServiceClass, WebApplicationClass, etc.)
   - Setup DeploymentPipeline

2. **Per Application**:
   - Create Project
   - For each deployable unit:
     - Create Component (defines source/build)
     - Create Workload (defines runtime needs)
     - Create Service/WebApplication/ScheduledTask (references class)

3. **The Platform Automatically**:
   - Creates Bindings per environment
   - Triggers Builds
   - Generates Releases
   - Deploys to DataPlanes

## Quick Decision Tree

**What type of component?**
- Exposes API/handles requests → Service
- Serves UI/frontend → WebApplication  
- Runs on schedule → ScheduledTask

**What visibility for endpoints?**
- Only within project → `visibility: Project`
- Across projects → `visibility: Organization`
- Public internet → `visibility: External`

**What build strategy?**
- Auto-detect with buildpacks → `strategy: buildpack`
- Custom Dockerfile → `strategy: dockerfile`

**Environment-specific config?**
- Use Bindings (auto-created) with environment overrides
- Don't modify the base Service/WebApplication/ScheduledTask