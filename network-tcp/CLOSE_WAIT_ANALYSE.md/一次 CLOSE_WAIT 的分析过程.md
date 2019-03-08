### 问题描述

> 业务反馈鲸云访问请求进不去，咨询是不是鲸云网络有问题，吓得我赶紧看一下那台flannel 是不是扑街了； flannel  正常后，开始和业务沟通，获取测试接口 "http://myou-ip-dev.test.seewo.com/mengyou-ip/ip/v2/rest?ip=171.115.155.145",
> 这个请求hung 住了，是不是网络有问题

### 分析过程

> 1. 从内网telnet 容器IP 容器PORT ，端口是通的
> 2. curl 接口，hung 住
> 3. 登录容器终端，netstat -nalp | grep CLOSE_WAIT , 发现有大量的 连接处于 CLOSE_WAIT状态，反馈给业务； --- 没有分析去结果，302 回来 ![CLOSE_WAIT 截图](https://ws2.sinaimg.cn/large/006tKfTcgy1g0vmkfggqdj30tf0njtnr.jpg)
> 4. 硬着头皮开始分析，要了代码，以及应用被调的几个接口
> 5. 因为应用是部署在 容器中的tomcat 里的，所以偷了个懒，通过分析 tomcat 的localhost_access_log.2019-03-08.txt 的api 接口来过滤异常请求
> 6. 因为 CLOSE_WAIT的连接数在200+ 左右，通过一段shell grep -v ，推测出 /ip/v2/batch/phone
> 7. 发现此接口如下代码片段有问题，数据库连接未关闭；

```java
SqlSession sqlSession = MyBatisHelper.getSqlSessionFactory().openSession();
IpplusPhoneInfoDao ipplusPhoneInfoDao = sqlSession.getMapper(IpplusPhoneInfoDao.class);
```
> 8. 那数据库连接未关闭，为什么会导致那不多 CLOSE_WAIT的连接，而且没有数据库的连接；
> 9. 一波 seq 1 1000 | xargs -I {} curl /mengyou-ip/ip/v2/batch/phone 单线程压测数据库，发现几个请求过后，curl 就hung 住了；连接池耗尽了，所以hung住？
> 10. 查看了代码的连接池配置，如下

```java
@Override
    public DataSource getDataSource() {
        DruidDataSource druidDataSource = new DruidDataSource();
        druidDataSource.setDriverClassName(this.properties.getProperty("driver"));
        druidDataSource.setUrl(this.properties.getProperty("url"));
        druidDataSource.setUsername(this.properties.getProperty("username"));
        druidDataSource.setPassword(this.properties.getProperty("password"));
        druidDataSource.setInitialSize(5);
        druidDataSource.setMinIdle(5);
        druidDataSource.setMaxActive(30);
        try {
            druidDataSource.init();
        } catch (SQLException e) {
            e.printStackTrace();
        }
        return druidDataSource;
    }
```
> 11. 为什么获取连接会hung 呢？ 看了一下DruidDataSource 的默认设置 

```java
protected volatile long                            maxWait                                   = DEFAULT_MAX_WAIT;
public final static int                            DEFAULT_MAX_WAIT                          = -1;
```

> 12. 在没有设置拿连接超时的情况下，会一直等可用连接，所有curl 请求就hung 住了； 
> 13. 此处引用开发的分析，是对的：`CLOSE_WAIT一个原因是自身处理请求慢，对方设置了读超时，所以关闭了请求, APM 不会收集到没有返回的请求`,

### 结论及解决

> 结论：数据库连接没有正常关闭，导致数据库连接池被耗尽，凡是需要查询DB 的接口，都会hung 住，调用方okhttp 请求设置了超时关闭，业务端由于hung 住，没能正常关闭这个http连接，连接状态进入 CLOSE_WAIT（客户端已经关闭），导致http 连接池逐步堆积，最终耗尽了tomcat 的http 连接池； ==> 健康检查失败，服务重启; 如下是客户端超时关闭后，服务端CLOSE_WAIT产生过程的包；
 ![okhttp-timeout](https://ws2.sinaimg.cn/large/006tKfTcgy1g0vnoiaazpj31h9053mzv.jpg)

> 解决：

+ 正常关闭sql 连接
```java
SqlSession sqlSession = MyBatisHelper.getSqlSessionFactory().openSession();
IpplusPhoneInfoDao ipplusPhoneInfoDao = sqlSession.getMapper(IpplusPhoneInfoDao.class);
sqlSession.close(); // 正常关闭连接
```
+ 数据库连接池设置： 获取连接超时设置

```java
druidDataSource.setMaxWait(1000); // 1000 ms
```

### 参考TCP 四次挥手过程,附上测试脚本

``` bash
# 压测耗尽DB连接，复现问题
seq 1 1000 | xargs -I {} curl -X POST \
  http://172.30.116.61:8080/mengyou-ip/ip/v2/batch/phone \
  -H 'Content-Type: application/json' \
  -H 'Postman-Token: 6f9c17da-357a-4f59-bde6-0285058b6e9c' \
  -H 'cache-control: no-cache' \
  -d '["13302503777", "18302890101"]'
  
# 超时复现CLOSE_WAIT
curl  --max-time 3 -X POST \
  http://172.30.130.102:8080/mengyou-ip/ip/v2/batch/phone \
  -H 'Content-Type: application/json' \
  -H 'Postman-Token: 6f9c17da-357a-4f59-bde6-0285058b6e9c' \
  -H 'cache-control: no-cache' \
  -d '["13302503777", "18302890101"]'
```