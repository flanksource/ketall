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

package options

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	diskcached "k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

const MaxInFlightDefault = 25

type KetallOptions struct {
	UseCache        bool     `json:"useCache,omitempty"`
	AllowIncomplete bool     `json:"allowIncomplete,omitempty"`
	Scope           string   `json:"scope,omitempty"`
	Since           string   `json:"since,omitempty"`
	Selector        string   `json:"selector,omitempty"`
	FieldSelector   string   `json:"fieldSelector,omitempty"`
	MaxInflight     int64    `json:"maxInflight,omitempty"`
	Namespace       string   `json:"namespace.omitempty"`
	Exclusions      []string `json:"exclusions,omitempty"` // Exclude resources by name or kind or shortname
	Kind            string   `json:"kind,omitempty"`       // Limits results on a specific kind

	Flags *KetAllConfigFlags
}

// KetAllConfigFlags is a wrapper around genericclioptions.ConfigFlags
// to support kubeconfig directly without needing the kubeconfig path.
type KetAllConfigFlags struct {
	*genericclioptions.ConfigFlags
	KubeConfig *rest.Config `json:"kubeConfig,omitempty"`
}

func (t *KetAllConfigFlags) ToRESTConfig() (*rest.Config, error) {
	if t.KubeConfig != nil {
		return t.KubeConfig, nil
	}

	return t.ConfigFlags.ToRESTConfig()
}

func (t *KetAllConfigFlags) ToRESTMapper() (meta.RESTMapper, error) {
	if t.KubeConfig != nil {
		discoveryClient, err := t.ToDiscoveryClient()
		if err != nil {
			return nil, err
		}

		mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
		expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
		return expander, nil
	}

	return t.ConfigFlags.ToRESTMapper()
}

func (t *KetAllConfigFlags) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	if t.KubeConfig != nil {
		config := t.KubeConfig
		cacheDir := os.TempDir()
		httpCacheDir := filepath.Join(cacheDir, "http")
		discoveryCacheDir := computeDiscoverCacheDir(filepath.Join(cacheDir, "discovery"), config.Host)
		return diskcached.NewCachedDiscoveryClientForConfig(config, discoveryCacheDir, httpCacheDir, time.Duration(6*time.Hour))
	}

	return t.ConfigFlags.ToDiscoveryClient()
}

// overlyCautiousIllegalFileCharacters matches characters that *might* not be supported.  Windows is really restrictive, so this is really restrictive
var overlyCautiousIllegalFileCharacters = regexp.MustCompile(`[^(\w/.)]`)

// computeDiscoverCacheDir takes the parentDir and the host and comes up with a "usually non-colliding" name.
func computeDiscoverCacheDir(parentDir, host string) string {
	// strip the optional scheme from host if its there:
	schemelessHost := strings.Replace(strings.Replace(host, "https://", "", 1), "http://", "", 1)
	// now do a simple collapse of non-AZ09 characters.  Collisions are possible but unlikely.  Even if we do collide the problem is short lived
	safeHost := overlyCautiousIllegalFileCharacters.ReplaceAllString(schemelessHost, "_")
	return filepath.Join(parentDir, safeHost)
}

func NewDefaultCmdOptions() *KetallOptions {
	return &KetallOptions{
		MaxInflight: MaxInFlightDefault,
		Flags: &KetAllConfigFlags{
			ConfigFlags: genericclioptions.NewConfigFlags(true),
		},
	}
}

func GetGenricCliFlags() *genericclioptions.ConfigFlags {
	return genericclioptions.NewConfigFlags(true)
}
