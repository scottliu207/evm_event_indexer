CREATE DATABASE IF NOT EXISTS `account_db`;

USE `account_db`;

-- user
CREATE TABLE `account_db`.`user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `account` varchar(20) NOT NULL COMMENT 'user account',
  `status` tinyint unsigned NOT NULL DEFAULT 1 COMMENT 'user status (1: enabled, 2: disabled)',
  `password` varchar(255) NOT NULL COMMENT 'user password',
  `auth_meta` json NOT NULL COMMENT 'user authentication metadata',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created at',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated at',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`account`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='user';



