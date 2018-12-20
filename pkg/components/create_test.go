package components

import (
	"reflect"
	"testing"
)

func TestParseManifests(t *testing.T) {
	networkPolicy := map[string]string{
		"templates/test-deny-metadata.yml": `
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: test-deny-metadata
  namespace: test-namespace
  spec:
    podSelector: {}
    policyTypes:
    - Egress
    egress:
    - to:
      - ipBlock:
        cidr: 0.0.0.0/0
        except:
        - 169.254.142.0/24
`,
	}
	networkPolicyManifest := []manifest{
		{
			kind:       "NetworkPolicy",
			apiVersion: "networking.k8s.io/v1",
			namespace:  "test-namespace",
			name:       "test-deny-metadata",
		},
	}

	twoResources := map[string]string{
		"templates/test-two-resources.yml": `
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: test-ingress
  namespace: test-namespace
spec:
  rules:
  - http:
      paths:
      - path: /testpath
        backend:
          serviceName: test
          servicePort: 80
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: a-config
  namespace: default
  data:
    color: \"red\"
    multi-line: |
      hello world
      how are you?
`,
	}
	twoResourcesManifest := []manifest{
		{
			kind:       "Ingress",
			apiVersion: "extensions/v1beta1",
			namespace:  "test-namespace",
			name:       "test-ingress",
		},
		{
			kind:       "ConfigMap",
			apiVersion: "v1",
			namespace:  "default",
			name:       "a-config",
		},
	}

	tests := []struct {
		name string
		raw  map[string]string
		want []manifest
	}{
		{
			name: "ingress",
			raw:  networkPolicy,
			want: networkPolicyManifest,
		},
		{
			name: "two-resources",
			raw:  twoResources,
			want: twoResourcesManifest,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := loadManifests(test.raw)
			if err != nil {
				t.Fatalf("failed to parse manifest: %v", err)
			}
			for i := range got {
				got[i].raw = nil
			}
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("wanted %#v, got %#v", test.want, got)
			}
		})
	}

}

func TestManifestURLPath(t *testing.T) {
	tests := []struct {
		apiVersion string
		namespace  string

		plural     string
		namespaced bool

		want string
	}{
		{"v1", "my-ns", "pods", true, "/api/v1/namespaces/my-ns/pods"},
		{"apps.k8s.io/v1beta1", "my-ns", "deployments", true, "/apis/apps.k8s.io/v1beta1/namespaces/my-ns/deployments"},
		{"v1", "", "nodes", false, "/api/v1/nodes"},
		{"apiextensions.k8s.io/v1beta1", "", "customresourcedefinitions", false, "/apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions"},
		// If non-namespaced, ignore the namespace field. This is to mimic kubectl create
		// behavior, which allows this but drops the namespace.
		{"apiextensions.k8s.io/v1beta1", "my-ns", "customresourcedefinitions", false, "/apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions"},
	}

	for _, test := range tests {
		m := manifest{
			apiVersion: test.apiVersion,
			namespace:  test.namespace,
		}
		got := m.urlPath(test.plural, test.namespaced)
		if test.want != got {
			t.Errorf("{&manifest{apiVersion:%q, namespace: %q}).urlPath(%q, %t); wanted=%q, got=%q",
				test.apiVersion, test.namespace, test.plural, test.namespaced, test.want, got)
		}
	}
}
