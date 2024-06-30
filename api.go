package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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

	router.HandleFunc("/account", makeHTTPHandlerFunc(s.handleAccount))
	router.HandleFunc("/account_transfer", makeHTTPHandlerFunc(s.handleTransfer))
	router.HandleFunc("/account_id/{id}", makeHTTPHandlerFunc(s.handleGetAccountByID))
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
	return fmt.Errorf("Method not allowed %s", r.Method)
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
	id := mux.Vars(r)["id"]
	ID, _ := strconv.Atoi(id)
	account, err := s.store.GetAccountByID(ID)
	if err != nil {
		return fmt.Errorf("no account found with Id: %d", ID)
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

// Helper one
func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type apiError struct {
	Error string
}

func makeHTTPHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		}
	}
}
