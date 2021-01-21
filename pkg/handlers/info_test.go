// Copyright 2020 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/openfaas/faas-provider/types"
)

func Test_InfoHandler(t *testing.T) {
	sha := "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
	version := "0.0.1"
	handler := MakeInfoHandler(version, sha)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler(w, r)

	resp := types.ProviderInfo{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("unexpected error unmarshalling the response")
	}

	if resp.Name != ProviderName {
		t.Fatalf("expected provider %q, got %q", ProviderName, resp.Name)
	}

	if resp.Orchestration != OrchestrationIdentifier {
		t.Fatalf("expected orchestration %q, got %q", OrchestrationIdentifier, resp.Orchestration)
	}

	if resp.Version.SHA != sha {
		t.Fatalf("expected orchestration %q, got %q", sha, resp.Version.SHA)
	}

	if resp.Version.Release != version {
		t.Fatalf("expected release %q, got %q", version, resp.Version.Release)
	}
}
