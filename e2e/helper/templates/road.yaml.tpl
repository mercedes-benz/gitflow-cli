# MB.OS ROAD.Kit CLI (Rapid Onboard Application Development) configuration file

# Configuration format.
# Older versions that were called "NCP-config.yaml" have version numbers prior to 1.1.0 (which is the first ROAD format)
ncpConfigFormat: 1.1.0

# The Version number you want to set in MB.OS Portal for your application.
# For development, you can specify any SemVer here ending with -dev. For production apps, omit the -dev.
# Please note, that production apps can only be pushed once in a specific version!
versionNumber: {{.Version}}

# The Schema for the ROAD manifest. Can be one of ncp (for Containerized Apps), starfish (Android),
# gen20x-service-extension, gen20x-native-app, onlineui or custom
schema: ncp

# The default geo for the app to be published
geo: emea

# The application type. Options are:
#   - headful (with UI - default)
#   - headless (without UI)
#   - systemService (special headless app on NTG7)
applicationType: headful

# The application section
app:
# Application name as specified in MB.OS Portal
name: MyAppNCP
# The Application UUID as it is generated from MB.OS Portal
uid: dc37b716-527e-45e0-8f7c-76db2fd5b288

# UI Section
# You can specify where the sources for your QML UI reside.
ui:
source: ./resources

# directory contains arbitrary resources to copied to the container image. This is optional.
resources:
# content of resources folder will be copied to /resources/ in container (for all arch)
- source: ./resources
target: /resources
permissions: "0755"
# following example shows distinction between third-party compiled resources for amd64 / arm64
- source: ./lib_arm/
target: /resources/mylib
arch: arm64
permissions: "0755"
- source: ./lib_amd/
target: /resources/mylib
arch: amd64
permissions: "0755"

# Executable
# Here, you need to specify an executable that has to be provided by your team's build process.
# For local development, you only need to specify an AMD64 binary, whereas for pushing the app, you also need to specify
# an executable built for ARM64 architecture.
executable:
amd64: ./app
arm64: ./app_arm64

# The application endpoints that can be invoked from the UI.
endpoints:
- id: string
path: string
port: int

# Dependencies section
dependencies:
# Here, you can specify all backend endpoints your application needs to call. They will be whitelisted.
backendEndpoints:
- id: string
url: string
name: string
description: string
countryOverrides:
- countryCode: string
url: string
# Here, you can specify all vehicle API ('VAS') endpoints you want to call. They will be whitelisted.
vehicleApiEndpoints:
- path: string
operations: [string]
# Here, you can specify all thrift ME API endpoints you want to call. They will be whitelisted.
thriftMeEndpoints:
- broker:
name: string
endpoints: [string]
services:
- serviceInstanceName: string
serviceType: string
serviceId: int
methods: [string]
events: [string]
# Here, you can specify all API endpoints from other containers you want to call. They will be whitelisted.
applicationEndpoints:
- id: string
path: string
operations: [string]

capabilities:
- name: string
canSignals:
receivableCANSignals: [string]
events:
- topic: string
hostPorts:
- id: string

# Volume mounts
# You can specify volumes that you want your container to use, e.g. for persistence or logging.
volumeMounts:
- name: string # e.g. storage
path: string # e.g. /usr/storage
namespace: string
readOnly: bool
subPath: string

# Container
# In this section, you can specify details of your container, for example environment variables you want
# to make available to the container. In minimum specify a tcp port you want to be exposed when running it.
container:
workingDir: string
ports:
- name: string
containerPort: int
protocol: string
env:
- name: string
value: string
# the following liveness and readiness checks will be AUTO-GENERATED into the podspec
# only use them if you need to overwrite the endpoint path, port or thresholds
livenessProbe:
httpGet:
path: /api/v1.0/appLifecycle/liveness
port: 8080
scheme: HTTP
initialDelaySeconds: 1
periodSeconds: 10
successThreshold: 1
timeoutSeconds: 2
failureThreshold: 3
readinessProbe:
httpGet:
path: /api/v1.0/appLifecycle/readiness
port: 8080
scheme: HTTP
initialDelaySeconds: 1
periodSeconds: 10
successThreshold: 1
timeoutSeconds: 2
failureThreshold: 3
# the following lifecycle endpoints will be AUTO-GENERATED into the podspec
# only use them if you need to overwrite the endpoint path or port
lifecycle:
postStart:
httpGet:
path: /api/v1.0/appLifecycle/postStart
port: 8080
scheme: HTTP
preStop:
httpGet:
path: /api/v1.0/appLifecycle/preStop
port: 8080
scheme: HTTP
lowMemory:
httpGet:
path: /api/v1.0/appLifecycle/lowMemory
port: 8080
scheme: HTTP
lowDisk:
httpGet:
path: /api/v1.0/appLifecycle/lowDisk
port: 8080
scheme: HTTP

ports:
- containerPort: 8080
protocol: tcp

restartPolicy: string
terminationGracePeriodSeconds: int
activeDeadlineSeconds: int

# Flag that indicates whether the application is able to handle
#  multiple users or whether a separate instance of the app has to be started per user. See
# [Container Development: Multi seat support](/develop/in-car-apps/development-guides/container-development/multi-seat)
isApplicationMultiTenant: bool

# Defines what connectivity to the Internet the application is going to use.
# Details are provided in [Container Development: Connectivity to the Internet](/develop/in-car-apps/development-guides/container-development/accessing-backend-apis#internet-connectivity).
connectivityType: string

metadata:
name: string
namespace: string
uid: string
annotations:
supportUserSwitching: bool
displayName: string
apkHash: string

---
# Here, you can specify the build variants for your configuration. Here, we only specified the geo so that
# rk-cli pushes the complete app to amap respective china, if you use those variants.
# During rk app push, you can address those variants with --variant variantName
# and similarly you can address geo with --geo geoName
# You can override ANY yaml key specified above with a variant-specific configuration. Please note that elements
# cannot be merged on a list level, meaning that lists have to repeated entirely in an override.
variant: nonprod
geo: amap
# e.g. executable is overridden here for amap nonprod
executable:
amd64: ./app_amap_nonprod
arm64: ./app_amap_nonprod_arm64
---
variant: nonprod
geo: china