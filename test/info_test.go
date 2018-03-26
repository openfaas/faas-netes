package test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openfaas/faas-netes/handlers"
	"github.com/openfaas/faas-provider/types"
)

const (
	infoTestVersion = "netestest"
	infoTestSHA     = "test"
)

func TestMakeInfoHandler(t *testing.T) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/system/info", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := handlers.MakeInfoHandler(infoTestVersion, infoTestSHA)
	infoRequest := types.InfoRequest{}

	handler(rr, req)
	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(body, &infoRequest)
	if err != nil {
		t.Fatal(err)
	}

	if required := http.StatusOK; rr.Code != required {
		t.Errorf("handler returned wrong status code - want: %v, got: %v", required, rr.Code)
	}

	if infoRequest.Orchestration != handlers.OrchestrationIdentifier {
		t.Errorf("handler returned wrong orchestration - want: %v, got: %v", handlers.OrchestrationIdentifier, infoRequest.Orchestration)
	}

	if infoRequest.Provider != handlers.ProviderName {
		t.Errorf("handler returned wrong provider - want: %v, got: %v", handlers.ProviderName, infoRequest.Provider)
	}

	if infoRequest.Version.Release != infoTestVersion {
		t.Errorf("handler returned wrong release version - want: %v, got: %v", infoTestVersion, infoRequest.Version.Release)
	}

	if infoRequest.Version.SHA != infoTestSHA {
		t.Errorf("handler returned wrong SHA string - want: %v, got: %v", infoTestSHA, infoRequest.Version.SHA)
	}
}
