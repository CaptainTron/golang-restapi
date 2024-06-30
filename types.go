package main

import (
	// "go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"time"
)

type CreateAccountRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Number    int64     `json:"number"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAccount(FirstName, LastName string) *Account {
	return &Account{
		// ID:        rand.Intn(10000),
		FirstName: FirstName,
		LastName:  LastName,
		Number:    int64(rand.Intn(100000)),
		CreatedAt: time.Now().UTC(),

		Balance: 3000,
	}
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
