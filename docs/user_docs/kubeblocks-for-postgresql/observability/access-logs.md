---
title: Access logs
description: How to access cluster log files
sidebar_position: 3
---

# Access logs
The KubeBlocks log enhancement function aims to simplify the complexities of troubleshooting. With kbcli, the command line tool of KubeBlocks, you can view all kinds of logs generated by the database clusters running on KubeBlocks, such as slow logs, error logs, audit logs, and the container running logs (Stdout and Stderr).
The KubeBlocks log enhancement function uses methods similar to kubectl exec and kubectl logs to ensure a self-closed loop and lightweight.

## Before you start

- The container image supports `tail` and `xargs` commands.
- KubeBlocks Operator is installed on the target Kubernetes cluster.

## Steps

1. Enable the log enhancement function.
   - If you create a cluster by running the `kbcli cluster create` command, add the `--enable-all-logs=true` option to enable the log enhancement function. When this option is `true`, all the log types defined by `spec.components.logConfigs` in `ClusterDefinition` are enabled automatically.

     ```bash
     kbcli cluster create pg-cluster --cluster-definition='postgresql' --enable-all-logs=true 
     ```
   - If you create a cluster by applying a YAML file, add the log type you need in `spec.components.enabledLogs`. As for PostgreSQL, running log is supported.

     ```YAML
     apiVersion: apps.kubeblocks.io/v1alpha1
     kind: Cluster
     metadata:
       name: pg-cluster
       namespace: default
     spec:
       clusterDefinitionRef: postgresql-cluster-definition
       appVersionRef: appversion-postgresql-latest
       components:
       - name: replicasets
         type: replicasets
         enabledLogs:
           - running
     ```
   
2. View the supported logs.
   
   Run the `kbcli cluster list-logs` command to view the enabled log types of the target cluster and the log file details. INSTANCE of each node is displayed.
   
   ***Example***

   ```bash
   kbcli cluster list-logs pg-cluster
   >
   INSTANCE                         LOG-TYPE        FILE-PATH                                          SIZE    LAST-WRITTEN                      COMPONENT
   pg-cluster-postgresql-0-0        running         /postgresql/data/log/postgresql-2023-03-03.csv     6.9K    Mar 03, 2023 07:17 (UTC+00:00)    postgresql
   pg-cluster-postgresql-0-0        running         /postgresql/data/log/postgresql-2023-03-03.log     326     Mar 03, 2023 07:17 (UTC+00:00)    postgresql         
   ```

3. Access the cluster log file.
   
   Run the `kbcli cluster logs <name>` command to view the details of the target log file generated by the target instance on the target cluster. You can use different options to view the log file details you need. 
   You can also run `kbcli cluster logs -h` to see the examples and option descriptions.   
   ```bash
   kbcli cluster logs -h
   >
   Access cluster log file

   Examples:
     # Return snapshot logs from cluster mycluster with default primary instance (stdout)
     kbcli cluster logs mycluster

     # Display only the most recent 20 lines from cluster mycluster with default primary instance (stdout)
     kbcli cluster logs --tail=20 mycluster

     # Return snapshot logs from cluster mycluster with specify instance my-instance-0 (stdout)
     kbcli cluster logs mycluster --instance my-instance-0

     # Return snapshot logs from cluster mycluster with specify instance my-instance-0 and specify container
     # my-container (stdout)
     kbcli cluster logs mycluster --instance my-instance-0 -c my-container

     # Return slow logs from cluster mycluster with default primary instance
     kbcli cluster logs mycluster --file-type=slow

     # Begin streaming the slow logs from cluster mycluster with default primary instance
     kbcli cluster logs -f mycluster --file-type=slow

     # Return the specify file logs from cluster mycluster with specify instance my-instance-0
     kbcli cluster logs mycluster --instance my-instance-0 --file-path=/var/log/yum.log

     # Return the specify file logs from cluster mycluster with specify instance my-instance-0 and specify
     # container my-container
     kbcli cluster logs mycluster --instance my-instance-0 -c my-container --file-path=/var/log/yum.log
   ```

4. (Optional) Troubleshooting.
     
     The log enhancement function does not affect the core process of KubeBlocks. If a configuration exception occurs, a warning shows to help troubleshoot.
     `warning` is recorded in the `event` and `status.Conditions` of the target database cluster. 

     View `warning` information.
     - Run `kbcli cluster describe <cluster-name>` to view the status of the target cluster. You can also run `kbcli cluster list events <cluster-name>` to view the event information of the target cluster directly.
 
     - Run `kubectl describe cluster <cluster-name>` to view the warning.
  
     ***Example***
     
     ```
     Status:
       Cluster Def Generation:  3
       Components:
          Replicasets:
            Phase:  Running
       Conditions:
         Last Transition Time:  2022-11-11T03:57:42Z
         Message:               EnableLogs of cluster component replicasets has invalid value [errora slowa] which isn't defined in cluster definition component replicasets
         Reason:                EnableLogsListValidateFail
         Status:                False
         Type:                  ValidateEnabledLogs
       Observed Generation:     2
       Operations:
         Horizontal Scalable:
            Name:  replicasets
         Restartable:
           replicasets
         Vertical Scalable:
           replicasets
       Phase:  Running
     Events:
       Type     Reason                      Age   From                Message
       ----     ------                      ----  ----                -------
       Normal   Creating                    49s   cluster-controller  Start Creating in Cluster: release-name-error
       Warning  EnableLogsListValidateFail  49s   cluster-controller  EnableLogs of cluster component replicasets has invalid value [errora slowa] which isn't defined in cluster definition component replicasets
       Normal   Running                     36s   cluster-controller  Cluster: release-name-error is ready, current phase is Running
     ```