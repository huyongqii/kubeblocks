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

package v1alpha1

import (
	"context"
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var clusterversionlog = logf.Log.WithName("clusterversion-resource")

func (r *ClusterVersion) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-dbaas-kubeblocks-io-v1alpha1-clusterversion,mutating=true,failurePolicy=fail,sideEffects=None,groups=dbaas.kubeblocks.io,resources=clusterversions,verbs=create;update,versions=v1alpha1,name=mclusterversion.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ClusterVersion{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ClusterVersion) Default() {
	clusterversionlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-dbaas-kubeblocks-io-v1alpha1-clusterversion,mutating=false,failurePolicy=fail,sideEffects=None,groups=dbaas.kubeblocks.io,resources=clusterversions,verbs=create;update,versions=v1alpha1,name=vclusterversion.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ClusterVersion{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterVersion) ValidateCreate() error {
	clusterversionlog.Info("validate create", "name", r.Name)
	return r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterVersion) ValidateUpdate(old runtime.Object) error {
	clusterversionlog.Info("validate update", "name", r.Name)
	// determine whether r.spec content is modified
	lastClusterVersion := old.(*ClusterVersion)
	if !reflect.DeepEqual(lastClusterVersion.Spec, r.Spec) {
		return newInvalidError(ClusterVersionKind, r.Name, "", "ClusterVersion.spec is immutable, you can not update it.")
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterVersion) ValidateDelete() error {
	clusterversionlog.Info("validate delete", "name", r.Name)
	return nil
}

// Validate ClusterVersion.spec is legal
func (r *ClusterVersion) validate() error {
	var (
		allErrs    field.ErrorList
		ctx        = context.Background()
		clusterDef = &ClusterDefinition{}
	)
	if webhookMgr == nil {
		return nil
	}
	err := webhookMgr.client.Get(ctx, types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Spec.ClusterDefinitionRef,
	}, clusterDef)

	if err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clusterDefinitionRef"),
			r.Spec.ClusterDefinitionRef, err.Error()))
	} else {
		notFoundComponentTypes, noContainersComponents := r.GetInconsistentComponentsInfo(clusterDef)

		if len(notFoundComponentTypes) > 0 {
			allErrs = append(allErrs, field.NotFound(field.NewPath("spec.components[*].type"),
				getComponentTypeNotFoundMsg(notFoundComponentTypes, r.Spec.ClusterDefinitionRef)))
		}

		if len(noContainersComponents) > 0 {
			allErrs = append(allErrs, field.NotFound(field.NewPath("spec.components[*].type"),
				fmt.Sprintf("containers are not defined in ClusterDefinition.spec.components[*]: %v", noContainersComponents)))
		}
	}

	if len(allErrs) > 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{Group: APIVersion, Kind: ClusterVersionKind},
			r.Name, allErrs)
	}
	return nil
}

// GetInconsistentComponentsInfo get clusterVersion invalid component type and no containers component compared with clusterDefinitionDef
func (r *ClusterVersion) GetInconsistentComponentsInfo(clusterDef *ClusterDefinition) ([]string, []string) {

	var (
		// clusterDefinition components to map. the value of map represents whether there is a default containers
		componentMap          = map[string]bool{}
		notFoundComponentType = make([]string, 0)
		noContainersComponent = make([]string, 0)
	)

	for _, v := range clusterDef.Spec.Components {
		if v.PodSpec == nil || v.PodSpec.Containers == nil || len(v.PodSpec.Containers) == 0 {
			componentMap[v.TypeName] = false
		} else {
			componentMap[v.TypeName] = true
		}
	}
	// get not found component type in clusterDefinition
	for _, v := range r.Spec.Components {
		if _, ok := componentMap[v.Type]; !ok {
			notFoundComponentType = append(notFoundComponentType, v.Type)
		} else if v.PodSpec.Containers != nil && len(v.PodSpec.Containers) > 0 {
			componentMap[v.Type] = true
		}
	}
	// get no containers components in clusterDefinition and clusterVersion
	for k, v := range componentMap {
		if !v {
			noContainersComponent = append(noContainersComponent, k)
		}
	}
	return notFoundComponentType, noContainersComponent
}

func getComponentTypeNotFoundMsg(invalidComponentTypes []string, clusterDefName string) string {
	return fmt.Sprintf(" %v is not found in ClusterDefinition.spec.components[*].typeName %s",
		invalidComponentTypes, clusterDefName)
}

// NewInvalidError create an invalid api error
func newInvalidError(kind, resourceName, path, reason string) error {
	return apierrors.NewInvalid(schema.GroupKind{Group: APIVersion, Kind: kind},
		resourceName, field.ErrorList{field.InternalError(field.NewPath(path),
			fmt.Errorf(reason))})
}