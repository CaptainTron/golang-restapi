## Recruitment Management System (RMS)
This is a REST server implementation for the Recruitment Management System (RMS) using Golang. It provides a scalable and well-designed API for managing users, profiles, jobs, and applicants.


## Project Structure
```
.
├── README.md
├── RMS.postman_collection.json
├── api.go
├── dockerfile
├── go.mod
├── go.sum
├── main.go
├── models.go
└── postgres.go

0 directories, 10 files
```

## Implementation Details:
- Created API endpoints for user authentication, profile creation, resume upload, job creation, and fetching job and applicant information.
- Implemented authentication using JWT tokens for secure access to the APIs.
- Saved uploaded resumes for future reference.
- Designed the database schema with appropriate models for users, profiles, and jobs.
- Utilizes Postgres as the database for efficient data storage and retrieval.
- Implemented role-based access control to restrict certain APIs to admin or applicant users.
- Handled error cases and implemented appropriate error handling and response messages.
- Follows a standard approach with the use of interfaces, structs, and methods for clean and maintainable code.
- Includes a Dockerfile for easy containerization and deployment.
- Built to be scalable to handle a large number of users and data.
- Implements a well-designed API with clear and consistent endpoints.
- Provides a Postman Collection for easy API reference and testing.

## Technologies:
- Golang: A powerful and efficient programming language for building robust applications.
- Postgres: A reliable and feature-rich relational database management system.
- Docker: A containerization platform for easy deployment and scalability.
- JWT: JSON Web Tokens for secure authentication and authorization.
