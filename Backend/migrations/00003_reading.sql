-- +goose Up
CREATE TABLE reading_records (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 reader_id BIGINT UNSIGNED NOT NULL, work_id BIGINT UNSIGNED NOT NULL, edition_id BIGINT UNSIGNED NOT NULL,
 current_page INT NOT NULL DEFAULT 0, credited_peak INT NOT NULL DEFAULT 0,
 status ENUM('planned','in_progress','paused','completed','abandoned') NOT NULL,
 started_at DATETIME(6) NULL, finished_at DATETIME(6) NULL,
 created_at DATETIME(6) NOT NULL, updated_at DATETIME(6) NOT NULL,
 UNIQUE KEY uq_shelf_reader_work(reader_id,work_id), KEY ix_shelf_status(reader_id,status,id),
 CONSTRAINT fk_shelf_reader FOREIGN KEY(reader_id) REFERENCES readers(user_id),
 CONSTRAINT fk_shelf_edition_work FOREIGN KEY(edition_id,work_id) REFERENCES editions(id,work_id),
 CHECK(current_page>=0 AND credited_peak>=0), CHECK(finished_at IS NULL OR started_at IS NULL OR finished_at>=started_at),
 CHECK(status<>'planned' OR current_page=0), CHECK(status<>'completed' OR finished_at IS NOT NULL)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE reading_progress_events (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, record_id BIGINT UNSIGNED NOT NULL, reader_id BIGINT UNSIGNED NOT NULL,
 previous_page INT NOT NULL, current_page INT NOT NULL, credited_pages INT NOT NULL,
 change_kind ENUM('advance','correction','edition_change','snapshot') NOT NULL,
 created_at DATETIME(6) NOT NULL, KEY ix_progress_reader_date(reader_id,created_at,id),
 FOREIGN KEY(record_id) REFERENCES reading_records(id) ON DELETE CASCADE,
 FOREIGN KEY(reader_id) REFERENCES readers(user_id),
 CHECK(previous_page>=0 AND current_page>=0 AND credited_pages>=0),
 CHECK(change_kind='advance' OR credited_pages=0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE posts (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
 reader_id BIGINT UNSIGNED NOT NULL, work_id BIGINT UNSIGNED NOT NULL,
 kind ENUM('diary','review') NOT NULL, title VARCHAR(200) NOT NULL DEFAULT '', text TEXT NOT NULL,
 data_url VARCHAR(2048) NULL, status ENUM('draft','published','hidden','removed','deleted') NOT NULL,
 contains_spoiler BOOLEAN NOT NULL, contains_sensitive_content BOOLEAN NOT NULL DEFAULT FALSE,
 minimum_age SMALLINT NOT NULL DEFAULT 0, published_at DATETIME(6) NULL, effects_applied_at DATETIME(6) NULL,
 created_at DATETIME(6) NOT NULL, updated_at DATETIME(6) NOT NULL,
 UNIQUE KEY uq_post_kind(id,kind), UNIQUE KEY uq_post_author_work(id,reader_id,work_id),
 KEY ix_feed_status_date(status,created_at,id), KEY ix_post_reader_kind(reader_id,kind,status,id), KEY ix_post_work(work_id,status,id),
 FOREIGN KEY(reader_id) REFERENCES readers(user_id), FOREIGN KEY(work_id) REFERENCES works(id),
 CHECK(minimum_age IN(0,10,12,14,16,18)), CHECK(CHAR_LENGTH(text)>0), CHECK(contains_spoiler IN(0,1)),CHECK(contains_sensitive_content IN(0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE reading_diaries (
 post_id BIGINT UNSIGNED NOT NULL PRIMARY KEY, kind ENUM('diary','review') NOT NULL DEFAULT 'diary',
 edition_id BIGINT UNSIGNED NOT NULL, current_page INT NOT NULL, occurred_at DATETIME(6) NULL,
 FOREIGN KEY(post_id,kind) REFERENCES posts(id,kind), FOREIGN KEY(edition_id) REFERENCES editions(id),
 CHECK(kind='diary'), CHECK(current_page>=0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE reviews (
 post_id BIGINT UNSIGNED NOT NULL PRIMARY KEY, kind ENUM('diary','review') NOT NULL DEFAULT 'review',
 reader_id BIGINT UNSIGNED NOT NULL, work_id BIGINT UNSIGNED NOT NULL,
 rating SMALLINT NOT NULL, active_slot SMALLINT NULL DEFAULT 1,
 UNIQUE KEY uq_review_active(reader_id,work_id,active_slot), KEY ix_review_work(work_id,rating),
 FOREIGN KEY(post_id,kind) REFERENCES posts(id,kind),
 FOREIGN KEY(post_id,reader_id,work_id) REFERENCES posts(id,reader_id,work_id),
 CHECK(kind='review'), CHECK(rating BETWEEN 1 AND 5), CHECK(active_slot IS NULL OR active_slot=1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE comments (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, post_id BIGINT UNSIGNED NOT NULL, reader_id BIGINT UNSIGNED NOT NULL,
 parent_id BIGINT UNSIGNED NULL, depth INT NOT NULL DEFAULT 0, text TEXT NOT NULL,
 status ENUM('published','hidden','removed','deleted') NOT NULL DEFAULT 'published',
 created_at DATETIME(6) NOT NULL, updated_at DATETIME(6) NOT NULL,
 UNIQUE KEY uq_comment_post(id,post_id), KEY ix_comment_tree(post_id,parent_id,created_at,id),
 FOREIGN KEY(post_id) REFERENCES posts(id), FOREIGN KEY(reader_id) REFERENCES readers(user_id),
 FOREIGN KEY(parent_id,post_id) REFERENCES comments(id,post_id), CHECK(depth BETWEEN 0 AND 20)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose StatementBegin
CREATE TRIGGER trg_comment_parent_insert BEFORE INSERT ON comments FOR EACH ROW
BEGIN
 IF NEW.parent_id IS NOT NULL AND NEW.id<>0 AND NEW.parent_id=NEW.id THEN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='A comment cannot reference itself';
 END IF;
END;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TRIGGER trg_comment_parent_immutable BEFORE UPDATE ON comments FOR EACH ROW
BEGIN
 IF NOT (NEW.parent_id <=> OLD.parent_id) OR NEW.post_id<>OLD.post_id THEN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='Comment parent and post are immutable';
 END IF;
END;
-- +goose StatementEnd
CREATE TABLE post_likes (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, reader_id BIGINT UNSIGNED NOT NULL, post_id BIGINT UNSIGNED NOT NULL,
 created_at DATETIME(6) NOT NULL, UNIQUE KEY uq_post_like(reader_id,post_id),KEY ix_post_like(post_id),
 FOREIGN KEY(reader_id) REFERENCES readers(user_id),FOREIGN KEY(post_id) REFERENCES posts(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE comment_likes (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, reader_id BIGINT UNSIGNED NOT NULL, comment_id BIGINT UNSIGNED NOT NULL,
 created_at DATETIME(6) NOT NULL, UNIQUE KEY uq_comment_like(reader_id,comment_id),KEY ix_comment_like(comment_id),
 FOREIGN KEY(reader_id) REFERENCES readers(user_id),FOREIGN KEY(comment_id) REFERENCES comments(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE TABLE notifications (
 id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, recipient_id BIGINT UNSIGNED NOT NULL,actor_id BIGINT UNSIGNED NULL,
 event VARCHAR(40) NOT NULL,post_id BIGINT UNSIGNED NULL,comment_id BIGINT UNSIGNED NULL,event_key VARCHAR(191) NOT NULL,
 read_at DATETIME(6) NULL,created_at DATETIME(6) NOT NULL,
 UNIQUE KEY uq_notification_event(event_key),KEY ix_notification_recipient(recipient_id,read_at,created_at,id),
 FOREIGN KEY(recipient_id) REFERENCES readers(user_id),FOREIGN KEY(actor_id) REFERENCES users(id),
 FOREIGN KEY(post_id) REFERENCES posts(id), FOREIGN KEY(comment_id) REFERENCES comments(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose Down
DROP TABLE notifications;
DROP TABLE comment_likes;
DROP TABLE post_likes;
DROP TABLE comments;
DROP TABLE reviews;
DROP TABLE reading_diaries;
DROP TABLE posts;
DROP TABLE reading_progress_events;
DROP TABLE reading_records;
