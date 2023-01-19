package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi"

	"github.com/usa4ev/gophermart/internal/config"
	"github.com/usa4ev/gophermart/internal/server"
	"github.com/usa4ev/gophermart/internal/storage"
)

func main() {
	cfg := config.New()
	strg, err := storage.New(cfg.DatabaseDSN())
	if err != nil {
		log.Fatal(err.Error())
	}

	srv := server.New(strg, cfg)

	r := newRouter(srv)
	webSrv := &http.Server{Addr: cfg.RunAddress(), Handler: r}

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		call := <-sig

		webSrv.Close()

		fmt.Printf("graceful shutdown, got call: %v\n", call.String())
	}()

	// Run the server
	err = webSrv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err.Error())
	}
}

func newRouter(srv server.Server) http.Handler {
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
func defaultRoute(srv server.Server) func(r chi.Router) {
	return func(r chi.Router) {
		r.With(srv.GzipMW).Method(http.MethodPost, "/api/user/register", http.HandlerFunc(srv.Register))
		r.With(srv.GzipMW).Method(http.MethodPost, "/api/user/login", http.HandlerFunc(srv.Login))
		r.With(srv.GzipMW, srv.AuthorisationMW).Method(http.MethodPost, "/api/user/orders", http.HandlerFunc(srv.StoreOrder))
		r.With(srv.GzipMW, srv.AuthorisationMW).Method(http.MethodGet, "/api/user/orders", http.HandlerFunc(srv.LoadOrders))
		r.With(srv.GzipMW, srv.AuthorisationMW).Method(http.MethodGet, "/api/user/balance", http.HandlerFunc(srv.LoadBalance))
		r.With(srv.GzipMW, srv.AuthorisationMW).Method(http.MethodPost, "/api/user/balance/withdraw", http.HandlerFunc(srv.Withdraw))
		r.With(srv.GzipMW, srv.AuthorisationMW).Method(http.MethodGet, "/api/user/withdrawals", http.HandlerFunc(srv.LoadWithdrawals))
	}
}
