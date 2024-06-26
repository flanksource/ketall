/*
Copyright 2019 Cornelius Weig

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

package ketall

import (
	"context"
	"fmt"

	"github.com/flanksource/ketall/client"
	"github.com/flanksource/ketall/filter"
	"github.com/flanksource/ketall/options"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

func KetAll(ctx context.Context, ketallOptions *options.KetallOptions) ([]*unstructured.Unstructured, error) {
	all, err := client.GetAllServerResources(ctx, ketallOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get server resources: %w", err)
	}

	filtered := filter.ApplyFilter(all, ketallOptions.Since)
	if filtered == nil {
		return nil, nil
	}

	items := filtered.(*v1.List).Items
	var unstructuredItems []*unstructured.Unstructured
	for _, item := range items {
		if unstructuredItem, ok := item.Object.(*unstructured.Unstructured); ok {
			unstructuredItems = append(unstructuredItems, unstructuredItem)
		} else {
			klog.V(1).Infof("item is not type *unstructured.Unstructured. It's of type %T\n", item)
		}
	}

	return unstructuredItems, nil
}

func KetOne(ctx context.Context, name, namespace, kind string, ketallOptions *options.KetallOptions) (*unstructured.Unstructured, error) {
	ketallOptions.Kind = kind

	// Override namespace and fieldselector
	ketallOptions.Namespace = namespace
	ketallOptions.FieldSelector = "metadata.name=" + name

	all, err := client.GetAllServerResources(ctx, ketallOptions)
	if err != nil {
		return nil, err
	}

	items := all.(*v1.List).Items
	if len(items) == 0 {
		return nil, nil
	}

	return items[0].Object.(*unstructured.Unstructured), nil
}
