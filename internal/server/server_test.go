package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	argon2hash "github.com/usa4ev/gophermart/internal/auth/argon2hash"
	conf "github.com/usa4ev/gophermart/internal/config"
	"github.com/usa4ev/gophermart/internal/mocks"
	"github.com/usa4ev/gophermart/internal/storage/storageerrs"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegister(t *testing.T) {
	cfg := conf.New()
	ctrl := gomock.NewController(t)
	strg := mocks.NewMockStorage(ctrl)

	// New test server
	ts := newTestSrv(cfg, strg)
	cl := newTestClient(ts)

	tests := []struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		wantCode int
		exists   bool
	}{
		{
			Login:    "testUser",
			Password: "111",
			wantCode: http.StatusOK,
			exists:   false,
		},
		{
			Login:    "testUser2",
			Password: "111",
			wantCode: http.StatusConflict,
			exists:   true,
		},
	}

	h, _ := argon2hash.GenerateFromPassword("111", argon2hash.DefaultParams())
	fmt.Println(h)
	for _, tt := range tests {
		t.Run("Register", func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			enc := json.NewEncoder(buf)
			enc.Encode(tt)

			req, err := http.NewRequest(http.MethodPost, "http://"+cfg.RunAddress()+"/api/user/register", buf)
			require.NoError(t, err)

			strg.EXPECT().UserExists(gomock.Any(), tt.Login).Return(tt.exists, nil).Times(1)
			if !tt.exists {
				strg.EXPECT().AddUser(gomock.Any(), tt.Login, gomock.Any())
			}

			res, err := cl.Do(req)
			defer res.Body.Close()
			require.NoError(t, err)
			errTxt, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCode, res.StatusCode, string(errTxt))
		})
	}
}

func TestLogin(t *testing.T) {
	cfg := conf.New()
	ctrl := gomock.NewController(t)
	strg := mocks.NewMockStorage(ctrl)

	// New test server
	ts := newTestSrv(cfg, strg)
	cl := newTestClient(ts)

	tests := []struct {
		Login     string `json:"login"`
		Password  string `json:"password"`
		hash      string
		wantCode  int
		hashValid bool
	}{
		{
			Login:     "testUser",
			Password:  "111",
			hash:      "$argon2id$v=19$m=65536,t=1,p=2$Xzph17PGYg8gJrz+5IdMgw$VuX8PBmssQyIjk4a7o2xLM9NoEFKaDzz+zIegjl2plk",
			wantCode:  http.StatusOK,
			hashValid: true,
		},
		{
			Login:     "testUser2",
			Password:  "112", // hash given for a different pwd
			hash:      "$argon2id$v=19$m=65536,t=1,p=2$Xzph17PGYg8gJrz+5IdMgw$VuX8PBmssQyIjk4a7o2xLM9NoEFKaDzz+zIegjl2plk",
			wantCode:  http.StatusUnauthorized,
			hashValid: false,
		},
	}

	for _, tt := range tests {
		t.Run("login", func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			enc := json.NewEncoder(buf)
			enc.Encode(tt)

			req, err := http.NewRequest(http.MethodPost, "http://"+cfg.RunAddress()+"/api/user/login", buf)
			require.NoError(t, err)

			strg.EXPECT().GetPasswordHash(gomock.Any(), tt.Login).Return("userID", tt.hash, nil).Times(1)

			res, err := cl.Do(req)
			defer res.Body.Close()
			require.NoError(t, err)
			errTxt, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCode, res.StatusCode, string(errTxt))
			sessionFound := false
			for _, cookie := range res.Cookies() {
				if cookie.Name == "Authorization" {
					sessionFound = true
				}
			}

			require.Equal(t, tt.hashValid, sessionFound)
		})
	}
}

func TestStoreOrder(t *testing.T) {
	cfg := conf.New()
	ctrl := gomock.NewController(t)
	strg := mocks.NewMockStorage(ctrl)

	// New test server
	ts := newTestSrv(cfg, strg)
	cl := newTestClient(ts)

	tests := []struct {
		name       string
		order      string
		wantCode   int
		exists     bool
		conflict   bool
		orderValid bool
	}{
		{
			name:       "new valid",
			order:      "12345678903",
			wantCode:   http.StatusAccepted,
			exists:     false,
			conflict:   false,
			orderValid: true,
		},
		{
			name:       "new valid with spaces",
			order:      " 12345678903 \n",
			wantCode:   http.StatusAccepted,
			exists:     false,
			conflict:   false,
			orderValid: true,
		},
		{
			name:       "new invalid - wrong ctrl number",
			order:      "12345678904",
			wantCode:   http.StatusUnprocessableEntity,
			exists:     false,
			conflict:   false,
			orderValid: false,
		},
		{
			name:       "new invalid - string",
			order:      "non int",
			wantCode:   http.StatusUnprocessableEntity,
			exists:     false,
			conflict:   false,
			orderValid: false,
		},
		{
			name:       "order already exists",
			order:      "12345678903",
			wantCode:   http.StatusOK,
			exists:     true,
			conflict:   false,
			orderValid: true,
		},
		{
			name:       "order conflict",
			order:      "12345678903",
			wantCode:   http.StatusConflict,
			exists:     false,
			conflict:   true,
			orderValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte(tt.order))
			req, err := http.NewRequest(http.MethodPost, "http://"+cfg.RunAddress()+"/api/user/orders", buf)
			require.NoError(t, err)

			if tt.orderValid {
				number := strings.TrimSpace(tt.order)
				if tt.exists {
					strg.EXPECT().StoreOrder(gomock.Any(), number, "TestUser").Return(storageerrs.ErrOrderLoaded).Times(1)
				} else if tt.conflict {
					strg.EXPECT().StoreOrder(gomock.Any(), number, "TestUser").Return(storageerrs.ErrOrderExists).Times(1)
				} else {
					strg.EXPECT().StoreOrder(gomock.Any(), number, "TestUser").Return(nil).Times(1)
				}
			}

			res, err := cl.Do(req)
			defer res.Body.Close()
			require.NoError(t, err)

			errTxt, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCode, res.StatusCode, string(errTxt))
		})
	}
}

