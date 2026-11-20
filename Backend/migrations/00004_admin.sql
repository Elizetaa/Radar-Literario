-- +goose Up
CREATE TABLE reports (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 reporter_id BIGINT UNSIGNED NOT NULL,
 reported_user_id BIGINT UNSIGNED NULL,
 reported_post_id BIGINT UNSIGNED NULL,
 reported_comment_id BIGINT UNSIGNED NULL,
 assigned_to BIGINT UNSIGNED NULL,
 reason VARCHAR(40) NOT NULL,
 description TEXT NOT NULL,
 report_data_url VARCHAR(2048) NOT NULL DEFAULT '',
 status VARCHAR(32) NOT NULL DEFAULT 'open',
 resolution_note TEXT NULL,
 resolved_at DATETIME(6) NULL,
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 CONSTRAINT fk_report_reporter FOREIGN KEY(reporter_id) REFERENCES readers(user_id) ON DELETE RESTRICT,
 CONSTRAINT fk_report_user FOREIGN KEY(reported_user_id) REFERENCES users(id) ON DELETE RESTRICT,
 CONSTRAINT fk_report_post FOREIGN KEY(reported_post_id) REFERENCES posts(id) ON DELETE RESTRICT,
 CONSTRAINT fk_report_comment FOREIGN KEY(reported_comment_id) REFERENCES comments(id) ON DELETE RESTRICT,
 CONSTRAINT fk_report_assignee FOREIGN KEY(assigned_to) REFERENCES administrators(user_id) ON DELETE RESTRICT,
 CONSTRAINT ck_report_target CHECK ((reported_user_id IS NOT NULL)+(reported_post_id IS NOT NULL)+(reported_comment_id IS NOT NULL)=1),
 CONSTRAINT ck_report_reason CHECK(reason IN ('spam','harassment','hate_speech','threats','inappropriate_content','privacy_violation','copyright_violation','impersonation','misinformation','other')),
 CONSTRAINT ck_report_status CHECK(status IN ('open','in_review','awaiting_information','resolved','dismissed','canceled')),
 CONSTRAINT ck_report_resolution CHECK(status NOT IN ('resolved','dismissed') OR (resolution_note IS NOT NULL AND CHAR_LENGTH(TRIM(resolution_note))>=3 AND resolved_at IS NOT NULL)),
 INDEX ix_report_queue(status,created_at,id),
 INDEX ix_report_reader(reporter_id,created_at,id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE report_messages (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 report_id BIGINT UNSIGNED NOT NULL,
 author_id BIGINT UNSIGNED NOT NULL,
 text TEXT NOT NULL,
 visibility VARCHAR(16) NOT NULL,
 created_at DATETIME(6) NOT NULL,
 CONSTRAINT fk_report_message_report FOREIGN KEY(report_id) REFERENCES reports(id) ON DELETE RESTRICT,
 CONSTRAINT fk_report_message_author FOREIGN KEY(author_id) REFERENCES users(id) ON DELETE RESTRICT,
 CONSTRAINT ck_report_message_visibility CHECK(visibility IN ('reporter','internal')),
 INDEX ix_report_message(report_id,id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE penalties (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 user_id BIGINT UNSIGNED NOT NULL,
 applied_by BIGINT UNSIGNED NOT NULL,
 revoked_by BIGINT UNSIGNED NULL,
 type VARCHAR(24) NOT NULL,
 reason TEXT NOT NULL,
 starts_at DATETIME(6) NOT NULL,
 ends_at DATETIME(6) NULL,
 revoked_at DATETIME(6) NULL,
 revocation_reason TEXT NULL,
 created_at DATETIME(6) NOT NULL,
 CONSTRAINT fk_penalty_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE RESTRICT,
 CONSTRAINT fk_penalty_applied_by FOREIGN KEY(applied_by) REFERENCES administrators(user_id) ON DELETE RESTRICT,
 CONSTRAINT fk_penalty_revoked_by FOREIGN KEY(revoked_by) REFERENCES administrators(user_id) ON DELETE RESTRICT,
 CONSTRAINT ck_penalty_type CHECK(type IN ('warning','suspension','restriction','ban')),
 CONSTRAINT ck_penalty_period CHECK((type IN ('warning','ban') AND ends_at IS NULL) OR (type IN ('suspension','restriction') AND ends_at IS NOT NULL AND ends_at>starts_at)),
 CONSTRAINT ck_penalty_revocation CHECK((revoked_at IS NULL AND revoked_by IS NULL AND revocation_reason IS NULL) OR (revoked_at IS NOT NULL AND revoked_by IS NOT NULL AND revocation_reason IS NOT NULL)),
 INDEX ix_penalty_effect(user_id,revoked_at,starts_at,ends_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE report_penalties (
 report_id BIGINT UNSIGNED NOT NULL,
 penalty_id BIGINT UNSIGNED NOT NULL,
 PRIMARY KEY(report_id,penalty_id),
 CONSTRAINT fk_report_penalty_report FOREIGN KEY(report_id) REFERENCES reports(id) ON DELETE RESTRICT,
 CONSTRAINT fk_report_penalty_penalty FOREIGN KEY(penalty_id) REFERENCES penalties(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE faq_entries (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 question VARCHAR(500) NOT NULL,
 answer TEXT NOT NULL,
 sort_order INT NOT NULL DEFAULT 0,
 status VARCHAR(16) NOT NULL DEFAULT 'draft',
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 CONSTRAINT ck_faq_status CHECK(status IN ('draft','published')),
 CONSTRAINT ck_faq_order CHECK(sort_order>=0),
 INDEX ix_faq_public(status,sort_order,id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE support_tickets (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 requester_id BIGINT UNSIGNED NOT NULL,
 assigned_to BIGINT UNSIGNED NULL,
 subject VARCHAR(200) NOT NULL,
 status VARCHAR(24) NOT NULL DEFAULT 'open',
 created_at DATETIME(6) NOT NULL,
 updated_at DATETIME(6) NOT NULL,
 CONSTRAINT fk_ticket_requester FOREIGN KEY(requester_id) REFERENCES readers(user_id) ON DELETE RESTRICT,
 CONSTRAINT fk_ticket_assignee FOREIGN KEY(assigned_to) REFERENCES administrators(user_id) ON DELETE RESTRICT,
 CONSTRAINT ck_ticket_status CHECK(status IN ('open','in_progress','resolved','closed')),
 INDEX ix_ticket_queue(status,created_at,id),
 INDEX ix_ticket_owner(requester_id,created_at,id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE support_messages (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 ticket_id BIGINT UNSIGNED NOT NULL,
 author_id BIGINT UNSIGNED NOT NULL,
 text TEXT NOT NULL,
 created_at DATETIME(6) NOT NULL,
 CONSTRAINT fk_support_message_ticket FOREIGN KEY(ticket_id) REFERENCES support_tickets(id) ON DELETE RESTRICT,
 CONSTRAINT fk_support_message_author FOREIGN KEY(author_id) REFERENCES users(id) ON DELETE RESTRICT,
 INDEX ix_support_message(ticket_id,id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE support_messages;
DROP TABLE support_tickets;
DROP TABLE faq_entries;
DROP TABLE report_penalties;
DROP TABLE penalties;
DROP TABLE report_messages;
DROP TABLE reports;
