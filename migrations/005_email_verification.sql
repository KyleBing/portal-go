-- 邮箱验证时间：NULL 表示未验证；已有用户迁移时全部标为已验证
ALTER TABLE `users`
  ADD COLUMN `email_verified_at` DATETIME NULL COMMENT '邮箱验证时间' AFTER `password`;

UPDATE `users` SET `email_verified_at` = COALESCE(`register_time`, NOW()) WHERE `email_verified_at` IS NULL;

CREATE TABLE IF NOT EXISTS `email_tokens` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `uid` INT(11) NOT NULL,
  `purpose` ENUM('verify', 'reset') NOT NULL,
  `token_hash` CHAR(64) NOT NULL COMMENT 'SHA-256 hex of raw token',
  `expires_at` DATETIME NOT NULL,
  `used_at` DATETIME NULL,
  PRIMARY KEY (`id`),
  KEY `idx_email_tokens_hash` (`token_hash`),
  KEY `idx_email_tokens_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
