/*
Copyright 2019 Konvoy authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ProxyTemplateAnnotation defines an annotation that can be put on Pods
	// in order to associate them with a particular ProxyTemplate.
	// Annotation value must be a name of a ProxyTemplate resource in the same Namespace as Pod.
	ProxyTemplateAnnotation = "mesh.getkonvoy.io/proxy-template"
)

// Important: Run "make" to regenerate code after modifying this file

// ProxyTemplateSpec defines the desired state of ProxyTemplate
type ProxyTemplateSpec = map[string]interface{}

// ProxyTemplate is the Schema for the proxytemplates API
type ProxyTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProxyTemplateSpec `json:"spec,omitempty"`
}

// ProxyTemplateList contains a list of ProxyTemplate
type ProxyTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProxyTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProxyTemplate{}, &ProxyTemplateList{})
}
