-- 去掉高德组队码：用户资料字段，以及二维码上的展示开关。
ALTER TABLE `users` DROP COLUMN `gaode`;
ALTER TABLE `qrs` DROP COLUMN `is_show_gaode`;
