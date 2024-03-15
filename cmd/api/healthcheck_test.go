package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kvnloughead/greenlight/internal/assert"
)

func TestHealthcheck(t *testing.T) {
	app := &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		config: config{env: "testing"},
	}

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	rs, err := ts.Client().Get(ts.URL + "/v1/healthcheck")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, rs.StatusCode, http.StatusOK)

	var response struct {
		Status     string `json:"status"`
		SystemInfo struct {
			Environment string `json:"environment"`
			Version     string `json:"version"`
		} `json:"system_info"`
	}

	defer rs.Body.Close()
	err = json.NewDecoder(rs.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, response.Status, "available")
	assert.Equal(t, response.SystemInfo.Environment, "testing")
	assert.Equal(t, response.SystemInfo.Version, version)
}
