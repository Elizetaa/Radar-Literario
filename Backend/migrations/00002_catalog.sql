-- +goose Up
CREATE TABLE authors (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 name VARCHAR(160) NOT NULL, biography TEXT NOT NULL,
 birth_date DATE NULL, death_date DATE NULL, country_of_origin VARCHAR(100) NOT NULL DEFAULT '', sex VARCHAR(80) NULL,
 created_at DATETIME(6) NOT NULL DEFAULT UTC_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT UTC_TIMESTAMP(6),
 CONSTRAINT ck_author_dates CHECK (death_date IS NULL OR birth_date IS NULL OR death_date >= birth_date),
 INDEX ix_authors_name(name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE publishers (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 name VARCHAR(160) NOT NULL, website_url VARCHAR(2048) NOT NULL DEFAULT '', country_of_origin VARCHAR(100) NOT NULL DEFAULT '',
 created_at DATETIME(6) NOT NULL DEFAULT UTC_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT UTC_TIMESTAMP(6), INDEX ix_publishers_name(name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE genres (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 name VARCHAR(100) NOT NULL UNIQUE, active BOOLEAN NOT NULL DEFAULT TRUE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE works (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 title VARCHAR(300) NOT NULL, synopsis TEXT NOT NULL, original_language VARCHAR(35) NOT NULL,
 first_publication_date DATE NULL,
 status VARCHAR(20) NOT NULL DEFAULT 'draft', minimum_age TINYINT UNSIGNED NOT NULL DEFAULT 0,
 version BIGINT UNSIGNED NOT NULL DEFAULT 1,
 created_at DATETIME(6) NOT NULL DEFAULT UTC_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT UTC_TIMESTAMP(6),
 CONSTRAINT ck_work_status CHECK (status IN ('draft','pending_review','published','rejected','hidden','archived')),
 CONSTRAINT ck_work_age CHECK (minimum_age IN (0,10,12,14,16,18)),
 INDEX ix_works_public(status,minimum_age,id), INDEX ix_works_title(title)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE work_authors (
 work_id BIGINT UNSIGNED NOT NULL, author_id BIGINT UNSIGNED NOT NULL,
 PRIMARY KEY(work_id,author_id), INDEX ix_author_works(author_id,work_id),
 CONSTRAINT fk_wa_work FOREIGN KEY(work_id) REFERENCES works(id) ON DELETE RESTRICT,
 CONSTRAINT fk_wa_author FOREIGN KEY(author_id) REFERENCES authors(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE work_genres (
 work_id BIGINT UNSIGNED NOT NULL, genre_id BIGINT UNSIGNED NOT NULL,
 PRIMARY KEY(work_id,genre_id), INDEX ix_genre_works(genre_id,work_id),
 CONSTRAINT fk_wg_work FOREIGN KEY(work_id) REFERENCES works(id) ON DELETE RESTRICT,
 CONSTRAINT fk_wg_genre FOREIGN KEY(genre_id) REFERENCES genres(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE editions (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 work_id BIGINT UNSIGNED NOT NULL, publisher_id BIGINT UNSIGNED NOT NULL,
 title VARCHAR(300) NOT NULL, isbn VARCHAR(13) CHARACTER SET ascii COLLATE ascii_bin NULL,
 isbn_absence_reason VARCHAR(300) NOT NULL DEFAULT '', edition_number INT UNSIGNED NOT NULL,
 page_count INT UNSIGNED NOT NULL, publication_date DATE NULL,
 language VARCHAR(35) NOT NULL, image_url VARCHAR(2048) NOT NULL DEFAULT '',
 status VARCHAR(20) NOT NULL DEFAULT 'draft', minimum_age TINYINT UNSIGNED NOT NULL DEFAULT 0,
 version BIGINT UNSIGNED NOT NULL DEFAULT 1,
 created_at DATETIME(6) NOT NULL DEFAULT UTC_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT UTC_TIMESTAMP(6),
 UNIQUE KEY uq_editions_isbn(isbn), UNIQUE KEY uq_edition_work(id,work_id),
 CONSTRAINT fk_edition_work FOREIGN KEY(work_id) REFERENCES works(id) ON DELETE RESTRICT,
 CONSTRAINT fk_edition_publisher FOREIGN KEY(publisher_id) REFERENCES publishers(id) ON DELETE RESTRICT,
 CONSTRAINT ck_edition_status CHECK (status IN ('draft','pending_review','published','rejected','hidden','archived')),
 CONSTRAINT ck_edition_age CHECK (minimum_age IN (0,10,12,14,16,18)),
 CONSTRAINT ck_edition_pages CHECK (page_count > 0), CONSTRAINT ck_edition_number CHECK (edition_number > 0),
 CONSTRAINT ck_edition_isbn CHECK ((isbn IS NOT NULL AND CHAR_LENGTH(isbn)=13) OR CHAR_LENGTH(isbn_absence_reason)>0),
 INDEX ix_editions_work(work_id,status,id), INDEX ix_editions_publisher(publisher_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
ALTER TABLE readers ADD CONSTRAINT fk_reader_favorite FOREIGN KEY(favorite_work_id) REFERENCES works(id) ON DELETE RESTRICT;
CREATE TABLE reader_genres (
 reader_id BIGINT UNSIGNED NOT NULL, genre_id BIGINT UNSIGNED NOT NULL,
 PRIMARY KEY(reader_id,genre_id),
 CONSTRAINT fk_rg_reader FOREIGN KEY(reader_id) REFERENCES readers(user_id) ON DELETE RESTRICT,
 CONSTRAINT fk_rg_genre FOREIGN KEY(genre_id) REFERENCES genres(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE reader_preferred_works (
 reader_id BIGINT UNSIGNED NOT NULL, work_id BIGINT UNSIGNED NOT NULL,
 PRIMARY KEY(reader_id,work_id),
 CONSTRAINT fk_rpw_reader FOREIGN KEY(reader_id) REFERENCES readers(user_id) ON DELETE RESTRICT,
 CONSTRAINT fk_rpw_work FOREIGN KEY(work_id) REFERENCES works(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE catalog_suggestions (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 reader_id BIGINT UNSIGNED NOT NULL, operation VARCHAR(10) NOT NULL, resource VARCHAR(10) NOT NULL,
 target_id BIGINT UNSIGNED NULL, expected_version BIGINT UNSIGNED NULL, proposal JSON NOT NULL,
 status VARCHAR(15) NOT NULL DEFAULT 'pending', decision_by BIGINT UNSIGNED NULL,
 decision_reason VARCHAR(1000) NOT NULL DEFAULT '', decided_at DATETIME(6) NULL,
 created_at DATETIME(6) NOT NULL DEFAULT UTC_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT UTC_TIMESTAMP(6),
 CONSTRAINT fk_suggestion_reader FOREIGN KEY(reader_id) REFERENCES readers(user_id) ON DELETE RESTRICT,
 CONSTRAINT fk_suggestion_admin FOREIGN KEY(decision_by) REFERENCES administrators(user_id) ON DELETE RESTRICT,
 CONSTRAINT ck_suggestion_op CHECK (operation IN ('create','update','delete')),
 CONSTRAINT ck_suggestion_resource CHECK (resource IN ('work','edition')),
 CONSTRAINT ck_suggestion_status CHECK (status IN ('pending','approved','rejected')),
 CONSTRAINT ck_suggestion_target CHECK ((operation='create' AND target_id IS NULL AND expected_version IS NULL) OR (operation IN ('update','delete') AND target_id IS NOT NULL AND expected_version IS NOT NULL)),
 INDEX ix_suggestions_queue(status,created_at,id), INDEX ix_suggestions_reader(reader_id,id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE catalog_suggestions;
DROP TABLE reader_preferred_works;
DROP TABLE reader_genres;
ALTER TABLE readers DROP FOREIGN KEY fk_reader_favorite;
DROP TABLE editions;
DROP TABLE work_genres;
DROP TABLE work_authors;
DROP TABLE works;
DROP TABLE genres;
DROP TABLE publishers;
DROP TABLE authors;
