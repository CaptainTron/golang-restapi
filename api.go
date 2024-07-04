package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type Storage interface {
	SignUp(*User) error
	LoginUser(*Login) (*User, error)
	PostJob(*Job) error
	FetchJobApplicants(int) (*Job, error)
	ListUsers() ([]*User, error)
	GetApplicant(int) (*Profile, error)
	ListJobs() ([]*Job, error)
	GetJob_byId(int) (*Job, error)
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

	router.HandleFunc("/signup", (makeHTTPHandlerFunc(s.SignUp)))
	router.HandleFunc("/login", (makeHTTPHandlerFunc(s.LoginUser)))
	router.HandleFunc("/uploadresume", withJWTAuth(makeHTTPHandlerFunc(s.UploadResume), s.store))
	router.HandleFunc("/admin_createjob", (makeHTTPHandlerFunc(s.Admin_Createjob)))
	router.HandleFunc("/admin_jobs/{id}", (makeHTTPHandlerFunc(s.Admin_jobsInfo)))
	router.HandleFunc("/admin_applicants", (makeHTTPHandlerFunc(s.ListApplicants)))
	router.HandleFunc("/admin_applicant/{id}", (makeHTTPHandlerFunc(s.GetApplicant)))
	router.HandleFunc("/jobs", (makeHTTPHandlerFunc(s.ListJobs)))
	router.HandleFunc("/apply_jobs/{id}", (makeHTTPHandlerFunc(s.GetJob_byId)))

	log.Println("Server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) SignUp(w http.ResponseWriter, r *http.Request) error {
	req := User{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	user, err := SignUpAccount(req.Name, req.Email, req.Address, req.UserType, req.ProfileHeadline, req.PasswordHash)
	if err != nil {
		return err
	}
	if err := s.store.SignUp(user); err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, user)
}

func (s *APIServer) LoginUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method is not allowed %s", r.Method)
	}

	login := &Login{}
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		return err
	}

	if _, err := s.store.LoginUser(login); err != nil {
		return err
	}

	// Create New Token everytime user login
	tokenString, err := createJWT(login)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, map[string]string{"token": tokenString})
}

func (s *APIServer) UploadResume(w http.ResponseWriter, r *http.Request) error {
	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return err
	}

	// Retrieve the file from the form
	file, handler, err := r.FormFile("resume")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return err
	}
	defer file.Close()

	// Define the path where the file will be saved
	uploadPath := "./uploads"
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		http.Error(w, "Unable to create upload directory", http.StatusInternalServerError)
		return err
	}

	// Create the file on the server
	dst, err := os.Create(filepath.Join(uploadPath, handler.Filename))
	if err != nil {
		http.Error(w, "Unable to create the file", http.StatusInternalServerError)
		return err
	}
	defer dst.Close()

	// Copy the uploaded file data to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Unable to save the file", http.StatusInternalServerError)
		return err
	}

	// Respond to the client
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully: %s", handler.Filename)
	return nil
}

func (s *APIServer) Admin_Createjob(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed")
	}

	job := &Job{}
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		return err
	}

	newJob, err := CreateNewJob(job.Title, job.Description, job.TotalApplications, job.CompanyName, job.PostedBy)
	if err != nil {
		return err
	}

	err = s.store.PostJob(newJob)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "Job Created"})

	return nil
}

// This will fetch information about JobInfo and list of applicants
func (s *APIServer) Admin_jobsInfo(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}

	jobs, err := s.store.FetchJobApplicants(id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, jobs)

	return nil
}

// Fetch list of all the users in the system
func (s *APIServer) ListApplicants(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return fmt.Errorf("method not allowed")
	}

	jobs, err := s.store.ListUsers()
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusCreated, jobs)

	return nil
}

// Get Single Applicant by Id
func (s *APIServer) GetApplicant(w http.ResponseWriter, r *http.Request) error {
	applicant_id, err := getId(r)
	if err != nil {
		return err
	}

	applicant, err := s.store.GetApplicant(applicant_id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusCreated, applicant)
	return nil
}

// Get All the Jobs
func (s *APIServer) ListJobs(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return fmt.Errorf("method not allowed")
	}

	jobs, err := s.store.ListJobs()
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusCreated, jobs)
	return nil
}

func (s *APIServer) GetJob_byId(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed")
	}

	applicant_id, err := getId(r)
	if err != nil {
		return err
	}

	applicant, err := s.store.GetJob_byId(applicant_id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusCreated, applicant)
	return nil
}

		///////////////////////////////////////////
////////                 JSON Token               /////////////////////////

func createJWT(account *Login) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":    1500,
		"accountEmail": account.Email,
	}
	signKey := "PASSWORD"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signKey))
}

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

		// userId, err := getId(r)
		// if err != nil {
		// 	writeJSON(w, http.StatusForbidden, apiError{Error: "invalid token"})
		// 	return
		// }

		// account, err := s.GetAccountByID(userId)
		// if err != nil {
		// 	writeJSON(w, http.StatusForbidden, apiError{Error: "invalid token"})
		// 	return
		// }

		// claims := token.Claims.(jwt.MapClaims)
		// fmt.Println(claims, account.Number)
		// if account.Number != claims["accountEmail"] {
		// 	writeJSON(w, http.StatusForbidden, apiError{Error: "invalid token"})
		// 	return
		// }
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

// Helper One
// //////////////////////////////////////////
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
