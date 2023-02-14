/*
Copyright ApeCloud, Inc.

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

package component

import (
	"embed"
	"encoding/json"

	"github.com/leaanthony/debme"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"

	dbaasv1alpha1 "github.com/apecloud/kubeblocks/apis/dbaas/v1alpha1"
	intctrlutil "github.com/apecloud/kubeblocks/internal/controllerutil"
)

// ClusterDefinitionComponent CharacterType Const Define
const (
	kMysql             = "mysql"
	defaultMonitorPort = 9104
)

var (
	supportedCharacterTypeFunc = map[string]func(cluster *dbaasv1alpha1.Cluster, component *Component) error{
		kMysql: setMysqlComponent,
	}
	//go:embed cue/*
	cueTemplates embed.FS
)

type mysqlMonitorConfig struct {
	SecretName      string `json:"secretName"`
	InternalPort    int32  `json:"internalPort"`
	Image           string `json:"image"`
	ImagePullPolicy string `json:"imagePullPolicy"`
}

func buildMysqlMonitorContainer(key string, monitor *mysqlMonitorConfig) (*corev1.Container, error) {
	cueFS, _ := debme.FS(cueTemplates, "cue/monitor")

	cueTpl, err := intctrlutil.NewCUETplFromBytes(cueFS.ReadFile("mysql_template.cue"))
	if err != nil {
		return nil, err
	}
	cueValue := intctrlutil.NewCUEBuilder(*cueTpl)

	mysqlMonitorStrByte, err := json.Marshal(monitor)
	if err != nil {
		return nil, err
	}

	if err = cueValue.Fill("monitor", mysqlMonitorStrByte); err != nil {
		return nil, err
	}

	containerStrByte, err := cueValue.Lookup("container")
	if err != nil {
		return nil, err
	}
	container := corev1.Container{}
	if err = json.Unmarshal(containerStrByte, &container); err != nil {
		return nil, err
	}
	return &container, nil
}

func setMysqlComponent(cluster *dbaasv1alpha1.Cluster, component *Component) error {
	image := viper.GetString(intctrlutil.KBImage)
	imagePullPolicy := viper.GetString(intctrlutil.KBImagePullPolicy)

	mysqlMonitorConfig := &mysqlMonitorConfig{
		SecretName: cluster.Name,
		// TODO: port value will be checked against other containers.
		InternalPort:    defaultMonitorPort,
		Image:           image,
		ImagePullPolicy: imagePullPolicy,
	}

	container, err := buildMysqlMonitorContainer(cluster.Name, mysqlMonitorConfig)
	if err != nil {
		return err
	}

	component.PodSpec.Containers = append(component.PodSpec.Containers, *container)
	component.Monitor = &MonitorConfig{
		Enable:     true,
		ScrapePath: "/metrics",
		ScrapePort: mysqlMonitorConfig.InternalPort,
	}
	return nil
}

// isSupportedCharacterType check CharacterType is wellknown
func isSupportedCharacterType(characterType string) bool {
	return isMappedCharacterType(characterType, supportedCharacterTypeFunc)
}

func isMappedCharacterType(characterType string,
	processors map[string]func(*dbaasv1alpha1.Cluster, *Component) error) bool {
	if val, ok := processors[characterType]; ok && val != nil {
		return true
	}
	return false
}