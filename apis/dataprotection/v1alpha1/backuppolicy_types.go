/*
Copyright 2022 The KubeBlocks Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupPolicySpec defines the desired state of BackupPolicy
type BackupPolicySpec struct {
	// policy can inherit from backup config and override some fields.
	// +optional
	BackupPolicyTemplateName string `json:"backupPolicyTemplateName,omitempty"`

	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	// +kubebuilder:default="0 7 * * *"
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// which backup tool to perform database backup, only support one tool.
	// +optional
	BackupToolName string `json:"backupToolName,omitempty"`

	// TTL is a time.Duration-parseable string describing how long
	// the Backup should be retained for.
	// +optional
	TTL metav1.Duration `json:"ttl,omitempty"`

	// database cluster service
	// +kubebuilder:validation:Required
	Target TargetCluster `json:"target"`

	// execute hook commands for backup.
	Hooks BackupPolicyHook `json:"hooks"`

	// array of remote volumes from CSI driver definition.
	// +kubebuilder:validation:Required
	RemoteVolume corev1.Volume `json:"remoteVolume"`

	// count of backup stop retries on fail.
	// +optional
	OnFailAttempted int32 `json:"onFailAttempted,omitempty"`
}

// TargetCluster TODO (dsj): target cluster need redefined from Cluster API
type TargetCluster struct {
	// database engine to support in the backup.
	// +kubebuilder:validation:Enum={mysql}
	// +kubebuilder:validation:Required
	DatabaseEngine string `json:"databaseEngine"`

	// database engine to support in the backup.
	// +kubebuilder:validation:Enum={5.6,5.7,8.0}
	// +optional
	DatabaseEngineVersions []string `json:"databaseEngineVersion,omitempty"`

	// LabelSelector is used to find matching pods.
	// Pods that match this label selector are counted to determine the number of pods
	// in their corresponding topology domain.
	// +kubebuilder:validation:Required
	LabelsSelector *metav1.LabelSelector `json:"labelsSelector"`

	// target db cluster access secret
	// +kubebuilder:validation:Required
	Secret BackupPolicySecret `json:"secret"`
}

// BackupPolicySecret defined for the target database secret that backup tool can connect.
type BackupPolicySecret struct {

	// the secret name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// which key name for user
	// +kubebuilder:default=username
	// +optional
	KeyUser string `json:"keyUser,omitempty"`

	// which key name for password
	// +kubebuilder:default=password
	// +optional
	KeyPassword string `json:"keyPassword,omitempty"`
}

// BackupPolicyHook defined for the database execute commands before and after backup.
type BackupPolicyHook struct {

	// pre backup to perform commands
	// +optional
	PreCommands []string `json:"preCommands,omitempty"`

	// post backup to perform commands
	// +optional
	PostCommands []string `json:"postCommands,omitempty"`

	// exec command with image
	// TODO(dsj): use opendbaas-core
	// +kubebuilder:default="rancher/kubectl:v1.23.7"
	// +optional
	Image string `json:"image,omitempty"`

	// which container can exec command
	// +kubebuilder:default=mysql
	// +optional
	ContainerName string `json:"ContainerName,omitempty"`
}

// BackupPolicyStatus defines the observed state of BackupPolicy
type BackupPolicyStatus struct {
	// backup policy phase valid value: available, failed, new.
	// +optional
	Phase BackupPolicyTemplatePhase `json:"phase,omitempty"`

	// the reason if backup policy check failed.
	// +optional
	FailureReason string `json:"failureReason,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={dbaas},scope=Namespaced

// BackupPolicy is the Schema for the backuppolicies API  (defined by User)
type BackupPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupPolicySpec   `json:"spec,omitempty"`
	Status BackupPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BackupPolicyList contains a list of BackupPolicy
type BackupPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupPolicy{}, &BackupPolicyList{})
}