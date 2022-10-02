package server

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/golang/mock/mockgen/model"
	"github.com/usa4ev/gophermart/internal/orders"
	"github.com/usa4ev/gophermart/internal/storage/storageerrs"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	ctPlainText = "text/plain"
	ctJSON      = "application/json"
)

type (
	Server struct {
		strg    Storage
		cfg     config
		running bool
	}

	config interface {
		SessionLifetime() time.Duration
		AccrualSysAddr() string
	}

	Storage interface {
		StoreOrder(ctx context.Context, orderNum, userID string) error
		LoadOrders(ctx context.Context, userID string) ([]byte, error)
		LoadBalance(ctx context.Context, userID string) (float64, float64, error)
		Withdraw(ctx context.Context, userID, number string, sum float64) error
		LoadWithdrawals(ctx context.Context, userID string) ([]byte, error)
		OrdersToProcess(ctx context.Context) (map[string]string, error)
		UpdateStatuses(ctx context.Context, batch []orders.Status) error
		UpdateBalances(ctx context.Context) error
		AddUser(ctx context.Context, username, hash string) (string, error)
		UserExists(ctx context.Context, userName string) (bool, error)
		GetPasswordHash(ctx context.Context, userName string) (string, string, error)
	}
)

// New return new Server with started background processes
func New(strg Storage, cfg config) Server {
	srv := Server{
		strg: strg,
		cfg:  cfg,
	}

	srv.start()

	return srv
}
func (srv Server) updateStatuses(servicePath string) {
	if servicePath == "" {
		log.Fatal("Accrual service path is not defined")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ordersToProcess, err := srv.strg.OrdersToProcess(ctx)

	batch := make([]orders.Status, 0)

	for order, currentStatus := range ordersToProcess {

		res, err := http.Get(servicePath + "/api/orders/" + order)
		if err != nil {
			log.Printf("failed to access accrual system: %v\n", err)
			continue
		}

		if res.StatusCode == http.StatusTooManyRequests {
			// limit's been reached, so we're done here for now
			break
		} else if res.StatusCode != http.StatusOK {
			body, err := io.ReadAll(res.Body)
			if err != nil {
				log.Printf("failed to read accural system error response: %v\n", err)
			}

			log.Printf("unexpected response from accrual system. code: %v message: %v", res.StatusCode, string(body))
			continue
		}

		status := orders.Status{}
		dec := json.NewDecoder(res.Body)
		dec.Decode(&status)

		if currentStatus == status.Status {
			// nothing changed, needless to update this order
			continue
		}

		batch = append(batch, status)
	}

	err = srv.strg.UpdateStatuses(ctx, batch)
	if err != nil {
		if err != nil {
			log.Printf("%v", err)
		}
	}
}

func (srv Server) updateBalance() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err := srv.strg.UpdateBalances(ctx)
	if err != nil {
		log.Printf("%v", err)
	}
}

func (srv Server) updBalances() {
	// once a day
	timer := time.NewTimer(time.Until(time.Now().AddDate(0, 0, 1).Round(24 * time.Hour)))

	for {
		<-timer.C
		srv.updateBalance()
		timer = time.NewTimer(time.Until(time.Now().AddDate(0, 0, 1).Round(24 * time.Hour)))
	}
}

func (srv Server) updStatuses() {
	// every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)

	for {
		<-ticker.C
		srv.updateStatuses(srv.cfg.AccrualSysAddr())
	}
}

func (srv Server) start() error {
	if srv.running {
		return fmt.Errorf("server is already running")
	}

	go srv.updBalances()
	go srv.updStatuses()

	return nil
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (srv Server) GzipMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		supportedEncoding := r.Header.Values("Accept-Encoding")
		if len(supportedEncoding) > 0 {
			for _, v := range supportedEncoding {
				if v == "gzip" {
					writer, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)

						return
					}
					defer writer.Close()
					w.Header().Set("Content-Encoding", "gzip")
					gzipW := gzipWriter{w, writer}
					next.ServeHTTP(gzipW, r)

					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (srv Server) StoreOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "request context is missing user ID", http.StatusInternalServerError)

		return
	}

	ct := r.Header.Get("Content-Type")
	if ct != "" && ct != ctPlainText {
		http.Error(w, fmt.Sprintf("unexpected content-type %v", ct), http.StatusBadRequest)

		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read request message: %v", err), http.StatusInternalServerError)

		return
	}

	message := strings.TrimSpace(string(body))

	if !orders.OrderNumValid(message) {
		http.Error(w, fmt.Sprintf("invalid order number: %v", err), http.StatusUnprocessableEntity)

		return
	}

	err = srv.strg.StoreOrder(r.Context(), message, userID)
	if errors.Is(err, storageerrs.ErrOrderExists) {
		http.Error(w, err.Error(), http.StatusConflict)

		return
	} else if errors.Is(err, storageerrs.ErrOrderLoaded) {
		// return 200 OK: order already has been loaded by this user
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("failed to save new order: %v", err), http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (srv Server) LoadOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "request context is missing user ID", http.StatusInternalServerError)

		return
	}

	res, err := srv.strg.LoadOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get orders from database: %v", err), http.StatusInternalServerError)

		return
	}

	w.Write(res)
}
