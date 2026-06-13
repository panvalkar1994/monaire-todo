//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"todo/internal/config"
	"todo/internal/database"
	"todo/internal/handler"
	appmigrate "todo/internal/migrate"
	gormrepo "todo/internal/repository/gorm"
	"todo/internal/service"

	"github.com/spf13/viper"
)

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	viper.Reset()
	config.InitViper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	viper.Set("server.addr", ":8080")
	viper.Set("server.gin_mode", "test")
	viper.Set("database.driver", "sqlite")
	dbFile := filepath.Join(t.TempDir(), "test.db")
	viper.Set("database.dsn", "file:"+dbFile+"?cache=shared")
	viper.Set("migrations.path", filepath.Join(root, "migrations"))

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if err := appmigrate.Up(cfg); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	db, err := database.Open(cfg.Database)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	svc := service.NewTodoService(gormrepo.NewTodoRepository(db))
	return httptest.NewServer(handler.NewRouter(svc))
}

func TestTodoCRUDFlow(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()
	base := ts.URL + "/api/v1/todos"

	create := func(text, due string) string {
		body, _ := json.Marshal(map[string]string{"text": text, "due_date": due})
		resp, err := http.Post(base, "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create status: %d", resp.StatusCode)
		}
		var out map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&out)
		return out["id"].(string)
	}

	id1 := create("later", "2026-06-20")
	id2 := create("sooner", "2026-06-10")

	resp, err := http.Get(base)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	defer resp.Body.Close()
	var list []map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&list)
	if len(list) != 2 || list[0]["due_date"] != "2026-06-10" {
		t.Fatalf("unexpected list order: %+v", list)
	}

	patchBody, _ := json.Marshal(map[string]bool{"completed": true})
	req, _ := http.NewRequest(http.MethodPatch, base+"/"+id2, bytes.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	patchResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	patchResp.Body.Close()

	resp, _ = http.Get(base)
	defer resp.Body.Close()
	_ = json.NewDecoder(resp.Body).Decode(&list)
	if len(list) != 1 || list[0]["id"] != id1 {
		t.Fatalf("expected only incomplete todo: %+v", list)
	}

	delReq, _ := http.NewRequest(http.MethodDelete, base+"/"+id1, nil)
	delResp, err := http.DefaultClient.Do(delReq)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	delResp.Body.Close()
	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status: %d", delResp.StatusCode)
	}

	getResp, _ := http.Get(base + "/" + id1)
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", getResp.StatusCode)
	}
}

func TestPutReplaceUpdatesTodo(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()
	base := ts.URL + "/api/v1/todos"

	createBody, _ := json.Marshal(map[string]string{"text": "original", "due_date": "2026-06-10"})
	createResp, err := http.Post(base, "application/json", bytes.NewReader(createBody))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer createResp.Body.Close()
	var created map[string]interface{}
	_ = json.NewDecoder(createResp.Body).Decode(&created)
	id := created["id"].(string)

	putBody, _ := json.Marshal(map[string]interface{}{
		"text":      "replaced",
		"due_date":  "2026-06-25",
		"completed": true,
	})
	putReq, _ := http.NewRequest(http.MethodPut, base+"/"+id, bytes.NewReader(putBody))
	putReq.Header.Set("Content-Type", "application/json")
	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	defer putResp.Body.Close()
	if putResp.StatusCode != http.StatusOK {
		t.Fatalf("put status: %d", putResp.StatusCode)
	}

	getResp, err := http.Get(base + "/" + id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer getResp.Body.Close()
	var got map[string]interface{}
	_ = json.NewDecoder(getResp.Body).Decode(&got)
	if got["text"] != "replaced" || got["due_date"] != "2026-06-25" || got["completed"] != true {
		t.Fatalf("unexpected todo after put: %+v", got)
	}
}

func TestListIncludeCompleted(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()
	base := ts.URL + "/api/v1/todos"

	post := func(text, due string) string {
		body, _ := json.Marshal(map[string]string{"text": text, "due_date": due})
		resp, err := http.Post(base, "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer resp.Body.Close()
		var out map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&out)
		return out["id"].(string)
	}

	openID := post("open task", "2026-06-10")
	doneID := post("done task", "2026-06-15")

	patchBody, _ := json.Marshal(map[string]bool{"completed": true})
	patchReq, _ := http.NewRequest(http.MethodPatch, base+"/"+doneID, bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchResp, err := http.DefaultClient.Do(patchReq)
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	patchResp.Body.Close()

	defaultList, err := http.Get(base)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	defer defaultList.Body.Close()
	var incomplete []map[string]interface{}
	_ = json.NewDecoder(defaultList.Body).Decode(&incomplete)
	if len(incomplete) != 1 || incomplete[0]["id"] != openID {
		t.Fatalf("default list: %+v", incomplete)
	}

	allList, err := http.Get(base + "?include_completed=true")
	if err != nil {
		t.Fatalf("list include_completed: %v", err)
	}
	defer allList.Body.Close()
	var all []map[string]interface{}
	_ = json.NewDecoder(allList.Body).Decode(&all)
	if len(all) != 2 {
		t.Fatalf("expected 2 todos with include_completed, got %d", len(all))
	}
	if all[0]["due_date"] != "2026-06-10" || all[1]["due_date"] != "2026-06-15" {
		t.Fatalf("unexpected sort with include_completed: %+v", all)
	}
}

func TestListIncludeCompletedQueryValues(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()
	base := ts.URL + "/api/v1/todos"

	post := func(text, due string) string {
		body, _ := json.Marshal(map[string]string{"text": text, "due_date": due})
		resp, err := http.Post(base, "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer resp.Body.Close()
		var out map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&out)
		return out["id"].(string)
	}

	post("open task", "2026-06-10")
	doneID := post("done task", "2026-06-15")

	patchBody, _ := json.Marshal(map[string]bool{"completed": true})
	patchReq, _ := http.NewRequest(http.MethodPatch, base+"/"+doneID, bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchResp, err := http.DefaultClient.Do(patchReq)
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	patchResp.Body.Close()

	assertListCount := func(t *testing.T, query string, want int) {
		t.Helper()
		resp, err := http.Get(base + query)
		if err != nil {
			t.Fatalf("list %q: %v", query, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("list %q status: %d", query, resp.StatusCode)
		}
		var list []map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&list)
		if len(list) != want {
			t.Fatalf("list %q: got %d todos, want %d", query, len(list), want)
		}
	}

	assertListCount(t, "", 1)
	assertListCount(t, "?include_completed=", 1)
	assertListCount(t, "?include_completed=false", 1)
	assertListCount(t, "?include_completed=0", 1)
	assertListCount(t, "?include_completed=no", 1)
	assertListCount(t, "?include_completed=true", 2)
	assertListCount(t, "?include_completed=1", 2)
	assertListCount(t, "?include_completed=yes", 2)
}

func TestListIncludeCompletedInvalidReturns417(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()
	base := ts.URL + "/api/v1/todos"

	for _, query := range []string{"?include_completed=maybe", "?include_completed=2", "?include_completed=t"} {
		resp, err := http.Get(base + query)
		if err != nil {
			t.Fatalf("list %q: %v", query, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusExpectationFailed {
			t.Fatalf("list %q status: %d, want 417", query, resp.StatusCode)
		}
		var body map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&body)
		want := "include_completed allowed values: true|false|empty"
		if body["error"] != want {
			t.Fatalf("list %q error: %q, want %q", query, body["error"], want)
		}
	}
}
