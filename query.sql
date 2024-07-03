-- Active: 1719642225490@@127.0.0.1@5432@postgres
CREATE TABLE IF NOT EXISTS users (
		id serial primary key,
		name varchar(50),
		email varchar(50),
		address varchar(50),
		user_type varchar(50),
		password_hash varchar(0),
		profile_headline varchar(100),
		created_at timestamp
	  )