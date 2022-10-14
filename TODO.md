# TODO items

## DBaaS controllers

### Cluster CR controller
- [x] secondary resources finalizer
- [x] CR delete handling
  - [x] delete secondary resources
  - [x] CR spec.terminationPolicy handling
- [x] managed resources handling
  - [x] roleGroup attached Service kind
  - [x] deployment workloads
  - [x] PDB
  - [x] label handling:
    - [x] deploy & sts workloads's labels and spec.template.metadata.labels (check https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/)
- [x] immutable spec properties handling (via validating webhook)
- [x] CR status handling
- [x] checked AppVersion CR status
- [x] checked ClusterDefinition CR status
- [x] CR update handling
  - [x] PVC volume expansion (spec.components[].volumeClaimTemplates only works for initial statefulset creation)
  - [x] spec.components[].serviceType
- [x] merge components from all the CRs

### ClusterDefinition CR controller
- [x] track changes and update associated CRs (Cluster, AppVersion) status
- [x] cannot delete ClusterDefinition CR if any referencing CRs (Cluster, AppVersion)

### AppVersion CR controller
- [x] immutable spec handling (via validating webhook)
- [x] CR status handling
- [x] cannot delete AppVersion CR if any referencing CRs (Cluster)

### Test
- [x] unit test