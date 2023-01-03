/*
Copyright ApeCloud Inc.

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

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"

	dbaasv1alpha1 "github.com/apecloud/kubeblocks/apis/dbaas/v1alpha1"
	"github.com/apecloud/kubeblocks/internal/cli/builder"
	"github.com/apecloud/kubeblocks/internal/cli/cluster"
	"github.com/apecloud/kubeblocks/internal/cli/patch"
	"github.com/apecloud/kubeblocks/internal/cli/types"
)

var clusterUpdateExample = templates.Examples(`
	# update cluster mycluster termination policy to Delete
	kbcli cluster update mycluster --termination-policy=Delete

	# enable cluster monitor
	kbcli cluster update mycluster --monitor=true

    # enable all logs
	kbcli cluster update mycluster --enable-all-logs=true

    # update cluster topology keys and affinity
	kbcli cluster update mycluster --topology-keys=kubernetes.io/hostname --pod-anti-affinity=Required

	# update cluster tolerations
	kbcli cluster update mycluster --tolerations='"key=engineType,value=mongo,operator=Equal,effect=NoSchedule","key=diskType,value=ssd,operator=Equal,effect=NoSchedule"'
`)

type updateOptions struct {
	name      string
	namespace string
	dynamic   dynamic.Interface
	cluster   *dbaasv1alpha1.Cluster

	UpdatableFlags
	*patch.Options
}

func newUpdateOptions(streams genericclioptions.IOStreams) *updateOptions {
	return &updateOptions{Options: patch.NewOptions(streams)}
}

func NewUpdateCmd(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := newUpdateOptions(streams)
	return builder.NewCmdBuilder().
		Use("update").
		Short("Update the cluster").
		Example(clusterUpdateExample).
		Options(o).
		IOStreams(streams).
		Factory(f).
		GVR(types.ClusterGVR()).
		CustomFlags(o.addFlags).
		CustomComplete(o.complete).
		Build(o.Build)
}

func (o *updateOptions) addFlags(c *builder.Command) {
	o.UpdatableFlags.addFlags(c.Cmd)
}

func (o *updateOptions) complete(c *builder.Command) error {
	var err error
	if len(c.Args) == 0 {
		return fmt.Errorf("missing updated cluster name")
	}
	if len(c.Args) > 1 {
		return fmt.Errorf("only support to update one cluster")
	}
	o.name = c.Args[0]

	// record the flags that been set by user
	var flags []*pflag.Flag
	c.Cmd.Flags().Visit(func(flag *pflag.Flag) {
		flags = append(flags, flag)
	})

	// nothing to do
	if len(flags) == 0 {
		return nil
	}

	if o.namespace, _, err = c.Factory.ToRawKubeConfigLoader().Namespace(); err != nil {
		return err
	}
	if o.dynamic, err = c.Factory.DynamicClient(); err != nil {
		return err
	}
	return o.buildPatch(flags)
}

func (o *updateOptions) buildPatch(flags []*pflag.Flag) error {
	var err error
	type buildFn func(obj map[string]interface{}, v string, field string) error

	buildFlagObj := func(obj map[string]interface{}, v string, field string) error {
		return unstructured.SetNestedField(obj, v, field)
	}
	buildTolObj := func(obj map[string]interface{}, v string, field string) error {
		tolerations := buildTolerations(o.TolerationsRaw)
		return unstructured.SetNestedField(obj, tolerations, field)
	}
	buildComps := func(obj map[string]interface{}, v string, field string) error {
		return o.buildComponents(field, v)
	}

	spec := map[string]interface{}{}
	affinity := map[string]interface{}{}
	type filedObj struct {
		field string
		obj   map[string]interface{}
		fn    buildFn
	}

	flagFieldMapping := map[string]*filedObj{
		"termination-policy": {field: "terminationPolicy", obj: spec, fn: buildFlagObj},
		"pod-anti-affinity":  {field: "podAntiAffinity", obj: affinity, fn: buildFlagObj},
		"topology-keys":      {field: "topologyKeys", obj: affinity, fn: buildFlagObj},
		"node-labels":        {field: "nodeLabels", obj: affinity, fn: buildFlagObj},
		"tolerations":        {field: "tolerations", obj: spec, fn: buildTolObj},
		"monitor":            {field: "monitor", obj: nil, fn: buildComps},
		"enable-all-logs":    {field: "enable-all-logs", obj: nil, fn: buildComps},
	}

	for _, flag := range flags {
		if f, ok := flagFieldMapping[flag.Name]; ok {
			if err = f.fn(f.obj, flag.Value.String(), f.field); err != nil {
				return err
			}
		}
	}

	if len(affinity) > 0 {
		if err = unstructured.SetNestedField(spec, affinity, "affinity"); err != nil {
			return err
		}
	}

	if o.cluster != nil {
		data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&o.cluster.Spec)
		if err != nil {
			return err
		}

		if err = unstructured.SetNestedField(spec, data["components"], "components"); err != nil {
			return err
		}
	}

	obj := unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": spec,
		},
	}
	bytes, err := obj.MarshalJSON()
	if err != nil {
		return err
	}
	o.Patch = string(bytes)
	return nil
}

func (o *updateOptions) buildComponents(field string, val string) error {
	if o.cluster == nil {
		cluster, err := cluster.GetClusterByName(o.dynamic, o.name, o.namespace)
		if err != nil {
			return err
		}
		o.cluster = cluster
	}

	switch field {
	case "monitor":
		return o.setMonitor(val)
	case "enable-all-logs":
		return o.setEnabledLog(val)
	default:
		return nil
	}
}

func (o *updateOptions) setEnabledLog(val string) error {
	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}

	// disable all monitor
	if !boolVal {
		for _, c := range o.cluster.Spec.Components {
			c.EnabledLogs = nil
		}
		return nil
	}

	// enable all monitor
	cd, err := cluster.GetClusterDefByName(o.dynamic, o.cluster.Spec.ClusterDefRef)
	if err != nil {
		return err
	}
	setEnableAllLogs(o.cluster, cd)
	return nil
}

func (o *updateOptions) setMonitor(val string) error {
	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}

	for i := range o.cluster.Spec.Components {
		o.cluster.Spec.Components[i].Monitor = boolVal
	}
	return nil
}