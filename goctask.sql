/*
 Navicat Premium Data Transfer

 Source Server         : docker-centos7-master
 Source Server Type    : MySQL
 Source Server Version : 50730
 Source Host           : 192.168.56.10:3306
 Source Schema         : goctask

 Target Server Type    : MySQL
 Target Server Version : 50730
 File Encoding         : 65001

 Date: 27/05/2020 12:37:43
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for task
-- ----------------------------
DROP TABLE IF EXISTS `task`;
CREATE TABLE `task`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `t_title` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '任务名称',
  `t_crontab` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '1' COMMENT 'Crontab字符串',
  `t_content` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '任务内容',
  `t_start_time` datetime(0) NULL DEFAULT NULL COMMENT '开始时间',
  `t_end_time` datetime(0) NULL DEFAULT NULL COMMENT '结束时间',
  `t_status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态：1 正常， 2禁用',
  `t_run_status` tinyint(1) NOT NULL DEFAULT 9 COMMENT '运行状态：1 等待运行， 2 运行中,  9 未运行',
  `created_at` datetime(0) NULL DEFAULT NULL,
  `updated_at` datetime(0) NULL DEFAULT NULL,
  `c_id` int(11) NULL DEFAULT 0 COMMENT '负责人id，通知报警用',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of task
-- ----------------------------
INSERT INTO `task` VALUES (3, '测试', '*/1 * * * * ?', 'php test.php', '2020-05-20 16:48:22', '2020-06-06 16:48:25', 1, 9, '2020-05-22 16:48:31', '2020-05-22 16:48:34', 0);

SET FOREIGN_KEY_CHECKS = 1;
