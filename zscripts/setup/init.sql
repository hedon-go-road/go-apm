CREATE DATABASE IF NOT EXISTS `ordersvc`;
CREATE DATABASE IF NOT EXISTS `skusvc`;
CREATE DATABASE IF NOT EXISTS `usrsvc`;
CREATE DATABASE IF NOT EXISTS `dogalarm`;

CREATE TABLE IF NOT EXISTS `ordersvc`.`t_order` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `order_id` varchar(255) NOT NULL,
    `ctime` timestamp DEFAULT current_timestamp() NOT NULL,
    `utime` timestamp DEFAULT current_timestamp() ON UPDATE current_timestamp() NOT NULL,
    `sku_id` bigint(20) NOT NULL,
    `num` bigint(20) NOT NULL,
    `price` bigint(20) NOT NULL,
    `uid` bigint(20) NOT NULL,
    CONSTRAINT `order_pk2` UNIQUE (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE IF NOT EXISTS `skusvc`.`t_sku` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `name` varchar(10) NOT NULL,
    `price` bigint(20) NOT NULL,
    `num` bigint(20) NOT NULL,
    `ctime` timestamp DEFAULT current_timestamp() NOT NULL,
    `utime` timestamp DEFAULT current_timestamp() ON UPDATE current_timestamp() NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE IF NOT EXISTS `usrsvc`.`t_user` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `name` varchar(255) NOT NULL,
    `ctime` timestamp DEFAULT current_timestamp() NOT NULL,
    `utime` timestamp DEFAULT current_timestamp() ON UPDATE current_timestamp() NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `usrsvc`.`t_user` (`id`, `name`) VALUES (1,'user1') ON DUPLICATE KEY UPDATE `id` = 1;
INSERT INTO `skusvc`.`t_sku` (`id`, `name`, `price`, `num`) VALUES (3, 'sku2', 200, 200) ON DUPLICATE KEY UPDATE `id` = 3;

CREATE TABLE IF NOT EXISTS `dogalarm`.`t_deploy_info` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `app` varchar(64) NOT NULL,
    `hosts` varchar(1024) NOT NULL COMMENT 'deployment hosts, split by comma',
    `port` int(11) NOT NULL COMMENT 'service port',
    `liveprobe` varchar(255) NOT NULL COMMENT 'live probe url',
    `phone_webhook` varchar(255) NOT NULL COMMENT 'phone webhook url',
    `dingding_webhook` varchar(255) NOT NULL COMMENT 'dingding webhook url'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
