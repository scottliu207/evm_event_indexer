CREATE DATABASE IF NOT EXISTS `event_db`;

USE `event_db`;

-- evm transaction event log
CREATE TABLE `event_db`.`event_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `chain_id` bigint unsigned NOT NULL COMMENT 'chain id',
  `address` varchar(128) NOT NULL COMMENT 'contract address',
  `block_hash` varchar(128) NOT NULL COMMENT 'block hash',
  `block_number` bigint unsigned NOT NULL COMMENT 'block number',
  `tx_hash` varchar(128) NOT NULL COMMENT 'tx hash',
  `tx_index` bigint unsigned NOT NULL COMMENT 'tx index',
  `log_index` bigint unsigned NOT NULL COMMENT 'log index',
  `data` blob COMMENT 'data',
  `topic_0` varchar(128) COMMENT 'event signature',
  `topic_1` varchar(128) COMMENT 'indexed parameter 1',
  `topic_2` varchar(128) COMMENT 'indexed parameter 2',
  `topic_3` varchar(128) COMMENT 'indexed parameter 3',
  `decoded_event` json NOT NULL COMMENT 'decoded event',
  `block_timestamp` timestamp NOT NULL COMMENT 'block timestamp',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created at',
  PRIMARY KEY (`id`),
  UNIQUE KEY (`chain_id`, `address`, `block_number`, `tx_index`, `log_index`),
  KEY `idx_chainId_addr_bt` (`chain_id`, `address`, `block_timestamp`), -- for targeting contract
  KEY `idx_chainId_t0_bt` (`chain_id`, `topic_0`, `block_timestamp`), -- for targeting event signature
  KEY `idx_chainId_t0_t1_bt` (`chain_id`, `topic_0`, `topic_1`, `block_timestamp`), -- for targeting event signature and indexed parameter 1
  KEY `idx_chainId_t0_t2_bt` (`chain_id`, `topic_0`, `topic_2`, `block_timestamp`), -- for targeting event signature and indexed parameter 2
  KEY `idx_chainId_txHash` (`chain_id`, `tx_hash`)  -- for targeting tx hash
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='event log';

-- block syncranization status
CREATE TABLE `event_db`.`block_sync` (
  `chain_id` bigint unsigned NOT NULL COMMENT 'chain id',
  `address` varchar(128) NOT NULL COMMENT 'contract address',
  `last_sync_number` bigint unsigned NOT NULL COMMENT 'last synced block number',
  `last_sync_hash` varchar(128) NOT NULL COMMENT 'last synced block hash',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated at',
  PRIMARY KEY (`chain_id`, `address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='block syncranization status';