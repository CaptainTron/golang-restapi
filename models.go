package main

import (
	//"fmt"
	// "database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Profile struct {
	Applicant         *User  `json:"user"`
	Id                string `json:"id"`
	ResumeFileAddress string `json:"resume_file_address"`
	Skills            string `json:"skills"`
	Education         int64  `json:"education"`
	Experience        string `json:"experience"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
}

type User struct {
	ID              int       `db:"id" json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Address         string    `json:"address"`
	UserType        string    `json:"user_type"`
	PasswordHash    string    `json:"-"`
	ProfileHeadline string    `json:"profile_headline"`
	CreatedAt       time.Time `json:"created_at"`
	Profile         *Profile  `json:"profile"`
}

func SignUpAccount(name, email, address, UserType, profileheadline, password_hash string) (*User, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(password_hash), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		Name:            name,
		Email:           email,
		Address:         address,
		PasswordHash:    string(encpw),
		UserType:        UserType,
		ProfileHeadline: profileheadline,
		CreatedAt:       time.Now().UTC(),
	}, nil
}

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Job struct {
	ID                int       `db:"id" json:"id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	PostedOn          time.Time `json:"posted_on"`
	TotalApplications int       `json:"total_application"`
	CompanyName       string    `json:"company_name"`
	// PostedBy          *User
}
