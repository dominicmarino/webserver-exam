package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// helper to set up test infra
func setupTestRouter() (*webServerStorage, *http.ServeMux) {
	storage := &webServerStorage{
		contentStorage:  make(map[string]map[string]string),
		metadataStorage: make(map[string]map[string]string),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /objects/{bucket}/{id}", storage.handleGet)
	mux.HandleFunc("PUT /objects/{bucket}/{id}", storage.handlePut)
	mux.HandleFunc("DELETE /objects/{bucket}/{id}", storage.handleDelete)
	return storage, mux
}

func TestHandlePutSimple(t *testing.T) {
	_, mux := setupTestRouter()

	body := []byte("hello world")
	req := httptest.NewRequest("PUT", "/objects/b1/o1", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"id":"o1"`) {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
}

func TestHandleGetSimple(t *testing.T) {
	storage, mux := setupTestRouter()

	storage.contentStorage["b1"] = map[string]string{"test-hash": "test-data"}
	storage.metadataStorage["b1"] = map[string]string{"o1": "test-hash"}

	req := httptest.NewRequest("GET", "/objects/b1/o1", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "Status: 200 OK {test-data}" {
		t.Errorf("expected 'Status: 200 OK {test-data}', got %s", rr.Body.String())
	}
}

func TestHandleDeleteSimple(t *testing.T) {
	storage, mux := setupTestRouter()

	storage.metadataStorage["b1"] = map[string]string{"o1": "h1"}
	storage.contentStorage["b1"] = map[string]string{"h1": "test-data"}

	req := httptest.NewRequest("DELETE", "/objects/b1/o1", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	if len(storage.metadataStorage["b1"]) != 0 || len(storage.contentStorage["b1"]) != 0 {
		t.Error("item or content was not fully deleted from storage")
	}
}

func TestHandleDeleteWithDeduplication(t *testing.T) {
	storage, mux := setupTestRouter()

	storage.metadataStorage["b1"] = map[string]string{
		"o1": "shared-hash",
		"o2": "shared-hash",
	}
	storage.contentStorage["b1"] = map[string]string{"shared-hash": "test-data"}

	// delete id o1
	req := httptest.NewRequest("DELETE", "/objects/b1/o1", nil)
	mux.ServeHTTP(httptest.NewRecorder(), req)

	// o1 should be gone from metadata
	if _, ok := storage.metadataStorage["b1"]["o1"]; ok {
		t.Error("o1 metadata should be deleted")
	}
	// we should not have deleted the actual content since it's still referenced
	if _, ok := storage.contentStorage["b1"]["shared-hash"]; !ok {
		t.Error("content should still exist because o2 is using it")
	}

	// delete o2
	req2 := httptest.NewRequest("DELETE", "/objects/b1/o2", nil)
	mux.ServeHTTP(httptest.NewRecorder(), req2)

	// check if content is deleted
	if len(storage.contentStorage["b1"]) != 0 {
		t.Error("contentStorage should be empty now")
	}
}

func TestHandlePutBadURL(t *testing.T) {
	_, mux := setupTestRouter()

	// Missing ID in URL
	req := httptest.NewRequest("PUT", "/objects/b1", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 for mismatched route, got %d", rr.Code)
	}
}

func TestHandleGetBadURL(t *testing.T) {
	_, mux := setupTestRouter()

	req := httptest.NewRequest("GET", "/objects/onlyonepart", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleDeleteBadURL(t *testing.T) {
	_, mux := setupTestRouter()

	req := httptest.NewRequest("DELETE", "/objects/too/many/parts/here", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleGetNotPresent(t *testing.T) {
	_, mux := setupTestRouter()

	req := httptest.NewRequest("GET", "/objects/bucket/notexists", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	if rr.Body.String() != "Status 400 Not Found" {
		t.Errorf("expected \"Status 400 Not Found\", got %s\n", rr.Body.String())
	}
}

func TestHandleDeleteNotPreset(t *testing.T) {
	_, mux := setupTestRouter()

	req := httptest.NewRequest("DELETE", "/objects/bucket/notexists", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	if rr.Body.String() != "Status 400 Not Found" {
		t.Errorf("expected \"Status 400 Not Found\", got %s\n", rr.Body.String())
	}
}
