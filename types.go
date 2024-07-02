package main

import (
	"math/rand"
	// "fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Login struct {
	Number    int64     `json:"number"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAccountRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Number    int64  `json:"number"`
}

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Number    int64     `json:"number"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

func LoginAccount(FirstName, LastName, Password string, numb int64) (*Login, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &Login{
		FirstName: FirstName,
		LastName:  LastName,
		Password:  string(encpw),
		Number:    int64(rand.Intn(100000)),
		CreatedAt: time.Now().UTC(),
	}, nil
}

func NewAccount(FirstName, LastName string, Number int64) (*Account, error) {
	return &Account{
		FirstName: FirstName,
		LastName:  LastName,
		Number:    Number,
		CreatedAt: time.Now().UTC(),
	}, nil
}

type Transfer_amount struct {
	FromID int `json:"fromId"`
	ToID   int `json:"toId"`
	Amount int `json:"amount"`
}

type Message struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}
