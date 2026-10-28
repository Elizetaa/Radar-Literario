-- +goose Up
CREATE TABLE users (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 name VARCHAR(120) NOT NULL, nickname VARCHAR(30) COLLATE utf8mb4_bin NOT NULL,
 email VARCHAR(254) COLLATE utf8mb4_bin NOT NULL, password_hash VARCHAR(255) NOT NULL,
 status ENUM('pending_verification','active','inactive','suspended','deactivated','banned') NOT NULL DEFAULT 'active',
 birth_date DATE NULL, profile_image_url VARCHAR(2048) NOT NULL DEFAULT '', email_verified_at DATETIME(6) NULL,
 deleted_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 UNIQUE KEY uq_users_email(email), UNIQUE KEY uq_users_nickname(nickname)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE readers (
 user_id BIGINT UNSIGNED NOT NULL PRIMARY KEY, biography VARCHAR(2000) NOT NULL DEFAULT '',
 profile_visibility ENUM('public','registered_users_only','private') NOT NULL DEFAULT 'public',
 sex VARCHAR(50) NULL, country_of_origin VARCHAR(100) NULL, favorite_work_id BIGINT UNSIGNED NULL,
 CONSTRAINT fk_reader_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE administrators (
 user_id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
 role ENUM('user_manager','catalog_manager','administrator','system_administrator') NOT NULL,
 status ENUM('active','inactive') NOT NULL DEFAULT 'active',
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6), updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 CONSTRAINT fk_admin_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE RESTRICT,
 KEY ix_admin_role_status(role,status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE sessions (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, user_id BIGINT UNSIGNED NOT NULL,
 token_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 csrf_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 created_at DATETIME(6) NOT NULL, last_seen_at DATETIME(6) NOT NULL, expires_at DATETIME(6) NOT NULL, revoked_at DATETIME(6) NULL,
 UNIQUE KEY uq_session_token(token_hash), KEY ix_session_user(user_id,revoked_at), KEY ix_session_expiry(expires_at),
 CONSTRAINT fk_session_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE password_resets (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, user_id BIGINT UNSIGNED NOT NULL,
 token_hash CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 expires_at DATETIME(6) NOT NULL, consumed_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL,
 UNIQUE KEY uq_reset_token(token_hash), KEY ix_reset_expiry(expires_at),
 CONSTRAINT fk_reset_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE audit_logs (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, actor_id BIGINT UNSIGNED NULL,
 target_type VARCHAR(64) NOT NULL, target_id BIGINT UNSIGNED NULL, action VARCHAR(64) NOT NULL, reason TEXT NOT NULL,
 created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
 CONSTRAINT fk_audit_actor FOREIGN KEY(actor_id) REFERENCES users(id) ON DELETE RESTRICT,
 KEY ix_audit_actor(actor_id,created_at,id), KEY ix_audit_target(target_type,target_id,created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE mail_outbox (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, recipient VARCHAR(254) NOT NULL,
 payload MEDIUMTEXT NOT NULL, attempts INT NOT NULL DEFAULT 0, available_at DATETIME(6) NOT NULL,
 claimed_at DATETIME(6) NULL, delivered_at DATETIME(6) NULL, created_at DATETIME(6) NOT NULL,
 KEY ix_outbox_pending(delivered_at,available_at,claimed_at), CHECK(attempts>=0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE mail_outbox;
DROP TABLE audit_logs;
DROP TABLE password_resets;
DROP TABLE sessions;
DROP TABLE administrators;
DROP TABLE readers;
DROP TABLE users;
