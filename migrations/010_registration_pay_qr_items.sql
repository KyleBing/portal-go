-- 注册申请收款码列表：任意金额，每档支付宝/微信各一张图和备注
CREATE TABLE IF NOT EXISTS `registration_pay_qr_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `amount_yuan` int NOT NULL COMMENT '应付金额（元），0 表示自定义金额',
  `alipay_url` varchar(1024) NOT NULL DEFAULT '' COMMENT '支付宝收款码图片地址',
  `alipay_note` varchar(255) NOT NULL DEFAULT '' COMMENT '支付宝收款码备注',
  `wechat_url` varchar(1024) NOT NULL DEFAULT '' COMMENT '微信收款码图片地址',
  `wechat_note` varchar(255) NOT NULL DEFAULT '' COMMENT '微信收款码备注',
  `date_create` datetime NOT NULL,
  `date_modify` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_reg_pay_qr_amount` (`amount_yuan`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='注册申请收款码（按金额）';
