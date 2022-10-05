package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/usa4ev/gophermart/internal/orders"
)

type balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (srv Server) LoadBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(srvCtxKey("userID")).(string)
	if !ok {
		http.Error(w, "request context is missing user ID", http.StatusInternalServerError)

		return
	}

	total, withdrawn, err := srv.strg.LoadBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get balance from database: %v", err), http.StatusInternalServerError)

		return
	}

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err = enc.Encode(balance{total, withdrawn})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to encode answer: %v", err), http.StatusInternalServerError)

		return
	}

	w.Header().Add("Content-Type", ctJSON)

	w.Write(buf.Bytes())
}

func (srv Server) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(srvCtxKey("userID")).(string)
	if !ok {
		errtxt := "request context is missing user ID"
		http.Error(w, errtxt, http.StatusInternalServerError)
		log.Printf(errtxt + "\n")

		return
	}

	ct := r.Header.Get("Content-Type")
	if ct != "" && ct != ctJSON {
		errtxt := fmt.Sprintf("unexpected content-type %v", ct)
		http.Error(w, errtxt, http.StatusBadRequest)
		log.Printf(errtxt + "\n")

		return
	}

	defer r.Body.Close()

	op := orders.Withdrawal{}

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&op)
	if err != nil {
		errtxt := fmt.Sprintf("failed to decode a message: %v", err)
		http.Error(w, errtxt, http.StatusBadRequest)
		log.Printf(errtxt + "\n")

		return
	}

	if !orders.OrderNumValid(op.Order) {
		errtxt := fmt.Sprintf("invalid order number: %v", err)
		http.Error(w, errtxt, http.StatusUnprocessableEntity)
		log.Printf(errtxt + "\n")

		return
	}

	if total, _, err := srv.strg.LoadBalance(r.Context(), userID); err != nil {
		errtxt := fmt.Sprintf("failed to get balnce from database: %v", err)
		http.Error(w, errtxt, http.StatusInternalServerError)
		log.Printf(errtxt + "\n")

		return
	} else if total < op.Sum {
		http.Error(w, "not enough coins", http.StatusPaymentRequired)

		return
	}

	err = srv.strg.Withdraw(r.Context(), userID, op.Order, op.Sum)
	if err != nil {
		errtxt := fmt.Sprintf("failed to process a withdraw operation: %v", err)
		http.Error(w, errtxt, http.StatusInternalServerError)
		log.Printf(errtxt + "\n")

		return
	}
}

func (srv Server) LoadWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(srvCtxKey("userID")).(string)
	if !ok {
		errtxt := "request context is missing user ID"
		http.Error(w, errtxt, http.StatusInternalServerError)
		log.Printf(errtxt + "\n")

		return
	}

	res, err := srv.strg.LoadWithdrawals(r.Context(), userID)
	if err != nil {
		errtxt := fmt.Sprintf("failed to get withdrawals from database: %v", err)
		http.Error(w, errtxt, http.StatusInternalServerError)
		log.Printf(errtxt + "\n")

		return
	}

	w.Header().Add("Content-Type", ctJSON)

	w.Write(res)
}
