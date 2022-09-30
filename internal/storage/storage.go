package storage

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/usa4ev/gophermart/internal/auth"
	"github.com/usa4ev/gophermart/internal/orders"
	"github.com/usa4ev/gophermart/internal/storage/storageerrs"
	"strings"
	"time"

	_ "github.com/jackc/pgx/stdlib"
)

type (
	Database struct {
		*sql.DB
		//stmts  statements
		//buffer *asyncBuf
	}
	//statements struct {
	//	// USERS
	//	userExists *sql.Stmt
	//	storeUser  *sql.Stmt
	//	loadHash   *sql.Stmt
	//
	//	//ORDERS
	//	storeOrder      *sql.Stmt
	//	updateStatus    *sql.Stmt
	//	loadOrders      *sql.Stmt
	//	orderExists     *sql.Stmt
	//	ordersToProcess *sql.Stmt
	//
	//	// BALANCE
	//	getBalance      *sql.Stmt
	//	updateBalance   *sql.Stmt
	//	withdraw        *sql.Stmt
	//	loadWithdrawals *sql.Stmt
	//}
	//Item struct {
	//	ID     string
	//	UserID string
	//}
	//asyncBuf struct {
	//	*bufio.Writer
	//	enc *gob.Encoder
	//	mx  sync.Mutex
	//	ew  *bytes.Buffer
	//	t   *time.Ticker
	//}
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
                pwdhash VARCHAR(100) not null);`

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
					balance int not null,
					total_withdraw float not null,
					FOREIGN KEY (customer)
				REFERENCES users (id));`

	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	return err
}

//func (db Database) prepareStatements() (statements, error) {
//	userExists, err := db.PrepareContext(db.ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)")
//	if err != nil {
//		return statements{}, err
//	}
//
//	storeUser, err := db.PrepareContext(db.ctx, "INSERT INTO users(id, username, pwdhash) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING;"+
//		"INSERT INTO balances(customer, ts, balance, total_withdraw) VALUES ($1, now()::timestamptz, 0, 0) ON CONFLICT (customer) DO NOTHING;")
//	if err != nil {
//		return statements{}, err
//	}
//
//	orderExists, err := db.PrepareContext(db.ctx, "SELECT EXISTS(SELECT 1 FROM orders WHERE number = $1 && customer = $2)")
//	if err != nil {
//		return statements{}, err
//	}
//
//	loadHash, err := db.PrepareContext(db.ctx, "SELECT id, pwdhash FROM users WHERE username = $1")
//	if err != nil {
//		return statements{}, err
//	}
//
//	storeOrder, err := db.PrepareContext(db.ctx, "INSERT INTO orders(number, customer, ts, uploaded, status, income) VALUES ($1, $2, now()::timestamptz, now(), 'NEW', 0)")
//	if err != nil {
//		return statements{}, err
//	}
//
//	updateStatus, err := db.PrepareContext(db.ctx, "UPDATE orders SET number = $1, income = $2, status = $3, ts) WHERE number = $4")
//	if err != nil {
//		return statements{}, err
//	}
//
//	loadOrders, err := db.PrepareContext(db.ctx, "SELECT number, uploaded, income, status FROM orders WHERE customer = $1")
//	if err != nil {
//		return statements{}, err
//	}
//
//	getBalance, err := db.PrepareContext(db.ctx, "SELECT balances.balance + Sum(o.income) - Sum(w.withdraw) balance, "+
//		"balances.total_withdraw + Sum(w.withdraw) withdraw"+
//		"FROM balances LEFT JOIN orders o ON o.customer = balances.customer && o.ts > balances.ts "+
//		"	LEFT JOIN withdraws w ON w.customer = balances.customer && w.ts > balances.ts "+
//		"WHERE balances.customer = $1 GROUP BY balances.balance")
//	if err != nil {
//		return statements{}, err
//	}
//
//	updateBalance, err := db.PrepareContext(db.ctx, "UPDATE balances "+
//		"SET balance = balance + sum(ISNULL(tmp.sum,0)), total_withdraw = total_withdraw + sum(ISNULL(ts.withdraw,0)), ts = Max(ISNULL(tmp.ts, ts)) "+
//		"FROM (SELECT income sum, 0 withdraw, o.ts ts"+
//		"		FROM orders o INNER JOIN balances b on o.customer = b.customer && o.ts > b.ts"+
//		"UNION"+
//		"SELECT 0, withdraw, w.ts FROM withdraws w INNER JOIN balances w on w.customer = b.customer && w.ts > b.ts) tmp "+
//		"WHERE tmp.customer = balances.customer")
//	if err != nil {
//		return statements{}, err
//	}
//
//	withdraw, err := db.PrepareContext(db.ctx, "INSERT INTO withdrawals(number, customer, withdraw, ts, processed) VALUES ($1, $2, 3, now()::timestamptz, now())")
//	if err != nil {
//		return statements{}, err
//	}
//
//	loadWithdrawals, err := db.PrepareContext(db.ctx, "SELECT number, processed, withdraw, status FROM withdrawals WHERE customer = $1")
//	if err != nil {
//		return statements{}, err
//	}
//
//	ordersToProcess, err := db.PrepareContext(db.ctx, "SELECT number, status FROM orders WHERE status NOT IN ('INVALID','PROCESSED')")
//	if err != nil {
//		return statements{}, err
//	}
//
//	return statements{
//		userExists:      userExists,
//		storeUser:       storeUser,
//		loadHash:        loadHash,
//		storeOrder:      storeOrder,
//		updateStatus:    updateStatus,
//		loadOrders:      loadOrders,
//		orderExists:     orderExists,
//		ordersToProcess: ordersToProcess,
//		getBalance:      getBalance,
//		updateBalance:   updateBalance,
//		withdraw:        withdraw,
//		loadWithdrawals: loadWithdrawals,
//	}, nil
//}

