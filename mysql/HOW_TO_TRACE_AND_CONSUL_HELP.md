### 背景

`线上遇到MySQL bug， 能被感知到的，通常都是极为严重的，mysql5.7 的bug 非常多，本人在测试环境遇到一个crash bug，不停crash； 一个死锁bug，导致线上数据库无法再建连； 本文目的是为了总结遇到这类问题时，该如何处理`

### 【问题1】当你遇到Crash的时候，该怎么办呢？

- `当务之急是尽快恢复服务`
- 重启后服务能恢复的话，先抗住，分析原因，隔离服务
- 重启后服务仍频繁crash，该怎么办呢？ ( 如下是开放性问题 )
  + 找出问题服务，先隔离问题服务？
  + 如何隔离问题服务？
  + MySQL 配置中，是否会导致丢事务？ innodb_flush_log_at_trx_commit && sync_binlog，业务是否接受，是否需要恢复

#### 技术上分析出问题服务

- 分析error-log 中的堆栈信息，![crash log](https://ws2.sinaimg.cn/large/006tNc79gy1g2b49pnc5hj30pk0nnafy.jpg)
- 分析 binary log，找到对应的 thread_id --> ，即可找到连接的库名,![binlog details](https://ws1.sinaimg.cn/large/006tNc79gy1g2b928xfwuj30tp0g4goy.jpg)


#### 服务隔离

- 根据连接用户，可以先回收用户权限
- 可以停掉对应的服务


#### 根据crash 的堆栈信息，分析问题并查询最近的版本中，是否已经有人已经遇到咯

- 堆栈信息推测是因为插入的语句导致的crash，
- 从general_log 或者 业务日志中 或者tcpdump 中拿到 insert 的sql，如下 resource_data 为text 字段类型

``` sql
INSERT INTO t_meta_resource (uid, channel_name, status, resource_data, resource_type, resource_union_id)
        VALUES
        ( ...... )
        ON DUPLICATE KEY UPDATE
        resource_data = VALUES(resource_data),
        is_deleted = 0
```

- 查看 [mysql release notes](https://dev.mysql.com/doc/relnotes/mysql/5.7/en/news-5-7-21.html) 发现 Bug #26734162 已经在5.7.21 分支中被解决了； 当前使用的分支是5.7.19
- Good News 这个bug 已经被解决了
- 升级MySQL 版本，你们升级版本前会做什么工作呢? 【open 问题， 压测？ 灰发？ 直接上线】,
- MySQL 版本是自己编译的？ 还是官网上下的二进制呢？

### 【问题2】遇到MySQL 死锁

`当触发这个死锁bug 时，数据库无法再建立, MySQL 服务端在鉴权之前就死了，这一现象，未遇到的童鞋很难有真实体会，我写了一段复现问题的代码，在5.7.22~5.7.24之间 可以复现这个问题，因为这个是5.7.22 引入的bug`

#### 问题复现

- 参考 [deadlock reproduce](https://github.com/jianhaiqing/debug-problem/tree/master/src/deadlock_reproduce) 来复现死锁， mysql版本: 5.7.22 ~ 5.7.24; 

#### 问题分析

- 当这个死锁问题出现时，我们该怎么办呢？
- 假如业务容忍，保留现场后，如何分析这个问题呢？
- 本人的分析过程，参考 [deadlock](https://dba.stackexchange.com/questions/234769/mysql-failed-to-handshake-due-to-lock-thread-cache-not-released)

#### 工具介绍

- pstack & gdb, 用gdb 将 mysqld 的堆栈信息打印出来 参考 [pstack-strace](https://github.com/bangerlee/strace_pstack/blob/master/pstack.sh)
- pt-pmp 打印 堆栈信息


#### 关于社区寻求帮助

- 可以去这里咨询一下[dba stackexchange](https://dba.stackexchange.com/questions/234769/mysql-failed-to-handshake-due-to-lock-thread-cache-not-released)
- 本人用的是percona 版本，在percona 社区得到了及时的响应[percona forums](https://www.percona.com/forums/questions-discussions/mysql-and-percona-server/percona-server-5-7/53900-mysql-failed-to-handshake-due-to-lock_thread_cache-not-released)
- MySQL QQ 技术群793818397 得到叶金荣老师的指导，拓宽了问题定位的手段
