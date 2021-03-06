/*
Copyright 2018 The Kubernetes Authors.

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

package k8sdeps

import (
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/ifc"
)

var _ ifc.Kunstructured = &UnstructAdapter{}

// UnstructAdapter wraps unstructured.Unstructured from
// https://github.com/kubernetes/apimachinery/blob/master/
//     pkg/apis/meta/v1/unstructured/unstructured.go
// to isolate dependence on apimachinery.
type UnstructAdapter struct {
	unstructured.Unstructured
}

// NewKunstructuredFromObject returns a new instance of Kunstructured.
func NewKunstructuredFromObject(obj runtime.Object) (ifc.Kunstructured, error) {
	// Convert obj to a byte stream, then convert that to JSON (Unstructured).
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return &UnstructAdapter{}, err
	}
	var u unstructured.Unstructured
	err = u.UnmarshalJSON(marshaled)
	// creationTimestamp always 'null', remove it
	u.SetCreationTimestamp(metav1.Time{})
	return &UnstructAdapter{Unstructured: u}, err
}

// GetGvk returns the Gvk name of the object.
func (fs *UnstructAdapter) GetGvk() gvk.Gvk {
	x := fs.GroupVersionKind()
	return gvk.Gvk{
		Group:   x.Group,
		Version: x.Version,
		Kind:    x.Kind,
	}
}

// Copy provides a copy behind an interface.
func (fs *UnstructAdapter) Copy() ifc.Kunstructured {
	return &UnstructAdapter{*fs.DeepCopy()}
}

// Map returns the unstructured content map.
func (fs *UnstructAdapter) Map() map[string]interface{} {
	return fs.Object
}

// SetMap overrides the unstructured content map.
func (fs *UnstructAdapter) SetMap(m map[string]interface{}) {
	fs.Object = m
}

// GetFieldValue returns value at the given fieldpath.
func (fs *UnstructAdapter) GetFieldValue(fieldPath string) (string, error) {
	return getFieldValue(fs.UnstructuredContent(), strings.Split(fieldPath, "."))
}

func getFieldValue(m map[string]interface{}, pathToField []string) (string, error) {
	if len(pathToField) == 0 {
		return "", fmt.Errorf("field not found")
	}
	if len(pathToField) == 1 {
		if v, found := m[pathToField[0]]; found {
			if s, ok := v.(string); ok {
				return s, nil
			}
			return "", fmt.Errorf("value at fieldpath is not of string type")
		}
		return "", fmt.Errorf("field at given fieldpath does not exist")
	}
	v := m[pathToField[0]]
	switch typedV := v.(type) {
	case map[string]interface{}:
		return getFieldValue(typedV, pathToField[1:])
	default:
		return "", fmt.Errorf("%#v is not expected to be a primitive type", typedV)
	}
}

func isEmptyYamlError(err error) bool {
	return strings.Contains(err.Error(), "is missing in 'null'")
}
