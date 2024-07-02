package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(url string) (*PostgresStore, error) {
	// connStr := "user=postgres dbname=postgres password=user sslmode=disable"
	db, err := sql.Open("postgres", url)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}

func (s *PostgresStore) createAccountTable() error {
	queryAccount := `CREATE TABLE IF NOT EXISTS account (
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		number serial,
		balance serial CHECK (balance >= 0),
		created_at timestamp
	  )`
	_, err := s.db.Exec(queryAccount)
	if err != nil {
		return err
	}

	queryUser := `CREATE TABLE IF NOT EXISTS users (
		first_name varchar(200),
		last_name varchar(200),
		number serial PRIMARY KEY,
		password varchar(200),
		created_at timestamp
	  )`
	_, err = s.db.Exec(queryUser)

	return err
}

func (s *PostgresStore) LoginUser(user *Login) error {
	query := `insert into users (first_name, last_name, number, password, created_at)
	values($1, $2, $3, $4, $5)`

	_, err := s.db.Query(query, user.FirstName, user.LastName, user.Number, user.Password, user.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

// this will help to create account
func (s *PostgresStore) CreateAccount(acc *Account) error {
	query := `insert into account (first_name, last_name, number, balance, created_at)
	values($1, $2, $3, $4, $5)`

	_, err := s.db.Query(query, acc.FirstName, acc.LastName, acc.Number, acc.Balance, acc.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) UpdateAccount(account *Account) error {
	stmt, err := s.db.Prepare("UPDATE account SET first_name = $1, last_name = $2, number = $3, balance = $4 WHERE id = $5")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(account.FirstName, account.LastName, account.Number, account.Balance, account.ID)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	stmt, err := s.db.Prepare("DELETE FROM account WHERE id = $1")
	if err != nil {
		return err
	}
	result, _ := stmt.Exec(id)
	defer stmt.Close()
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no account found with id: %d", id)
	}
	return nil
}

func (s *PostgresStore) GetAccountByNumber(number int) (*Account, error) {
	rows, err := s.db.Query("select * from account where number = $1", number)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account with number: %d does not exist", number)
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	rows, err := s.db.Query("select * from account where id = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account with id: %d does not exist", id)

}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("select * from account")
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		// This will get the account rows data
		account, err := scanIntoAccount(rows)
		if err != nil {
			log.Fatal(err)
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// Scan to Rows
func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.Balance,
		&account.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *PostgresStore) TransferAmount(fromID, toID, amount int) error {

	// 1. Open a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// 2. Defer closing the transaction (commit or rollback)
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// 3. Prepare statements with placeholders
	// 3.a. Update `from` account (subtract amount)
	fromStmt, err := tx.Prepare("UPDATE account SET balance = balance - $1 WHERE id = $2")
	if err != nil {
		return err
	}
	defer fromStmt.Close()

	// 3.b. Update `to` account (add amount)
	toStmt, err := tx.Prepare("UPDATE account SET balance = balance + $1 WHERE id = $2")
	if err != nil {
		return err
	}
	defer toStmt.Close()

	// 4. Execute updates within the transaction
	_, err = fromStmt.Exec(amount, fromID)
	if err != nil {
		return fmt.Errorf("insufficient balance for account %d", fromID)
	}

	_, err = toStmt.Exec(amount, toID)
	if err != nil {
		return err
	}

	// 5. If all successful, commit the transaction
	return nil
}