// AddUser adds new row to Database and return new user ID or error if addition failed
func (db Database) AddUser(ctx context.Context, username, hash string) (string, error) {
	id := uuid.New().String()

	query := "INSERT INTO users(id, username, pwdhash) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING;" +
		"INSERT INTO balances(customer, ts, balance, total_withdraw) VALUES ($1, now()::timestamptz, 0, 0) ON CONFLICT (customer) DO NOTHING;"

	rowsAffected, err := db.execInsUpdStatement(ctx, query, id, username, hash)

	if err != nil {
		return "", err
	} else if rowsAffected == 0 {
		return "", auth.ErrUserAlreadyExists
	}

	return id, nil
}

// UserExists returns true if user found by given userName or false otherwise
func (db Database) UserExists(ctx context.Context, userName string) (bool, error) {
	var userID, hash string

	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"

	err := db.QueryRowContext(ctx, query, userName).Scan(&userID, &hash)

	if errors.Is(err, sql.ErrNoRows) {

		return false, nil
	} else if err != nil {

		return false, fmt.Errorf("failed to get a password hash from Database: %w", err)
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

func (db Database) execInsUpdStatement(ctx context.Context, query string, args ...any) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, query, args)

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
	query := "INSERT INTO orders(number, customer, ts, uploaded, status, income) VALUES ($1, $2, now()::timestamptz, now(), 'NEW', 0)"

	rowsAffected, err := db.execInsUpdStatement(ctx, query, orderNum, userID)

	if err != nil {
		return err
	} else if rowsAffected == 0 {
		err := db.QueryRow(query, orderNum, userID).Scan()

		if errors.Is(err, sql.ErrNoRows) {
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
	Accrual    int       `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (db Database) LoadOrders(ctx context.Context, userID string) ([]byte, error) {
	orders := make([]order, 0)
	query := "SELECT number, uploaded, income, status FROM orders WHERE customer = $1"

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to read orders from Database: %w", err)
	}

	for rows.Next() {
		order := order{}
		err := rows.Scan(&order.Number, &order.UploadedAt, &order.Accrual, &order.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan values from batabase result: %w", err)
		}

		orders = append(orders, order)
	}

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err = enc.Encode(orders)
	if err != nil {
		return nil, fmt.Errorf("failed to encode orders: %w", err)
	}

	return buf.Bytes(), nil
}

func (db Database) LoadBalance(ctx context.Context, userID string) (float64, float64, error) {
	query := "SELECT balances.balance + Sum(o.income) - Sum(w.withdraw) balance, " +
		"balances.total_withdraw + Sum(w.withdraw) withdraw" +
		"FROM balances LEFT JOIN orders o ON o.customer = balances.customer && o.ts > balances.ts " +
		"	LEFT JOIN withdraws w ON w.customer = balances.customer && w.ts > balances.ts " +
		"WHERE balances.customer = $1 GROUP BY balances.balance"

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read orders from Database: %w", err)
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
	query := "INSERT INTO withdrawals(number, customer, withdraw, ts, processed) VALUES ($1, $2, 3, now()::timestamptz, now())"

	rowsAffected, err := db.execInsUpdStatement(ctx, query, userID, number, sum)

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
		return nil, fmt.Errorf("failed to read orders from Database: %w", err)
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
	orders := make(map[string]string)

	query := "SELECT number, status FROM orders WHERE status NOT IN ('INVALID','PROCESSED')"

	rows, err := db.QueryContext(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("failed to read orders from Database: %w", err)
	}

	for rows.Next() {
		order := order{}
		err := rows.Scan(&order.Number, &order.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan values from batabase result: %w", err)
		}

		orders[order.Number] = order.Status
	}

	return orders, nil
}

func (db Database) UpdateStatuses(ctx context.Context, batch []orders.Status) error {
	valueStrings := make([]string, 0, len(batch))
	valueArgs := make([]interface{}, 0, len(batch)*2)

	c := 1
	for _, status := range batch {
		valueStrings = append(valueStrings, fmt.Sprintf("($%v, $%v, $%v, TRUE)", c, c+1))
		valueArgs = append(valueArgs, status.Order, status.Accrual, status.Status)
		c += 2
	}

	query := fmt.Sprintf("UPDATE orders SET status = tmp.status, income = tmp.income from (VALUES %s) as tmp (number, income, status) "+
		"WHERE orders.number = tmp.nember",
		strings.Join(valueStrings, ","))

	_, err := db.execInsUpdStatement(ctx, query, valueArgs)
	if err != nil {
		return fmt.Errorf("failed to update statuses in Database: %w", err)
	}

	return nil
}

func (db Database) UpdateBalances(ctx context.Context) error {
	query := "UPDATE balances " +
		"SET balance = balance + sum(ISNULL(tmp.sum,0)), total_withdraw = total_withdraw + sum(ISNULL(ts.withdraw,0)), ts = Max(ISNULL(tmp.ts, ts)) " +
		"FROM (SELECT income sum, 0 withdraw, o.ts ts" +
		"		FROM orders o INNER JOIN balances b on o.customer = b.customer && o.ts > b.ts" +
		"UNION" +
		"SELECT 0, withdraw, w.ts FROM withdrawals w INNER JOIN balances b on w.customer = b.customer && w.ts > b.ts) tmp " +
		"WHERE tmp.customer = balances.customer"

	_, err := db.execInsUpdStatement(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to update statuses in Database: %w", err)
	}

	return nil
}
