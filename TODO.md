# TODO items

## KubeBlocks controllers

### Cluster CR controller
- [x] secondary resources finalizer
- [x] CR delete handling
  - [x] delete secondary resources
  - [x] CR spec.terminationPolicy handling
- [x] managed resources handling
  - [x] deployment workloads
  - [x] PDB
  - [x] label handling:
    - [x] deploy & sts workloads' labels and spec.template.metadata.labels (check https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/)
- [x] immutable spec properties handling (via validating webhook)
- [x] CR status handling
- [x] checked ClusterVersion CR status
- [x] checked ClusterDefinition CR status
- [x] CR update handling
  - [x] PVC volume expansion (spec.components[].volumeClaimTemplates only works for initial statefulset creation)
  - [x] spec.components[].serviceType
- [x] merge components from all the CRs

### ClusterDefinition CR controller
- [x] track changes and update associated CRs (Cluster, ClusterVersion) status
- [x] cannot delete ClusterDefinition CR if any referencing CRs (Cluster, ClusterVersion)

### ClusterVersion CR controller
- [x] immutable spec handling (via validating webhook)
- [x] CR status handling
- [x] cannot delete ClusterVersion CR if any referencing CRs (Cluster)

### Test
- [x] unit test