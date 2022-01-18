/*
Copyright 2019 Kuma authors.

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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SampleTrafficRoute is test object that wraps a sample TrafficRoute.
// +kubebuilder:object:root=true
type SampleTrafficRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Mesh string `json:"mesh"`

	Spec *apiextensionsv1.JSON `json:"spec,omitempty"`
}

// SampleTrafficRouteList contains a list of SampleTrafficRoute
// +kubebuilder:object:root=true
type SampleTrafficRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SampleTrafficRoute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SampleTrafficRoute{}, &SampleTrafficRouteList{})
}
