-- 付费注册申请台账、邀请码申请人信息、用户注册来源、默认收款 ¥28
CREATE TABLE IF NOT EXISTS `registration_applications` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `email` varchar(255) NOT NULL COMMENT '联系邮箱',
  `paid_account` varchar(255) NOT NULL COMMENT '付费昵称',
  `pay_platform` varchar(32) NOT NULL COMMENT 'alipay|wechat',
  `pay_amount_yuan` int DEFAULT NULL COMMENT '提交申请时系统配置的应付金额（元），0 表示自定义金额',
  `status` varchar(32) NOT NULL DEFAULT 'pending' COMMENT 'pending|approved|rejected',
  `invitation_code` varchar(255) DEFAULT NULL COMMENT '审核通过时下发的邀请码',
  `admin_note` text DEFAULT NULL COMMENT '审核备注',
  `date_create` datetime NOT NULL,
  `date_review` datetime DEFAULT NULL,
  `reviewer_uid` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_reg_apply_email` (`email`),
  KEY `idx_reg_apply_status` (`status`),
  KEY `idx_reg_apply_date_create` (`date_create`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='付费注册申请（台账）';

SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'invitations' AND COLUMN_NAME = 'applicant_email'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE `invitations` ADD COLUMN `applicant_email` varchar(255) DEFAULT NULL COMMENT ''申请邮箱'' AFTER `is_shared`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'invitations' AND COLUMN_NAME = 'payment_user'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE `invitations` ADD COLUMN `payment_user` varchar(255) DEFAULT NULL COMMENT ''支付用户'' AFTER `applicant_email`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'register_channel'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE `users` ADD COLUMN `register_channel` varchar(20) NOT NULL DEFAULT ''unknown'' COMMENT ''注册来源 bootstrap/admin/paid/gift/config/unknown'' AFTER `group_id`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'system_config' AND COLUMN_NAME = 'registration_pay_amount_yuan'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE `system_config` ADD COLUMN `registration_pay_amount_yuan` int NOT NULL DEFAULT 28 COMMENT ''注册申请当前收款金额：0=自定义，6/8/18/28/38/48'' AFTER `register_tip`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 已有绑定邀请码的用户：有申请邮箱视为付费邀请，否则为赠送
UPDATE `users` u
INNER JOIN (
  SELECT `binding_uid`,
    MAX(CASE WHEN `applicant_email` IS NOT NULL AND TRIM(`applicant_email`) <> '' THEN 1 ELSE 0 END) AS is_paid
  FROM `invitations`
  WHERE `binding_uid` IS NOT NULL
  GROUP BY `binding_uid`
) i ON i.`binding_uid` = u.`uid`
SET u.`register_channel` = IF(i.is_paid = 1, 'paid', 'gift')
WHERE u.`register_channel` = 'unknown' OR u.`register_channel` = '';

UPDATE `users` u
INNER JOIN (SELECT MIN(`uid`) AS min_uid FROM `users`) t ON u.`uid` = t.min_uid
SET u.`register_channel` = 'bootstrap'
WHERE u.`register_channel` IN ('unknown', '');
