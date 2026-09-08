-- seed new diary categories introduced after portal 9b4a6a75 (idempotent)
INSERT IGNORE INTO `diary_category` (`sort_id`, `name_en`, `name`, `count`, `color`, `date_init`)
VALUES
  (11, 'enlightenment', '感悟', 0, '#555C8C', '2025-07-09 03:35:52'),
  (16, 'code', '代码', 0, '#04999B', '2026-06-06 06:36:23');
