configFormat: 1.0.0
versionNumber: {{.Version}}
schema: example-schema
geo: region-1
applicationType: standard
app:
  name: ExampleApp
  uid: 00000000-0000-0000-0000-000000000000

ui:
  source: ./ui

resources:
  - source: ./example-resources
    target: /app-resources
    permissions: "0755"
  - source: ./lib-xyz
    target: /app-resources/lib
    arch: arm64
    permissions: "0755"

---
variant: test
geo: region-2
executable:
  amd64: ./example-app-region2
  arm64: ./example-app-region2-arm64

---
variant: test
geo: region-3
