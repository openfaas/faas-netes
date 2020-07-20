package k8s

import (
	"strings"
	"testing"

	corelister "k8s.io/client-go/listers/core/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type FakeLister struct {
}

func (f FakeLister) List(selector labels.Selector) (ret []*corev1.Endpoints, err error) {
	return nil, nil
}

func (f FakeLister) Endpoints(namespace string) corelister.EndpointsNamespaceLister {

	return FakeNSLister{}
}

type FakeNSLister struct {
}

func (f FakeNSLister) List(selector labels.Selector) (ret []*corev1.Endpoints, err error) {
	return nil, nil
}

func (f FakeNSLister) Get(name string) (*corev1.Endpoints, error) {
	ep := corev1.Endpoints{
		Subsets: []corev1.EndpointSubset{
			corev1.EndpointSubset{
				Addresses: []corev1.EndpointAddress{
					corev1.EndpointAddress{IP: "127.0.0.1"},
				},
			},
		},
	}

	return &ep, nil
}

func Test_FunctionLookup(t *testing.T) {

	lister := FakeLister{}

	resolver := NewFunctionLookup("testDefault", lister)

	cases := []struct {
		name     string
		funcName string
		expError string
		expUrl   string
	}{
		{
			name:     "function without namespace uses default namespace",
			funcName: "testfunc",
			expUrl:   "http://127.0.0.1:8080",
		},
		{
			name:     "function with namespace uses the given namespace",
			funcName: "testfunc.othernamespace",
			expUrl:   "http://127.0.0.1:8080",
		},
		{
			name:     "url parse errors are returned",
			funcName: "testfunc.kube-system",
			expError: "namespace not allowed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			url, err := resolver.Resolve(tc.funcName)
			if tc.expError == "" && err != nil {
				t.Fatalf("expected no error, got %s", err)
			}

			if tc.expError != "" && (err == nil || !strings.Contains(err.Error(), tc.expError)) {
				t.Fatalf("expected %s, got %s", tc.expError, err)
			}

			if url.String() != tc.expUrl {
				t.Fatalf("expected url %s, got %s", tc.expUrl, url.String())
			}
		})
	}
}
