-- +migrate Up
CREATE TABLE users (
	id UUID NOT NULL UNIQUE PRIMARY KEY,
	first_name VARCHAR NOT NULL,
	last_name VARCHAR NOT NULL,
	nick_name VARCHAR NOT NULL,
	gender VARCHAR NOT NULL,
	age NUMERIC NOT NULL CHECK (age > 12),
	email VARCHAR NOT NULL,
	country VARCHAR NOT NULL,
	lower_body_clothing_size NUMERIC NOT NULL CHECK (lower_body_clothing_size > 0),
	upper_body_clothing_size NUMERIC NOT NULL CHECK (upper_body_clothing_size > 0),
	footwear_size NUMERIC NOT NULL CHECK (footwear_size > 0),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMPL WITH TIME ZONE,
	active BOOLLEAN NOT NULL DEFAULT TRUE
);

-- +migrate Down
DROP TABLE users CASCADE;

