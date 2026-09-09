-- diaries: speed up uid-scoped list / bill / calendar queries
CREATE INDEX `idx_diaries_uid_date` ON `diaries` (`uid`, `date`);