func TestLoadOrders(t *testing.T) {
	cfg := conf.New()
	ctrl := gomock.NewController(t)
	strg := mocks.NewMockStorage(ctrl)

	// New test server
	ts := newTestSrv(cfg, strg)
	cl := newTestClient(ts)

	tests := []struct {
		name       string
		order      string
		wantCode   int
		exists     bool
		conflict   bool
		orderValid bool
	}{
		{
			name:       "new valid",
			order:      "12345678903",
			wantCode:   http.StatusAccepted,
			exists:     false,
			conflict:   false,
			orderValid: true,
		},
		{
			name:       "new valid with spaces",
			order:      " 12345678903 \n",
			wantCode:   http.StatusAccepted,
			exists:     false,
			conflict:   false,
			orderValid: true,
		},
		{
			name:       "new invalid - wrong ctrl number",
			order:      "12345678904",
			wantCode:   http.StatusUnprocessableEntity,
			exists:     false,
			conflict:   false,
			orderValid: false,
		},
		{
			name:       "new invalid - string",
			order:      "non int",
			wantCode:   http.StatusUnprocessableEntity,
			exists:     false,
			conflict:   false,
			orderValid: false,
		},
		{
			name:       "order already exists",
			order:      "12345678903",
			wantCode:   http.StatusOK,
			exists:     true,
			conflict:   false,
			orderValid: true,
		},
		{
			name:       "order conflict",
			order:      "12345678903",
			wantCode:   http.StatusConflict,
			exists:     false,
			conflict:   true,
			orderValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte(tt.order))
			req, err := http.NewRequest(http.MethodPost, "http://"+cfg.RunAddress()+"/api/user/orders", buf)
			require.NoError(t, err)

			if tt.orderValid {
				number := strings.TrimSpace(tt.order)
				if tt.exists {
					strg.EXPECT().StoreOrder(gomock.Any(), number, "TestUser").Return(storageerrs.ErrOrderLoaded).Times(1)
				} else if tt.conflict {
					strg.EXPECT().StoreOrder(gomock.Any(), number, "TestUser").Return(storageerrs.ErrOrderExists).Times(1)
				} else {
					strg.EXPECT().StoreOrder(gomock.Any(), number, "TestUser").Return(nil).Times(1)
				}
			}

			res, err := cl.Do(req)
			defer res.Body.Close()
			require.NoError(t, err)

			errTxt, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCode, res.StatusCode, string(errTxt))
		})
	}
}

func newTestClient(ts *httptest.Server) *http.Client {
	cl := ts.Client()

	return cl
}

func newTestSrv(cfg *conf.Config, strg *mocks.MockStorage) *httptest.Server {
	s := New(strg, cfg)
	r := newRouter(s)

	l, err := net.Listen("tcp", cfg.RunAddress())
	if err != nil {
		panic(fmt.Sprintf("httptest: failed to listen on %v: %v", cfg.RunAddress(), err))
	}

	ts := httptest.NewUnstartedServer(r)
	ts.Listener = l
	ts.Start()

	return ts
}

func newRouter(srv Server) http.Handler {
	r := chi.NewRouter()
	r.Route("/", defaultRoute(srv))

	return r
}

// POST /api/user/register — регистрация пользователя;
// POST /api/user/login — аутентификация пользователя;
// POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
// GET /api/user/orders — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
// GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
// POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
// GET /api/user/balance/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.
func defaultRoute(srv Server) func(r chi.Router) {
	return func(r chi.Router) {
		r.With(srv.GzipMW).Method(http.MethodPost, "/api/user/register", http.HandlerFunc(srv.Register))
		r.With(srv.GzipMW).Method(http.MethodPost, "/api/user/login", http.HandlerFunc(srv.Login))
		r.With(authMock).Method(http.MethodPost, "/api/user/orders", http.HandlerFunc(srv.StoreOrder))
		r.With(srv.GzipMW, srv.AuthorisationMW).Method(http.MethodGet, "/api/user/orders", http.HandlerFunc(srv.LoadOrders))
		r.With(srv.GzipMW, srv.AuthorisationMW).Method(http.MethodGet, "/api/user/balance", http.HandlerFunc(srv.LoadBalance))
		r.With(srv.GzipMW, srv.AuthorisationMW).Method(http.MethodPost, "/api/user/balance/withdraw", http.HandlerFunc(srv.Withdraw))
		r.With(srv.GzipMW, srv.AuthorisationMW).Method(http.MethodGet, "/api/user/balance/withdrawals", http.HandlerFunc(srv.LoadWithdrawals))
	}
}

func authMock(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "userID", "TestUser")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
