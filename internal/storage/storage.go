package storage

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/usa4ev/gophermart/internal/auth"
	"github.com/usa4ev/gophermart/internal/orders"
	"github.com/usa4ev/gophermart/internal/storage/storageerrs"

	_ "github.com/jackc/pgx/stdlib"
)

type (
	Database struct {
		*sql.DB
	}
)

func New(dsn string) (Database, error) {
	var (
		db  Database
		err error
	)

	db.DB, err = sql.Open("pgx", dsn)
	if err != nil {
		return db, fmt.Errorf("cannot connect to Database: %w", err)
	}

	err = db.initDB()
	if err != nil {
		return db, fmt.Errorf("cannot init Database: %w", err)
	}

	return db, nil
}

func (db Database) initDB() error {
	query := `CREATE TABLE IF NOT EXISTS users (
				id VARCHAR(100) PRIMARY KEY,
				username VARCHAR(256) not null,
                pwdhash VARCHAR(256) not null);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	query = `CREATE TABLE IF NOT EXISTS orders (
					number VARCHAR(100) PRIMARY KEY UNIQUE,
					ts timestamptz not null,
					uploaded date not null,
					customer varchar(100) not null,
					income float not null,
					status VARCHAR(30),
					FOREIGN KEY (customer)
				REFERENCES users (id));`

	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	query = `CREATE TABLE IF NOT EXISTS withdrawals (
					number VARCHAR(100) PRIMARY KEY UNIQUE,
					ts timestamptz not null,
					processed date not null,
					customer varchar(100) not null,
					withdraw float not null,
					FOREIGN KEY (customer)
				REFERENCES users (id));`

	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	query = `CREATE TABLE IF NOT EXISTS balances (
					customer varchar(100) primary key,
					ts timestamptz not null,
					balance float not null,
					total_withdraw float not null,
					FOREIGN KEY (customer)
				REFERENCES users (id));`

	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	return err
}

// AddUser adds new row to Database and return new user ID or error if addition failed
func (db Database) AddUser(ctx context.Context, username, hash string) (string, error) {
	id := uuid.New().String()

	query := `INSERT INTO users(id, username, pwdhash) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING;`

	rowsAffected, err := db.execInsUpdStatement(ctx, query, id, username, hash)

	if err != nil {
		return "", err
	} else if rowsAffected == 0 {
		return "", auth.ErrUserAlreadyExists
	}

	query = `INSERT INTO balances(customer, ts, balance, total_withdraw) VALUES ($1, now()::timestamptz, 0, 0) ON CONFLICT (customer) DO NOTHING;`

	_, err = db.execInsUpdStatement(ctx, query, id)

	if err != nil {
		return "", err
	}

	return id, nil
}

// UserExists returns true if user found by given userName or false otherwise
func (db Database) UserExists(ctx context.Context, userName string) (bool, error) {
	var exists bool

	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"

	err := db.QueryRowContext(ctx, query, userName).Scan(&exists)

	if !exists || errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return true, nil
}

// GetPasswordHash returns user ID and pwd hash found by given userName or empty string as user ID if user not found
func (db Database) GetPasswordHash(ctx context.Context, userName string) (string, string, error) {
	var userID, hash string

	query := "SELECT id, pwdhash FROM users WHERE username = $1"

	err := db.QueryRowContext(ctx, query, userName).Scan(&userID, &hash)

	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	} else if err != nil {
		return "", "", fmt.Errorf("failed to get a password hash from Database: %w", err)
	}

	return userID, hash, nil
}

func (db Database) execInsUpdStatement(ctx context.Context, query string, args ...interface{}) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("error when executing query context %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error when finding rows affected %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

func (db Database) StoreOrder(ctx context.Context, orderNum, userID string) error {
	query := "INSERT INTO orders(number, customer, ts, uploaded, status, income) VALUES ($1, $2, now()::timestamptz, now(), 'NEW', 0) ON CONFLICT (number) DO NOTHING"

	rowsAffected, err := db.execInsUpdStatement(ctx, query, orderNum, userID)

	if err != nil {
		return err
	} else if rowsAffected == 0 {
		query := "SELECT EXISTS(SELECT 1 FROM orders WHERE number = $1 AND customer = $2)"
		exists := false
		err := db.QueryRow(query, orderNum, userID).Scan(&exists)

		if exists {
			return storageerrs.ErrOrderLoaded
		} else if err != nil {
			return fmt.Errorf("failed to get an order info from Database: %w", err)
		}

		return storageerrs.ErrOrderExists
	}

	return nil
}

type order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (db Database) LoadOrders(ctx context.Context, userID string) ([]byte, error) {
	orderBatch := make([]order, 0)
	query := "SELECT number, uploaded, income, status FROM orders WHERE customer = $1"

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to read orders from Database: %w", err)
	} else if rows.Err() != nil {
		return nil, fmt.Errorf("failed to read orders from Database: %w", err)
	}

	for rows.Next() {
		order := order{}

		err := rows.Scan(&order.Number, &order.UploadedAt, &order.Accrual, &order.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan values from batabase result: %w", err)
		}

		orderBatch = append(orderBatch, order)
	}

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)

	if err = enc.Encode(orderBatch); err != nil {
		return nil, fmt.Errorf("failed to encode orders: %w", err)
	}

	return buf.Bytes(), nil
}

func (db Database) LoadBalance(ctx context.Context, userID string) (float64, float64, error) {
	query := `SELECT balances.balance + Sum(COALESCE(o.income, 0)) - Sum(COALESCE(w.withdraw,0)) balance,
       		balances.total_withdraw + Sum(COALESCE(w.withdraw,0)) withdraw
		FROM balances LEFT JOIN orders o ON o.customer = balances.customer AND o.ts > balances.ts
					  LEFT JOIN withdrawals w ON w.customer = balances.customer AND w.ts > balances.ts
		WHERE balances.customer = $1 GROUP BY balances.balance, balances.total_withdraw`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read balance from Database: %w", err)
	} else if rows.Err() != nil {
		return 0, 0, fmt.Errorf("failed to read balance from Database: %w", err)
	}

	var total, withdrawn float64

	for rows.Next() {
		err := rows.Scan(&total, &withdrawn)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to scan values from batabase result: %w", err)
		}
	}

	return total, withdrawn, nil
}

func (db Database) Withdraw(ctx context.Context, userID, number string, sum float64) error {
	query := "INSERT INTO withdrawals(number, customer, withdraw, ts, processed) VALUES ($1, $2, $3, now()::timestamptz, now())"

	rowsAffected, err := db.execInsUpdStatement(ctx, query, number, userID, sum)

	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return fmt.Errorf("no withdraws were added to Database for some reason")
	}

	return nil
}

func (db Database) LoadWithdrawals(ctx context.Context, userID string) ([]byte, error) {
	withdrawals := make([]orders.Withdrawal, 0)

	query := "SELECT number, withdraw, processed FROM withdrawals WHERE customer = $1 ORDER BY processed desc"

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to read withdrawals from Database: %w", err)
	} else if rows.Err() != nil {
		return nil, fmt.Errorf("failed to read withdrawals from Database: %w", err)
	}

	for rows.Next() {
		order := orders.Withdrawal{}

		err := rows.Scan(&order.Order, &order.Sum, &order.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan values from batabase result: %w", err)
		}

		withdrawals = append(withdrawals, order)
	}

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err = enc.Encode(withdrawals)
	if err != nil {
		return nil, fmt.Errorf("failed to encode withdrawals: %w", err)
	}

	return buf.Bytes(), nil
}

func (db Database) OrdersToProcess(ctx context.Context) (map[string]string, error) {
	ordersMap := make(map[string]string)

	query := "SELECT number, status FROM orders WHERE status NOT IN ('INVALID','PROCESSED')"

	rows, err := db.QueryContext(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("failed to read orders from Database: %w", err)
	} else if rows.Err() != nil {
		return nil, fmt.Errorf("failed to read orders from Database: %w", err)
	}

	for rows.Next() {
		order := order{}
		err := rows.Scan(&order.Number, &order.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan values from batabase result: %w", err)
		}

		ordersMap[order.Number] = order.Status
	}

	return ordersMap, nil
}

func (db Database) UpdateStatuses(ctx context.Context, batch []orders.Status) error {
	valueStrings := make([]string, 0, len(batch))
	valueArgs := make([]interface{}, 0, len(batch)*2)

	c := 1
	for _, status := range batch {
		valueStrings = append(valueStrings, fmt.Sprintf("($%v, $%v::float, $%v)", c, c+1, c+2))
		valueArgs = append(valueArgs, status.Order, status.Accrual, status.Status)
		c += 3
	}

	query := fmt.Sprintf(`UPDATE orders SET status = tmp.status, income = tmp.income, ts = now()::timestamptz 
              FROM (VALUES %s) as tmp (number, income, status) 
			WHERE orders.number = tmp.number`,
		strings.Join(valueStrings, ","))

	_, err := db.execInsUpdStatement(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to update statuses in Database: %w", err)
	}

	return nil
}

func (db Database) UpdateBalances(ctx context.Context) error {
	query := `WITH agr AS(
			SELECT b.balance + sum(COALESCE(tmp.sum,0)) - sum(COALESCE(tmp.withdraw,0)) balance, b.total_withdraw + sum(COALESCE(tmp.withdraw,0)) total_withdraw,
				   Max(COALESCE(tmp.ts, b.ts)) ts, b.customer
		
			FROM balances b LEFT JOIN
				 (SELECT income sum, 0 withdraw, o.ts ts, o.customer customer
				  FROM orders o INNER JOIN balances b on o.customer = b.customer AND o.ts > b.ts
				  UNION
				  SELECT 0, withdraw, w.ts, w.customer FROM withdrawals w INNER JOIN balances b on w.customer = b.customer AND w.ts > b.ts) tmp
				 ON tmp.customer = b.customer
			GROUP BY b.balance, b.total_withdraw, b.customer
		)
		UPDATE balances
		SET balance = agr.balance, total_withdraw = agr.total_withdraw, ts = agr.ts
		FROM agr
		WHERE agr.customer = balances.customer`

	_, err := db.execInsUpdStatement(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to update statuses in Database: %w", err)
	}

	return nil
}
