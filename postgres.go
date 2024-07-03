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
	queries := []string{
		`CREATE TABLE IF NOT EXISTS user_table (
		id serial primary key,
		name varchar(100),
		email varchar(100),
		address varchar(100),
		user_type varchar(200),
		password_hash varchar(200),
		profile_headline varchar(200),
		created_at timestamp
		)`,

		`CREATE TABLE IF NOT EXISTS Profile (
		user_id BIGINT UNIQUE REFERENCES user_table(id),
		id serial primary key,
		name varchar(100),
		resume_file_address varchar(100),
		skills varchar(100),
		education varchar(100),
		experience varchar(100),
		email varchar(100),
		phone varchar(100),
		created_at timestamp
		)`,

		`CREATE TABLE IF NOT EXISTS Jobs (
		user_id BIGINT UNIQUE REFERENCES user_table(id),
		id serial primary key,
		title varchar(100),
		description varchar(100),
		total_applications integer,
		company_name varchar(100),
		posted_on timestamp
		)`,
	}

	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStore) SignUp(user *User) error {
	query := `insert into user_table (name, email, user_type, address, password_hash, profile_headline, created_at)
	values($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.db.Query(query, user.Name, user.Email, user.UserType, user.Address, user.PasswordHash, user.ProfileHeadline, user.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

// Login with "email" and "password" to generate JWT token
func (s *PostgresStore) LoginUser(acc *Login) (*User, error) {
	var user User
	err := s.db.QueryRow("select * from user_table where email = $1", acc.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Address, &user.UserType,
		&user.PasswordHash, &user.ProfileHeadline, &user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("no account found with given email: %s", acc.Email)
	}
	return &user, nil
}

// create new job
func (s *PostgresStore) PostJob(job *Job) error {
	query := `insert into Jobs (title, description, total_applications, company_name, posted_on)
	values($1, $2, $3, $4, $5)`

	_, err := s.db.Query(query, job.Title, job.Description, job.TotalApplications, job.CompanyName, job.PostedOn)
	if err != nil {
		return err
	}
	return nil
}

// Fetch Details about jobs and list of applicants
func (s *PostgresStore) FetchJobApplicants(id int) (*Job, error) {
	jobs := &Job{}
	query := `select * from Jobs where id = $1`
	err := s.db.QueryRow(query, id).Scan(&jobs.ID, &jobs.Title, &jobs.Description, &jobs.TotalApplications, &jobs.CompanyName, &jobs.PostedOn)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

// List all the users
func (s *PostgresStore) ListUsers() ([]*User, error) {
	rows, err := s.db.Query("select * from user_table")
	if err != nil {
		return nil, err
	}

	users := []*User{}
	for rows.Next() {
		// Scan each row of database and store in User
		user, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
func scanIntoAccount(rows *sql.Rows) (*User, error) {
	user := &User{}
	err := rows.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Address,
		&user.UserType,
		&user.PasswordHash,
		&user.ProfileHeadline,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Get Single Applicant by Id
func (s *PostgresStore) GetApplicant(applicant_id int) (*Profile, error) {
	profile := &Profile{}
	query := "select * from Profile where id = $1"
	err := s.db.QueryRow(query, applicant_id).Scan(&profile.Id, &profile.Name, &profile.ResumeFileAddress, &profile.Skills, &profile.Education, &profile.Experience, &profile.Email, &profile.Phone)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

// List All the jobs
func (s *PostgresStore) ListJobs() ([]*Job, error) {
	rows, err := s.db.Query("select * from Jobs")
	if err != nil {
		return nil, err
	}

	Jobs := []*Job{}
	for rows.Next() {
		job := &Job{}
		err := rows.Scan(
			&job.ID,
			&job.Title,
			&job.Description,
			&job.TotalApplications,
			&job.CompanyName,
			&job.PostedOn,
		)
		if err != nil {
			return nil, err
		}
		Jobs = append(Jobs, job)
	}
	return Jobs, nil
}

func (s *PostgresStore) GetJob_byId(job_id int) (*Job, error) {
	job := &Job{}
	query := "select * from Jobs where id = $1"
	err := s.db.QueryRow(query, job_id).Scan(&job.ID, &job.Title, &job.Description, &job.TotalApplications, &job.CompanyName, &job.PostedOn)
	if err != nil {
		return nil, err
	}
	return job, nil
}
