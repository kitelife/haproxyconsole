CREATE DATABASE haproxyconsole;
USE haproxyconsole;
DROP TABLE IF EXISTS `haproxymapinfo`;
CREATE TABLE `haproxymapinfo` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `servers` varchar(1024) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  `vport` int(10) NOT NULL,
  `comment` varchar(1024) DEFAULT '',
  `logornot` int(1) DEFAULT '1',
  `datetime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;