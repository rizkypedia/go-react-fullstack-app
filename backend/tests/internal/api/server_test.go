package api_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fullstack-app/backend/internal/api"
)

func testHandler() http.Handler {
	return api.NewServer(api.Config{
		JWTSecret:    []byte("test-signing-secret-that-is-at-least-32-bytes"),
		DemoUsername: "demo", DemoPassword: "correct-horse",
		FrontendOrigin: "http://localhost:5173",
	})
}

func TestEchoRequiresValidToken(t *testing.T) {
	handler := testHandler()

	request := httptest.NewRequest(http.MethodPost, "/api/echo", bytes.NewBufferString(`{"value":"hello"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated request status = %d, want %d", response.Code, http.StatusUnauthorized)
	}

	token, err := api.IssueToken([]byte("test-signing-secret-that-is-at-least-32-bytes"), "demo")
	if err != nil {
		t.Fatal(err)
	}
	request = httptest.NewRequest(http.MethodPost, "/api/echo", bytes.NewBufferString(`{"value":"hello"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("authorized request status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body)
	}
	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["value"] != "hello" {
		t.Fatalf("echo value = %q, want %q", body["value"], "hello")
	}
}

func TestLoginIssuesTokenOnlyForConfiguredCredentials(t *testing.T) {
	handler := testHandler()
	for _, test := range []struct {
		name     string
		password string
		wantCode int
	}{
		{name: "valid", password: "correct-horse", wantCode: http.StatusOK},
		{name: "invalid", password: "wrong", wantCode: http.StatusUnauthorized},
	} {
		t.Run(test.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{"username": "demo", "password": test.password})
			request := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.wantCode {
				t.Fatalf("login status = %d, want %d", response.Code, test.wantCode)
			}
		})
	}
}

func TestTokenExpiresAfter12Hours(t *testing.T) {
	token, err := api.IssueToken([]byte("test-signing-secret-that-is-at-least-32-bytes"), "demo")
	if err != nil {
		t.Fatal(err)
	}
	parts := bytes.Split([]byte(token), []byte("."))
	claimsJSON, err := base64.RawURLEncoding.DecodeString(string(parts[1]))
	if err != nil {
		t.Fatal(err)
	}
	var claims api.TokenClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		t.Fatal(err)
	}
	if got, want := claims.ExpiresAt-claims.IssuedAt, int64((12 * time.Hour).Seconds()); got != want {
		t.Fatalf("token lifetime = %d seconds, want %d", got, want)
	}
}

func TestLogoutRevokesToken(t *testing.T) {
	handler := testHandler()
	token, err := api.IssueToken([]byte("test-signing-secret-that-is-at-least-32-bytes"), "demo")
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want %d; body: %s", response.Code, http.StatusNoContent, response.Body)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/echo", bytes.NewBufferString(`{"value":"hello"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("request with logged-out token status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestCORSAllowsOnlyConfiguredOrigin(t *testing.T) {
	handler := testHandler()
	request := httptest.NewRequest(http.MethodOptions, "/api/echo", nil)
	request.Header.Set("Origin", "https://untrusted.example")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("untrusted preflight status = %d, want %d", response.Code, http.StatusForbidden)
	}

	request = httptest.NewRequest(http.MethodOptions, "/api/echo", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("trusted preflight status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("allow-origin = %q, want configured origin", got)
	}
}
