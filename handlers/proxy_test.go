package handlers

import (
	"strings"
	"testing"
)

func Test_FunctionLookup(t *testing.T) {
	resolver := FunctionLookup{DefaultNamespace: "testDefault"}

	cases := []struct {
		name     string
		funcName string
		expError string
		expUrl   string
	}{
		{
			name:     "function without namespace uses default namespace",
			funcName: "testfunc",
			expUrl:   "http://testfunc.testDefault:8080",
		},
		{
			name:     "function with namespace uses the given namespace",
			funcName: "testfunc.othernamespace",
			expUrl:   "http://testfunc.othernamespace:8080",
		},
		{
			name:     "url parse errors are returned",
			funcName: "\ttestfunc.othernamespace",
			expError: "net/url: invalid control character in URL",
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
