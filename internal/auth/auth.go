package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/usa4ev/gophermart/internal/auth/argon2hash"
)

type (
	Authenticator struct {
		pwdGetter passwordGetter
	}
	passwordGetter interface {
		GetPasswordHash(cxt context.Context, userName string) (string, string, error)
	}
)

//func NewAuthenticator() *Authenticator {
//	return &Authenticator{pwdGetter: p}
//}

//func (a *Authenticator) AuthMW(next http.Handler, p passwordGetter) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		user, pwd, ok := r.BasicAuth()
//		if !ok {
//			http.Error(w, "", http.StatusUnauthorized)
//			return
//		}
//
//		userID, pwdHash, err := a.pwdGetter.GetPasswordHash(user)
//		if err != nil {
//			http.Error(w, "", http.StatusInternalServerError)
//			return
//		}
//
//		ok, err = argon2hash.ComparePasswordAndHash(pwd, pwdHash)
//		if !ok {
//			http.Error(w, "", http.StatusUnauthorized)
//			return
//		}
//
//		token, err := session.Open(userID, 30*time.Second)
//
//		if err != nil {
//			w.WriteHeader(http.StatusInternalServerError)
//			return
//		}
//
//		w.Header().Add("Authorization", token)
//
//		ctx := context.WithValue(context.Background(), "userID", userID)
//		next.ServeHTTP(w, r.WithContext(ctx))
//	})
//}

func Login(ctx context.Context, userName, password string, p passwordGetter) (string, error) {
	userID, pwdHash, err := p.GetPasswordHash(ctx, userName)
	if err != nil {
		return "", err
	} else if userID == "" {
		return "", ErrUnathorized
	}

	ok, err := argon2hash.ComparePasswordAndHash(password, pwdHash)
	if !ok {
		return "", ErrUnathorized
	}

	return userID, nil
}

type (
	Registrator struct {
		usrAdder UserCheckAdder
	}
	UserCheckAdder interface {
		AddUser(ctx context.Context, username, hash string) (string, error)
		UserExists(ctx context.Context, username string) (bool, error)
	}
)

func RegisterUser(ctx context.Context, userName, password string, ua UserCheckAdder) (string, error) {
	err := validateUserName(ctx, userName, ua)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return "", fmt.Errorf("invalid user name: %v", err)
		}

		return "", err
	}

	hash, err := argon2hash.GenerateFromPassword(password, argon2hash.DefaultParams())

	if err != nil {
		return "", fmt.Errorf("failed to generate hash from password: %w", err)
	}

	return ua.AddUser(ctx, userName, hash)
}

func validateUserName(ctx context.Context, userName string, ua UserCheckAdder) error {
	if ok, err := ua.UserExists(ctx, userName); err != nil {
		return fmt.Errorf("failed to check if user already exists %w", err)
	} else if !ok {
		return ErrUserAlreadyExists
	}

	return nil
}
