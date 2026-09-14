-- invitations: 标记邀请码是否已分享（旧版 init.sql 缺少该字段；已有列则跳过）
SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'invitations'
    AND COLUMN_NAME = 'is_shared'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE `invitations` ADD COLUMN `is_shared` int(1) NOT NULL DEFAULT 0 COMMENT ''标记邀请码是否已被分享'' AFTER `binding_uid`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
