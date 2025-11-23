CREATE DATABASE IF NOT EXISTS `event_db`;

USE `event_db`;

-- event log
CREATE TABLE `event_db`.`event_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `address` varchar(128) NOT NULL COMMENT 'contract address',
  `block_hash` varchar(128) NOT NULL COMMENT 'block hash',
  `block_number` bigint unsigned NOT NULL COMMENT 'block number',
  `tx_hash` varchar(128) NOT NULL COMMENT 'tx hash',
  `tx_index` bigint unsigned NOT NULL COMMENT 'tx index',
  `log_index` bigint unsigned NOT NULL COMMENT 'log index',
  `data` blob COMMENT 'data',
  `topics` json NOT NULL COMMENT 'topics',
  `block_timestamp` timestamp NOT NULL COMMENT 'block timestamp',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created at',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`block_number`, `tx_index`, `log_index`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='event log';

-- block syncranization status
CREATE TABLE `event_db`.`block_sync` (
  `address` varchar(128) NOT NULL COMMENT 'contract address',
  `last_sync_number` bigint unsigned NOT NULL COMMENT 'last synced block number',
  `last_sync_hash` varchar(128) NOT NULL COMMENT 'last synced block hash',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated at',
  PRIMARY KEY (`address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='block syncranization status';


-- eth event detail
CREATE TABLE `event_db`.`evm_event_detail` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `address` varchar(128) NOT NULL COMMENT 'contract address',
  `block_hash` varchar(128) NOT NULL COMMENT 'block hash',
  `block_number` bigint unsigned NOT NULL COMMENT 'block number',
  `tx_hash` varchar(128) NOT NULL COMMENT 'tx hash',
  `tx_index` bigint unsigned NOT NULL COMMENT 'tx index',
  `log_index` bigint unsigned NOT NULL COMMENT 'log index',
  `data` blob COMMENT 'data',
  `topics` json NOT NULL COMMENT 'topics',
  `block_timestamp` timestamp NOT NULL COMMENT 'block timestamp',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created at',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`block_number`, `tx_index`, `log_index`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='event log';