-- user_config: per-user UI preferences (theme / default category / editor)
CREATE TABLE IF NOT EXISTS `user_config` (
  `uid` int(11) NOT NULL COMMENT '用户 ID',
  `theme` varchar(20) NOT NULL DEFAULT '' COMMENT '主题',
  `default_diary_category` varchar(50) NOT NULL DEFAULT '' COMMENT '默认日记分类',
  `editor_mode` varchar(20) NOT NULL DEFAULT '' COMMENT '编辑器模式',
  `config_json` json DEFAULT NULL COMMENT '扩展配置',
  `date_modify` datetime DEFAULT NULL COMMENT '最后修改时间',
  PRIMARY KEY (`uid`) USING BTREE,
  CONSTRAINT `user_config_uid` FOREIGN KEY (`uid`) REFERENCES `users` (`uid`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=DYNAMIC;
