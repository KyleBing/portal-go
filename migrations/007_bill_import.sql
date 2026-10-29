-- 支付宝账单导入：别名配置、导入批次、订单号去重
CREATE TABLE IF NOT EXISTS `bill_alias_config` (
  `uid` int(11) NOT NULL COMMENT '用户 ID',
  `slots` json NOT NULL COMMENT '餐次时段',
  `rules` json NOT NULL COMMENT '有序别名规则',
  `date_modify` datetime DEFAULT NULL COMMENT '最后修改时间',
  PRIMARY KEY (`uid`) USING BTREE,
  CONSTRAINT `bill_alias_config_uid` FOREIGN KEY (`uid`) REFERENCES `users` (`uid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS `bill_import_batch` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL COMMENT '用户 ID',
  `filename` varchar(255) NOT NULL DEFAULT '' COMMENT '原文件名',
  `stored_path` varchar(512) NOT NULL DEFAULT '' COMMENT '相对 upload 路径',
  `period_start` datetime DEFAULT NULL COMMENT '流水起始',
  `period_end` datetime DEFAULT NULL COMMENT '流水终止',
  `row_count` int(11) NOT NULL DEFAULT 0,
  `ready_count` int(11) NOT NULL DEFAULT 0,
  `guess_count` int(11) NOT NULL DEFAULT 0,
  `confirm_count` int(11) NOT NULL DEFAULT 0,
  `unmapped_count` int(11) NOT NULL DEFAULT 0,
  `skip_count` int(11) NOT NULL DEFAULT 0,
  `applied_count` int(11) NOT NULL DEFAULT 0,
  `status` varchar(20) NOT NULL DEFAULT 'uploaded' COMMENT 'uploaded/applied/partial',
  `date_init` datetime DEFAULT NULL COMMENT '上传时间',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `bill_import_batch_uid` (`uid`) USING BTREE,
  CONSTRAINT `bill_import_batch_uid` FOREIGN KEY (`uid`) REFERENCES `users` (`uid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=DYNAMIC;

CREATE TABLE IF NOT EXISTS `bill_import_order` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `uid` int(11) NOT NULL COMMENT '用户 ID',
  `order_id` varchar(64) NOT NULL COMMENT '支付宝交易订单号',
  `batch_id` int(11) NOT NULL COMMENT '导入批次',
  `bill_date` date NOT NULL COMMENT '账单日期',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '写入短名称',
  `amount` decimal(12,2) NOT NULL DEFAULT 0.00 COMMENT '带符号金额',
  `date_init` datetime DEFAULT NULL COMMENT '写入时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `bill_import_order_uid_order` (`uid`, `order_id`),
  KEY `bill_import_order_batch` (`batch_id`) USING BTREE,
  CONSTRAINT `bill_import_order_uid` FOREIGN KEY (`uid`) REFERENCES `users` (`uid`) ON DELETE CASCADE,
  CONSTRAINT `bill_import_order_batch` FOREIGN KEY (`batch_id`) REFERENCES `bill_import_batch` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=DYNAMIC;
