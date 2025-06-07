-- ========= RESET =========
DROP DATABASE IF EXISTS toolcenter;
CREATE DATABASE toolcenter
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_general_ci;
USE toolcenter;

-- ========= CORE =========
CREATE TABLE users (
  user_id CHAR(36) PRIMARY KEY,               -- UUID v7 string
  username VARCHAR(50) UNIQUE NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  ip_address VARCHAR(45),
  password_hash VARCHAR(255) NOT NULL,
  avatar_url VARCHAR(255),
  banner_url VARCHAR(255),
  bio TEXT,
  is_verified TINYINT(1) DEFAULT 0,
  role ENUM('User','Moderator','Admin') DEFAULT 'User',
  account_status ENUM('Good','Limited','Very Limited','At Risk','Banned') DEFAULT 'Good',
  email_verified_at DATETIME,
  authenticator_secret VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  username_changed_at DATETIME,
  email_changed_at DATETIME,
  avatar_changed_at DATETIME,
  banner_changed_at DATETIME,
  password_changed_at DATETIME,
  last_login DATETIME,
  last_tool_posted DATETIME,
  last_tool_updated DATETIME,
  INDEX idx_users_created (created_at)
) ENGINE = InnoDB;

-- ========= AUTH & TOKENS =========
CREATE TABLE email_verification_tokens (
  token_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id  CHAR(36) NOT NULL,
  token    CHAR(64) UNIQUE NOT NULL,
  expires_at DATETIME NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE password_resets (
  reset_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id  CHAR(36) NOT NULL,
  token    CHAR(64) UNIQUE NOT NULL,
  expires_at DATETIME NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE user_tokens (
  token_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id  CHAR(36) NOT NULL,
  token CHAR(128) UNIQUE NOT NULL,
  device_info TEXT,
  expires_at DATETIME,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
) ENGINE = InnoDB;

-- ========= CONTENT =========
CREATE TABLE tags (
  tag_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name   VARCHAR(50) UNIQUE NOT NULL
) ENGINE = InnoDB;

CREATE TABLE tools (
  tool_id CHAR(36) PRIMARY KEY,
  user_id CHAR(36) NOT NULL,
  title VARCHAR(100),
  description TEXT,
  content_url VARCHAR(255),
  thumbnail_url VARCHAR(255),
  status ENUM('Published','Moderated','Hidden') DEFAULT 'Published',
  views INT UNSIGNED DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
) ENGINE = InnoDB ROW_FORMAT = COMPRESSED;

CREATE TABLE tool_tags (
  tool_id CHAR(36) NOT NULL,
  tag_id  INT UNSIGNED NOT NULL,
  PRIMARY KEY (tool_id, tag_id),
  FOREIGN KEY (tool_id) REFERENCES tools (tool_id) ON DELETE CASCADE,
  FOREIGN KEY (tag_id)  REFERENCES tags  (tag_id)  ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE tool_versions (
  version_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  tool_id CHAR(36) NOT NULL,
  title VARCHAR(100),
  description TEXT,
  content_url VARCHAR(255),
  thumbnail_url VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (tool_id) REFERENCES tools (tool_id) ON DELETE CASCADE
) ENGINE = InnoDB;

-- ========= SOCIAL =========
CREATE TABLE comments (
  comment_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  tool_id CHAR(36) NOT NULL,
  user_id CHAR(36) NOT NULL,
  parent_comment_id INT UNSIGNED,
  content TEXT,
  status ENUM('Visible','Hidden','Moderated') DEFAULT 'Visible',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (tool_id) REFERENCES tools (tool_id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
  FOREIGN KEY (parent_comment_id) REFERENCES comments (comment_id) ON DELETE CASCADE,
  INDEX idx_comments_user_created (user_id, created_at)
) ENGINE = InnoDB ROW_FORMAT = COMPRESSED;

CREATE TABLE favorites (
  favorite_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  tool_id CHAR(36) NOT NULL,
  user_id CHAR(36) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (tool_id, user_id),
  FOREIGN KEY (tool_id) REFERENCES tools (tool_id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
  INDEX idx_fav_user (user_id, created_at)
) ENGINE = InnoDB;

CREATE TABLE likes (
  like_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  tool_id CHAR(36) NOT NULL,
  user_id CHAR(36) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (tool_id, user_id),
  FOREIGN KEY (tool_id) REFERENCES tools (tool_id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
  INDEX idx_likes_user (user_id, created_at)
) ENGINE = InnoDB;

CREATE TABLE comment_likes (
  like_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  comment_id INT UNSIGNED NOT NULL,
  user_id CHAR(36) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (comment_id, user_id),
  FOREIGN KEY (comment_id) REFERENCES comments (comment_id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
) ENGINE = InnoDB;

-- ========= BUSINESS =========
CREATE TABLE subscriptions (
  subscription_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id CHAR(36) NOT NULL,
  type ENUM('basic','premium') DEFAULT 'basic',
  start_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  end_date DATETIME,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE payments (
  payment_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id CHAR(36) NOT NULL,
  provider VARCHAR(50),
  amount DECIMAL(10,2),
  currency CHAR(3),
  status ENUM('pending','succeeded','failed','refunded'),
  reference VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  CHECK (currency IN ('EUR','USD','GBP')),
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
) ENGINE = InnoDB;

-- ========= MODERATION =========
CREATE TABLE user_warns (
  warn_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id CHAR(36) NOT NULL,
  reason TEXT,
  warn_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME,
  moderator_id CHAR(36),
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
  FOREIGN KEY (moderator_id) REFERENCES users (user_id) ON DELETE SET NULL
) ENGINE = InnoDB;

CREATE TABLE moderation_actions (
  action_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  moderator_id CHAR(36),
  user_id CHAR(36) NOT NULL,
  action_type ENUM('Warn','Ban','Unban','Limit_Comments','Limit_Tools'),
  reason TEXT,
  start_date TIMESTAMP,
  end_date TIMESTAMP,
  action_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (moderator_id) REFERENCES users (user_id) ON DELETE SET NULL,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE reports (
  report_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  reporter_id CHAR(36) NOT NULL,
  target_type ENUM('Tool','Comment','User') NOT NULL,
  target_id INT UNSIGNED NOT NULL,
  reason TEXT,
  status ENUM('Open','Reviewed','Closed') DEFAULT 'Open',
  reviewed_by CHAR(36),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (reporter_id) REFERENCES users (user_id) ON DELETE CASCADE,
  FOREIGN KEY (reviewed_by) REFERENCES users (user_id) ON DELETE SET NULL
) ENGINE = InnoDB;

-- ========= NOTIFS & LOGS =========
CREATE TABLE notifications (
  notification_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id CHAR(36) NOT NULL,
  title VARCHAR(255),
  description TEXT,
  is_read BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE audit_logs (
  audit_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  event_type ENUM(
    'Ban','Unban','Warn','DeleteComment','DeleteTool',
    'ChangeRole','EditTool','EditComment','Payment',
    'Login','Logout'
  ) NOT NULL,
  actor_user_id  CHAR(36),
  target_user_id CHAR(36),
  target_resource VARCHAR(255),
  payload JSON,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_audit_event_time (event_type, created_at),
  FOREIGN KEY (actor_user_id)  REFERENCES users (user_id) ON DELETE SET NULL,
  FOREIGN KEY (target_user_id) REFERENCES users (user_id) ON DELETE SET NULL
) ENGINE = InnoDB ROW_FORMAT = COMPRESSED;

-- ========= STATS & SOCIAL =========
CREATE TABLE user_stats (
  stat_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id CHAR(36) NOT NULL,
  tools_posted_count INT UNSIGNED DEFAULT 0,
  comments_count INT UNSIGNED DEFAULT 0,
  likes_given INT UNSIGNED DEFAULT 0,
  likes_received INT UNSIGNED DEFAULT 0,
  favorites_count INT UNSIGNED DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE friends (
  friendship_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id  CHAR(36) NOT NULL,
  friend_id CHAR(36) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (user_id, friend_id),
  FOREIGN KEY (user_id)  REFERENCES users (user_id) ON DELETE CASCADE,
  FOREIGN KEY (friend_id) REFERENCES users (user_id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE friend_requests (
  request_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  from_id CHAR(36) NOT NULL,
  to_id   CHAR(36) NOT NULL,
  status ENUM('Pending','Accepted','Declined') DEFAULT 'Pending',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (from_id, to_id),
  FOREIGN KEY (from_id) REFERENCES users (user_id) ON DELETE CASCADE,
  FOREIGN KEY (to_id)   REFERENCES users (user_id) ON DELETE CASCADE
) ENGINE = InnoDB;

-- ========= MAIL QUEUE =========
CREATE TABLE email_queue (
  queue_id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  to_email VARCHAR(255),
  subject  VARCHAR(255),
  body     TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB;
