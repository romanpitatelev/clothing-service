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

CREATE TABLE brands
(
    id         UUID PRIMARY KEY,
    name       VARCHAR                  NOT NULL,
    logo       VARCHAR                  NOT NULL DEFAULT '',
    url        VARCHAR                  NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX brands_unique_name_and_deleted_at_null_idx
    ON brands (name)
    WHERE deleted_at IS NULL;

CREATE TABLE products
(
    id         UUID PRIMARY KEY,
    brand_id   UUID REFERENCES brands (id) NOT NULL,
    category   VARCHAR                     NOT NULL,
    name       VARCHAR                     NOT NULL,
    price      BIGINT                      NOT NULL,
    currency   VARCHAR                     NOT NULL,
    colors     JSONB,
    created_at TIMESTAMP WITH TIME ZONE    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX products_unique_brand_category_name_and_deleted_at_null_idx
    ON products (brand_id, category, name)
    WHERE deleted_at IS NULL;

CREATE TABLE variants
(
    id                      UUID PRIMARY KEY,
    product_id              UUID REFERENCES products (id) NOT NULL,
    color                   VARCHAR                       NOT NULL,
    color_hex               VARCHAR                       NOT NULL,
    sizes                   JSONB                         NOT NULL,
    sold_out_sizes          JSONB                         NOT NULL,
    product_url             VARCHAR                       NOT NULL,
    primary_photo           VARCHAR                       NOT NULL,
    photo_1                 VARCHAR,
    photo_2                 VARCHAR,
    photo_3                 VARCHAR,
    photo_4                 VARCHAR,
    photo_5                 VARCHAR,
    photo_6                 VARCHAR,
    photo_7                 VARCHAR,
    photo_8                 VARCHAR,
    photo_9                 VARCHAR,
    photo_10                VARCHAR,
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
    created_at              TIMESTAMP WITH TIME ZONE      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP WITH TIME ZONE      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at              TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX variants_unique_url_and_deleted_at_null_idx
    ON variants (product_url)
    WHERE deleted_at IS NULL;

-- +migrate Down
DROP INDEX users_unique_phone_and_deleted_at_null_idx;
DROP INDEX users_unique_email_and_deleted_at_null_idx;
DROP INDEX brands_unique_name_and_deleted_at_null_idx;
DROP INDEX products_unique_brand_category_name_and_deleted_at_null_idx;
DROP INDEX variants_unique_url_and_deleted_at_null_idx;
DROP TABLE users;
DROP TABLE brands;
DROP TABLE products;
DROP TABLE variants;
