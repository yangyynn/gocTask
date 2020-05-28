/*
 Navicat Premium Data Transfer

 Source Server         : wsl_localhost
 Source Server Type    : MySQL
 Source Server Version : 50730
 Source Host           : 127.0.0.1:3306
 Source Schema         : goctask

 Target Server Type    : MySQL
 Target Server Version : 50730
 File Encoding         : 65001

 Date: 28/05/2020 16:05:21
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for goc_notify
-- ----------------------------
DROP TABLE IF EXISTS `goc_notify`;
CREATE TABLE `goc_notify`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `t_id` int(11) NOT NULL,
  `created_at` datetime(0) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for goc_task
-- ----------------------------
DROP TABLE IF EXISTS `goc_task`;
CREATE TABLE `goc_task`  (
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
  `notify_num` tinyint(10) NULL DEFAULT 2 COMMENT '每天最大报警次数',
  `notify_id` varchar(11) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '0' COMMENT '负责人微信企业号id',
  `notify_email` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '负责人email',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for goc_task_log_1
-- ----------------------------
DROP TABLE IF EXISTS `goc_task_log_1`;
CREATE TABLE `goc_task_log_1`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `t_id` int(11) NOT NULL,
  `l_status` int(5) NOT NULL COMMENT '执行结果code',
  `l_result` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '执行结果',
  `l_use_time` float NOT NULL COMMENT '程序消耗时间,单位秒',
  `created_at` int(10) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `t_id`(`t_id`) USING BTREE,
  INDEX `created_at`(`created_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for goc_task_log_2
-- ----------------------------
DROP TABLE IF EXISTS `goc_task_log_2`;
CREATE TABLE `goc_task_log_2`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `t_id` int(11) NOT NULL,
  `l_status` int(5) NOT NULL COMMENT '执行结果code',
  `l_result` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '执行结果',
  `l_use_time` float NOT NULL COMMENT '程序消耗时间,单位秒',
  `created_at` int(10) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `t_id`(`t_id`) USING BTREE,
  INDEX `created_at`(`created_at`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
