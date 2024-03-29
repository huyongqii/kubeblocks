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

package configuration

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/StudioSol/set"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/apecloud/kubeblocks/apis/apps/v1alpha1"
	testapps "github.com/apecloud/kubeblocks/internal/testutil/apps"
	"github.com/apecloud/kubeblocks/test/testdata"
)

var _ = Describe("config_util", func() {

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("MergeAndValidateConfigs", func() {
		It("Should success with no error", func() {
			type args struct {
				configConstraint v1alpha1.ConfigConstraintSpec
				baseCfg          map[string]string
				updatedParams    []ParamPairs
				cmKeys           []string
			}

			configConstraintObj := testapps.NewCustomizedObj("resources/mysql-config-constraint.yaml",
				&v1alpha1.ConfigConstraint{}, func(cc *v1alpha1.ConfigConstraint) {
					if ccContext, err := testdata.GetTestDataFileContent("cue_testdata/pg14.cue"); err == nil {
						cc.Spec.ConfigurationSchema = &v1alpha1.CustomParametersValidation{
							CUE: string(ccContext),
						}
					}
					cc.Spec.FormatterConfig = &v1alpha1.FormatterConfig{
						Format: v1alpha1.Properties,
					}
				})

			cfgContext, err := testdata.GetTestDataFileContent("cue_testdata/pg14.conf")
			Expect(err).Should(Succeed())

			tests := []struct {
				name    string
				args    args
				want    map[string]string
				wantErr bool
			}{{
				name: "pg1_merge",
				args: args{
					configConstraint: configConstraintObj.Spec,
					baseCfg: map[string]string{
						"key":  string(cfgContext),
						"key2": "not support context",
					},
					updatedParams: []ParamPairs{
						{
							Key: "key",
							UpdatedParams: map[string]interface{}{
								"max_connections": "200",
								"shared_buffers":  "512M",
							},
						},
					},
					cmKeys: []string{"key", "key3"},
				},
				want: map[string]string{
					"max_connections": "200",
					"shared_buffers":  "512M",
				},
			}, {
				name: "not_support_key_updated",
				args: args{
					configConstraint: configConstraintObj.Spec,
					baseCfg: map[string]string{
						"key":  string(cfgContext),
						"key2": "not_support_context",
					},
					updatedParams: []ParamPairs{
						{
							Key: "key",
							UpdatedParams: map[string]interface{}{
								"max_connections": "200",
								"shared_buffers":  "512M",
							},
						},
					},
					cmKeys: []string{"key1", "key2"},
				},
				wantErr: true,
			}}
			for _, tt := range tests {
				got, err := MergeAndValidateConfigs(tt.args.configConstraint, tt.args.baseCfg, tt.args.cmKeys, tt.args.updatedParams)
				Expect(err != nil).Should(BeEquivalentTo(tt.wantErr))
				if tt.wantErr {
					continue
				}

				option := CfgOption{
					Type:    CfgTplType,
					CfgType: tt.args.configConstraint.FormatterConfig.Format,
				}

				patch, err := CreateMergePatch(&ConfigResource{
					ConfigData: tt.args.baseCfg,
				}, &ConfigResource{
					ConfigData: got,
				}, option)
				Expect(err).Should(Succeed())

				var patchJSON map[string]string
				Expect(json.Unmarshal(patch.UpdateConfig["key"], &patchJSON)).Should(Succeed())
				Expect(patchJSON).Should(BeEquivalentTo(tt.want))
			}
		})
	})
})

func TestFromUpdatedConfig(t *testing.T) {
	type args struct {
		base map[string]string
		sets *set.LinkedHashSetString
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{{
		name: "normal_test",
		args: args{
			base: map[string]string{
				"key1": "config context1",
				"key2": "config context2",
				"key3": "config context2",
			},
			sets: set.NewLinkedHashSetString("key1", "key3"),
		},
		want: map[string]string{
			"key1": "config context1",
			"key3": "config context2",
		},
	}, {
		name: "none_updated_test",
		args: args{
			base: map[string]string{
				"key1": "config context1",
				"key2": "config context2",
				"key3": "config context2",
			},
			sets: set.NewLinkedHashSetString(),
		},
		want: map[string]string{},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fromUpdatedConfig(tt.args.base, tt.args.sets); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fromUpdatedConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeUpdatedConfig(t *testing.T) {
	type args struct {
		baseMap    map[string]string
		updatedMap map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{{
		name: "normal_test",
		args: args{
			baseMap: map[string]string{
				"key1": "context1",
				"key2": "context2",
				"key3": "context3",
			},
			updatedMap: map[string]string{
				"key2": "new context",
			},
		},
		want: map[string]string{
			"key1": "context1",
			"key2": "new context",
			"key3": "context3",
		},
	}, {
		name: "not_expected_update_test",
		args: args{
			baseMap: map[string]string{
				"key1": "context1",
				"key2": "context2",
				"key3": "context3",
			},
			updatedMap: map[string]string{
				"key6": "context6",
			},
		},
		want: map[string]string{
			"key1": "context1",
			"key2": "context2",
			"key3": "context3",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeUpdatedConfig(tt.args.baseMap, tt.args.updatedMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeUpdatedConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
