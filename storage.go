package main

import (
	"database/sql"
	"fmt"

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
	query := `CREATE TABLE IF NOT EXISTS account (
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		number serial,
		balance serial CHECK (balance >= 0),
		created_at timestamp
	  )`
	_, err := s.db.Exec(query)
	return err
}

// this will help to create account
func (s *PostgresStore) CreateAccount(acc *Account) error {
	query := `insert into account (first_name, last_name, number, balance, created_at)
	values($1, $2, $3, $4, $5)`

	resp, err := s.db.Query(query, acc.FirstName, acc.LastName, acc.Number, acc.Balance, acc.CreatedAt)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", resp)
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

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	// 1. Prepare the statement with a placeholder
	stmt, err := s.db.Prepare("SELECT * FROM account WHERE id = $1")
	if err != nil {
		return nil, err
	}
	row := stmt.QueryRow(id)
	account := &Account{}
	err = row.Scan(&account.ID, &account.FirstName, &account.LastName, &account.Number, &account.Balance, &account.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer stmt.Close()
	return account, nil
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("select * from account")
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
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
		accounts = append(accounts, account)
	}
	return accounts, nil
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
