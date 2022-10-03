package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/usa4ev/gophermart/internal/auth"
	"github.com/usa4ev/gophermart/internal/session"
	"net/http"
)

type (
	credentials struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	srvCtxKey string
)

// Register handler adds a new user if one does not exist and opens a new session
func (srv Server) Register(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != "" && ct != ctJSON {
		http.Error(w, fmt.Sprintf("unexpected content-type %v", ct), http.StatusBadRequest)

		return
	}

	defer r.Body.Close()

	cred := credentials{}

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&cred)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode a message: %v", err), http.StatusBadRequest)

		return
	}

	userID, err := auth.RegisterUser(r.Context(), cred.Login, cred.Password, srv.strg)

	if errors.Is(err, auth.ErrUserAlreadyExists) {
		http.Error(w, err.Error(), http.StatusConflict)

		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("failed to create user: %v", err), http.StatusInternalServerError)

		return
	}

	token, expiresAt, err := session.Open(userID, srv.cfg.SessionLifetime())

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create new user: %v", err), http.StatusInternalServerError)

		return
	}

	http.SetCookie(w, &http.Cookie{Name: "Authorization", Value: token, Expires: expiresAt})
}

// Register handler adds a new user if one does not exist and opens a new session
func (srv Server) Login(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	if ct != "" && ct != ctJSON {
		http.Error(w, fmt.Sprintf("unexpected content-type %v", ct), http.StatusBadRequest)

		return
	}

	defer r.Body.Close()

	cred := credentials{}

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&cred)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to decode a message: %v", err), http.StatusBadRequest)

		return
	}

	userID, err := auth.Login(r.Context(), cred.Login, cred.Password, srv.strg)

	if errors.Is(err, auth.ErrUnathorized) {
		http.Error(w, err.Error(), http.StatusUnauthorized)

		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("authentication failed: %v", err), http.StatusInternalServerError)

		return
	}

	token, expiresAt, err := session.Open(userID, srv.cfg.SessionLifetime())

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create user: %v", err), http.StatusInternalServerError)

		return
	}

	http.SetCookie(w, &http.Cookie{Name: "Authorization", Value: token, Expires: expiresAt})
}

func (srv Server) AuthorisationMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("Authorisation")
		if err != nil {
			if err == http.ErrNoCookie {
				w.WriteHeader(http.StatusUnauthorized)

				return
			}

			w.WriteHeader(http.StatusBadRequest)

			return
		}

		tokenString := c.Value

		userID, err := session.Verify(tokenString)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(context.Background(), srvCtxKey("userID"), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
