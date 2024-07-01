package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	// "time"

	"github.com/gorilla/mux"

	"github.com/golang-jwt/jwt/v5"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
	TransferAmount(int, int, int) error
}

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listen string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listen,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/account", (makeHTTPHandlerFunc(s.handleAccount)))
	router.HandleFunc("/account_transfer", makeHTTPHandlerFunc(s.handleTransfer))
	router.HandleFunc("/account_id/{id}", withJWTAuth(makeHTTPHandlerFunc(s.handleGetAccountByID), s.store))
	router.HandleFunc("/account_delete/{id}", makeHTTPHandlerFunc(s.handleDeleteAccount))

	log.Println("Server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccounts(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

// Fetch all the accounts
func (s *APIServer) handleGetAccounts(w http.ResponseWriter, _ *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, accounts)
}

// Fetch account by ID
func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}

	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, account)
}

// Handle to create new account
func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	CreateAccountReq := CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(&CreateAccountReq); err != nil {
		return err
	}

	account := NewAccount(CreateAccountReq.FirstName, CreateAccountReq.LastName)
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	fmt.Println("Creating JWT")
	tokenString, err := createJWT(account)
	if err != nil {
		return err
	}
	fmt.Println(tokenString)
	return writeJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	ID, _ := strconv.Atoi(id)
	err := s.store.DeleteAccount(ID)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, Message{Message: "Successfully deleted!", Status: "Successfully Deleted!"})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transfer := &Transfer_amount{}
	if err := json.NewDecoder(r.Body).Decode(&transfer); err != nil {
		return err
	}

	if err := s.store.TransferAmount(transfer.FromID, transfer.ToID, transfer.Amount); err != nil {
		return err
	}
	message := &Message{"Transaction Completed", "Successful"}
	return writeJSON(w, http.StatusOK, message)
}

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50TnVtYmVyIjo3NzM0NCwiZXhwaXJlc0F0IjoxNTAwfQ.JignQQQkQvVYP9SyW1SIaKif8lAN0xwPtT6Dfd0Ao9o

func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":     1500,
		"accountNumber": account.Number,
	}
	signKey := "PASSWORD"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signKey))
}

// JSON Token
func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Calling JWT handler")

		tokenString := r.Header.Get("auth")
		token, err := validateJWT(tokenString)
		if err != nil {
			writeJSON(w, http.StatusNotAcceptable, apiError{Error: "invalid token"})
			return
		}

		if !token.Valid {
			writeJSON(w, http.StatusForbidden, apiError{Error: "invalid token"})
			return
		}

		userId, err := getId(r)
		if err != nil {
			writeJSON(w, http.StatusForbidden, apiError{Error: "invalid token"})
			return
		}

		account, err := s.GetAccountByID(userId)
		if err != nil {
			writeJSON(w, http.StatusForbidden, apiError{Error: "invalid token"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		fmt.Println(claims, account.Number)
		if account.Number != claims["accountNumber"] {
			writeJSON(w, http.StatusForbidden, apiError{Error: "invalid token"})
			return
		}
		handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := "PASSWORD"
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

// Helper one
func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type apiError struct {
	Error string `json:"error"`
}

func makeHTTPHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		}
	}
}

func getId(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}
	return id, nil
}
