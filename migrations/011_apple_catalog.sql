-- 苹果设备与芯片。一行一条，展示字段在 record_json，原始嵌套参数在 source_json。
CREATE TABLE IF NOT EXISTS `apple_devices` (
  `id` varchar(160) NOT NULL,
  `family` varchar(16) NOT NULL,
  `series` varchar(32) NOT NULL DEFAULT '',
  `name` varchar(160) NOT NULL,
  `name_short` varchar(80) NOT NULL DEFAULT '',
  `release_date` varchar(20) NOT NULL DEFAULT '',
  `year` smallint NOT NULL DEFAULT 0,
  `sort_order` int NOT NULL DEFAULT 0,
  `record_json` longtext NOT NULL,
  `source_json` longtext NULL,
  `date_modify` datetime NULL,
  PRIMARY KEY (`id`),
  KEY `idx_apple_device_family` (`family`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `apple_chips` (
  `id` varchar(160) NOT NULL,
  `family` varchar(16) NOT NULL DEFAULT '',
  `series` varchar(16) NOT NULL,
  `name` varchar(160) NOT NULL,
  `name_short` varchar(80) NOT NULL DEFAULT '',
  `release_date` varchar(20) NOT NULL DEFAULT '',
  `year` smallint NOT NULL DEFAULT 0,
  `sort_order` int NOT NULL DEFAULT 0,
  `record_json` longtext NOT NULL,
  `source_json` longtext NULL,
  `date_modify` datetime NULL,
  PRIMARY KEY (`id`),
  KEY `idx_apple_chip_series` (`series`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
