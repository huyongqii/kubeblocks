/*
Copyright © 2022 The OpenCli Authors

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

package cluster

import "jihulab.com/infracreate/dbaas-system/opencli/pkg/types"

// Installer defines the interface for handling the cluster(k3d/k3s) management
type Installer interface {
	Install() error
	Uninstall() error
	GenKubeconfig() error
	SetKubeconfig() error
	InstallDeps() error
	GetStatus() types.ClusterStatus
	PrintGuide() error
}