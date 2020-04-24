
CREATE TABLE `fudao_count_hist` (
  `id` bigint(10) unsigned NOT NULL AUTO_INCREMENT,
  `date_time` date NOT NULL DEFAULT '2000-01-01' COMMENT '用户id',
  `subject` int(32) NOT NULL DEFAULT '0' COMMENT '科目',
  `sys_count` int(11) DEFAULT '0' COMMENT '系统课数量',
  `course_count` int(11) DEFAULT '0' COMMENT '专题课数量',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE `dateSubject` (`date_time`,`subject`)
) ENGINE=InnoDB  COMMENT='数量信息';




CREATE TABLE `fudao_course_history` (
  `id` bigint(10) unsigned NOT NULL AUTO_INCREMENT,
  `date_time` date NOT NULL DEFAULT '2000-01-01' COMMENT '每日详情',
  `course_id` int(32) NOT NULL DEFAULT '0' COMMENT '课程id',
  `subject` int(32) NOT NULL DEFAULT '0' COMMENT '科目',
  `grade` int(11)  NOT NULL  DEFAULT '0' COMMENT '年级',
  `title` varchar(1024)  NOT NULL  DEFAULT '0' COMMENT '课程名称',
  `teacher` varchar(1024)  NOT NULL  DEFAULT '0' COMMENT '课程名称',
  `price` decimal(10,2)  NOT NULL  DEFAULT '0' COMMENT '价格',
  `detail`  json DEFAULT NULL COMMENT '课程详情原生数据',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE `dateSubject` (`date_time`,`course_id`)
) ENGINE=InnoDB  COMMENT='学科详情信息';