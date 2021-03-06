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

package transformers

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/internal/k8sdeps"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/types"
)

func TestImageTagTransformer(t *testing.T) {
	var rf = resource.NewFactory(
		k8sdeps.NewKunstructuredFactoryImpl())

	m := resmap.ResMap{
		resid.NewResId(deploy, "deploy1"): rf.FromMap(
			map[string]interface{}{
				"group":      "apps",
				"apiVersion": "v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "deploy1",
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"initContainers": []interface{}{
								map[string]interface{}{
									"name":  "nginx2",
									"image": "my-nginx:1.8.0",
								},
							},
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:1.7.9",
								},
								map[string]interface{}{
									"name":  "replaced-with-digest",
									"image": "foobar:1",
								},
							},
						},
					},
				},
			}),
		resid.NewResId(gvk.Gvk{Kind: "randomKind"}, "random"): rf.FromMap(
			map[string]interface{}{
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx1",
									"image": "nginx",
								},
								map[string]interface{}{
									"name":  "myimage",
									"image": "myprivaterepohostname:1234/my/image:latest",
								},
							},
						},
					},
				},
				"spec2": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx3",
									"image": "nginx:v1",
								},
								map[string]interface{}{
									"name":  "nginx4",
									"image": "my-nginx:latest",
								},
							},
						},
					},
				},
			}),
	}
	expected := resmap.ResMap{
		resid.NewResId(deploy, "deploy1"): rf.FromMap(
			map[string]interface{}{
				"group":      "apps",
				"apiVersion": "v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "deploy1",
				},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"initContainers": []interface{}{
								map[string]interface{}{
									"name":  "nginx2",
									"image": "my-nginx:previous",
								},
							},
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx",
									"image": "nginx:v2",
								},
								map[string]interface{}{
									"name":  "replaced-with-digest",
									"image": "foobar@sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3",
								},
							},
						},
					},
				},
			}),
		resid.NewResId(gvk.Gvk{Kind: "randomKind"}, "random"): rf.FromMap(
			map[string]interface{}{
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx1",
									"image": "nginx:v2",
								},
								map[string]interface{}{
									"name":  "myimage",
									"image": "myprivaterepohostname:1234/my/image:v1.0.1",
								},
							},
						},
					},
				},
				"spec2": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "nginx3",
									"image": "nginx:v2",
								},
								map[string]interface{}{
									"name":  "nginx4",
									"image": "my-nginx:previous",
								},
							},
						},
					},
				},
			}),
	}

	it, err := NewImageTagTransformer([]types.ImageTag{
		{Name: "nginx", NewTag: "v2"},
		{Name: "my-nginx", NewTag: "previous"},
		{Name: "myprivaterepohostname:1234/my/image", NewTag: "v1.0.1"},
		{Name: "foobar", Digest: "sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = it.Transform(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(m, expected) {
		err = expected.ErrorIfNotEqual(m)
		t.Fatalf("actual doesn't match expected: %v", err)
	}
}
