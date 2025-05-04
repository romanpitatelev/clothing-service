-- +migrate Up
CREATE TABLE users
(
    id             UUID PRIMARY KEY,
    first_name     VARCHAR,
    last_name      VARCHAR,
    nick_name      VARCHAR                  NOT NULL,
    gender         VARCHAR,
    birth_date     DATE,
    email          VARCHAR,
    email_verified BOOLEAN                  NOT NULL DEFAULT FALSE,
    phone          VARCHAR                  NOT NULL,
    phone_verified BOOLEAN                  NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at     TIMESTAMP WITH TIME ZONE,
    otp            VARCHAR,
    otp_created_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX users_unique_phone_and_deleted_at_null_idx
    ON users (phone)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX users_unique_email_and_deleted_at_null_idx
    ON users (email)
    WHERE deleted_at IS NULL;

CREATE TABLE clothes
(
    id                      UUID PRIMARY KEY,
    brand                   VARCHAR                  NOT NULL,
    category                VARCHAR                  NOT NULL,
    color                   VARCHAR                  NOT NULL,
    name                    VARCHAR                  NOT NULL,
    price                   INT                      NOT NULL,
    sizes                   VARCHAR                  NOT NULL,
    product_url             VARCHAR                  NOT NULL,
    primary_photo           VARCHAR                  NOT NULL,
    extra_photo_1           VARCHAR,
    extra_photo_2           VARCHAR,
    extra_photo_3           VARCHAR,
    extra_photo_4           VARCHAR,
    extra_photo_5           VARCHAR,
    related_product_url_1   VARCHAR,
    related_product_url_2   VARCHAR,
    related_product_url_3   VARCHAR,
    related_product_url_4   VARCHAR,
    related_product_url_5   VARCHAR,
    related_product_photo_1 VARCHAR,
    related_product_photo_2 VARCHAR,
    related_product_photo_3 VARCHAR,
    related_product_photo_4 VARCHAR,
    related_product_photo_5 VARCHAR,
    created_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at              TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX clothes_unique_url_and_deleted_at_null_idx
    ON clothes (product_url)
    WHERE deleted_at IS NULL;

-- +migrate Down
DROP INDEX users_unique_phone_and_deleted_at_null_idx;
DROP INDEX users_unique_email_and_deleted_at_null_idx;
DROP INDEX clothes_unique_url_and_deleted_at_null_idx;
DROP TABLE users;
DROP TABLE clothes;
