-- 注册申请收款码图片地址（支付宝 / 微信各一张，所有金额档位共用）
SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'system_config' AND COLUMN_NAME = 'registration_pay_qr_alipay'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE `system_config` ADD COLUMN `registration_pay_qr_alipay` varchar(1024) NOT NULL DEFAULT '''' COMMENT ''支付宝收款码图片地址'' AFTER `registration_pay_amount_yuan`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'system_config' AND COLUMN_NAME = 'registration_pay_qr_wechat'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE `system_config` ADD COLUMN `registration_pay_qr_wechat` varchar(1024) NOT NULL DEFAULT '''' COMMENT ''微信收款码图片地址'' AFTER `registration_pay_qr_alipay`',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
