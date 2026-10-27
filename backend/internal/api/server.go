package api

import (
	"crypto/sha256"
	"fullstack-app/backend/utilities"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Server struct {
	config       Config
	mux          *http.ServeMux
	revokedMu    sync.Mutex
	revokedToken map[[sha256.Size]byte]time.Time
}

func NewServer(config Config) http.Handler {
	server := &Server{
		config:       config,
		mux:          http.NewServeMux(),
		revokedToken: make(map[[sha256.Size]byte]time.Time),
	}
	server.mux.HandleFunc("POST /api/auth/login", server.login)
	server.mux.Handle("POST /api/auth/logout", server.requireAuth(http.HandlerFunc(server.logout)))
	server.mux.Handle("POST /api/echo", server.requireAuth(http.HandlerFunc(server.echo)))
	server.mux.HandleFunc("GET /api/health", server.health)
	return server.cors(server.mux)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := utilities.DecodeJSON(w, r, &request); err != nil {
		utilities.WriteError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	if !credentialsMatch(request.Username, s.config.DemoUsername) ||
		!credentialsMatch(request.Password, s.config.DemoPassword) {
		utilities.WriteError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	token, err := IssueToken(s.config.JWTSecret, request.Username)
	if err != nil {
		utilities.WriteError(w, http.StatusInternalServerError, "could not issue access token")
		return
	}
	utilities.WriteJSON(w, http.StatusOK, map[string]string{"access_token": token, "token_type": "Bearer"})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	_, token, _ := strings.Cut(r.Header.Get("Authorization"), " ")
	s.revokeToken(token)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) echo(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Value string `json:"value"`
	}
	if err := utilities.DecodeJSON(w, r, &request); err != nil {
		utilities.WriteError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	if strings.TrimSpace(request.Value) == "" {
		utilities.WriteError(w, http.StatusBadRequest, "value must not be empty")
		return
	}
	utilities.WriteJSON(w, http.StatusOK, map[string]string{"value": request.Value})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	utilities.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		scheme, token, found := strings.Cut(header, " ")
		if !found || !strings.EqualFold(scheme, "Bearer") || token == "" ||
			strings.ContainsAny(token, " \t\r\n") || !validateToken(s.config.JWTSecret, token) ||
			s.isTokenRevoked(token) {
			utilities.WriteError(w, http.StatusUnauthorized, "missing or invalid bearer token")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) revokeToken(token string) {
	hash := sha256.Sum256([]byte(token))
	s.revokedMu.Lock()
	defer s.revokedMu.Unlock()

	now := time.Now()
	for revokedHash, expiry := range s.revokedToken {
		if !expiry.After(now) {
			delete(s.revokedToken, revokedHash)
		}
	}
	s.revokedToken[hash] = now.Add(tokenLifetime)
}

func (s *Server) isTokenRevoked(token string) bool {
	hash := sha256.Sum256([]byte(token))
	s.revokedMu.Lock()
	defer s.revokedMu.Unlock()

	expiry, revoked := s.revokedToken[hash]
	if revoked && !expiry.After(time.Now()) {
		delete(s.revokedToken, hash)
		return false
	}
	return revoked
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == s.config.FrontendOrigin {
			w.Header().Set("Access-Control-Allow-Origin", s.config.FrontendOrigin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		}
		if r.Method == http.MethodOptions {
			if origin == "" || origin != s.config.FrontendOrigin {
				utilities.WriteError(w, http.StatusForbidden, "origin is not allowed")
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
