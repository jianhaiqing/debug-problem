### 公司内网安装脚本

- 安装以及查看nginx 编译的模块
```bash
curl http://swdl.gz.cvte.cn/server_deploy/install-openresty.sh | bash
nginx -V 
```

### 线上 NGINX 拓扑

- http nginx
- nfs nginx
- tcp nginx (独立）

- ![http & nfs nginx 拓扑](https://ws4.sinaimg.cn/large/006tNc79gy1g2xd4g2zaqj30h10bcmxm.jpg)

### 常见nginx 配置方法

#### 0. 普通配置 & http ssl 配置

- 只支持https
```
server {
        listen       443 ssl;
        server_name  myou.cvte.com;

        ssl_certificate /etc/nginx/https/cvte.com.crt;
        ssl_certificate_key /etc/nginx/https/cvte.com.key;
        location / {
          proxy_pass http://xx;
        }
}
```

- 同时支持https和http
```
server {
        listen       80;
        listen       443 ssl;
        server_name  myou.cvte.com;

        ssl_certificate /etc/nginx/https/cvte.com.crt;
        ssl_certificate_key /etc/nginx/https/cvte.com.key;
        location / {
          proxy_pass http://xx;
        }
}
```
- 强制http跳转到https
所有http的请求都跳转到https同样的URL
```
server {
        listen       80;
        listen       443 ssl;
        ssl_certificate /etc/nginx/https/seewo.com.crt;
        ssl_certificate_key /etc/nginx/https/seewo.com.key;
        server_name  www.seewo.com;
       
        if ($scheme = http ) {
                rewrite "^(.*)" https://$host$1 permanent;
                break;
        }
}
#  注意是所有请求，其次 是permanent
```
- http2 配置

```
server {
        listen       80;
        listen       443 ssl http2;
        ssl_certificate /etc/nginx/https/seewo.com.crt;
        ssl_certificate_key /etc/nginx/https/seewo.com.key;
        server_name  www.seewo.com;
}
# 一处配置，全局生效
```

#### 1. 转发配置

- /a/b/c 转发到后端的/a/b/c
```
upstream  seewo-common-service{
        server 10.168.58.189:10090;
        keepalive 16;
}
location / {
        proxy_pass http://seewo-common-service;
        proxy_redirect      off;
        proxy_set_header        X-Forwarded-Url "$scheme://$host$request_uri";
        proxy_http_version  1.1;
        proxy_set_header    Connection "";
        proxy_set_header    Cookie $http_cookie;
        proxy_set_header    Host $host;
        proxy_set_header    X-Real-IP $remote_addr;
        proxy_set_header    X-Forwarded-For $proxy_add_x_forwarded_for;
        client_max_body_size            512M;
        client_body_buffer_size         64M;
        proxy_connect_timeout           900;
        proxy_send_timeout              900;
        proxy_read_timeout              900;
        proxy_buffer_size               64M;
        proxy_buffers                   64 16M;
        proxy_busy_buffers_size         256M;
        proxy_temp_file_write_size      512M;
}
```
- /a/b/c 转发到后端的/a/api/b/c
```
upstream  seewo-common-service{
        server 10.168.58.189:10090;
        keepalive 16;
}
location /a/ {
        proxy_pass http://seewo-common-service/a/api/;
        proxy_redirect      off;
        proxy_set_header        X-Forwarded-Url "$scheme://$host$request_uri";
        proxy_http_version  1.1;
        proxy_set_header    Connection "";
        proxy_set_header    Cookie $http_cookie;
        proxy_set_header    Host $host;
        proxy_set_header    X-Real-IP $remote_addr;
        proxy_set_header    X-Forwarded-For $proxy_add_x_forwarded_for;
        client_max_body_size            512M;
        client_body_buffer_size         64M;
        proxy_connect_timeout           900;
        proxy_send_timeout              900;
        proxy_read_timeout              900;
        proxy_buffer_size               64M;
        proxy_buffers                   64 16M;
        proxy_busy_buffers_size         256M;
        proxy_temp_file_write_size      512M;
}
```

- /a 转发到后端的/b
```
upstream  seewo-common-service{
        server 10.168.58.189:10090;
        keepalive 16;
}
server {
        listen       80;
        server_name  ce.cvte.com;
        location /a {
          proxy_pass http://seewo-common-service/b;
          proxy_redirect      off;
          proxy_set_header        X-Forwarded-Url "$scheme://$host$request_uri";
          proxy_http_version  1.1;
          proxy_set_header    Connection "";
          proxy_set_header    Cookie $http_cookie;
          proxy_set_header    Host $host;
          proxy_set_header    X-Real-IP $remote_addr;
          proxy_set_header    X-Forwarded-For $proxy_add_x_forwarded_for;
          client_max_body_size            512M;
          client_body_buffer_size         64M;
          proxy_connect_timeout           900;
          proxy_send_timeout              900;
          proxy_read_timeout              900;
          proxy_buffer_size               64M;
          proxy_buffers                   64 16M;
          proxy_busy_buffers_size         256M;
          proxy_temp_file_write_size      512M;
    }
}
```

- websocket 代理
```
upstream ce-node {
        server 10.168.58.189:10090;
        keepalive 16;
}
map $http_upgrade $connection_upgrade {
        default upgrade;
        ''      close;
}
server {
        listen       80;
        server_name  ce.cvte.com;
        location /admin {
                proxy_pass http://ce-node/admin;
                proxy_redirect      off;
                proxy_set_header                X-Forwarded-Url         "$scheme://$host$request_uri";
                proxy_set_header    Upgrade $http_upgrade;
                proxy_set_header    Connection "upgrade";
                proxy_http_version  1.1;
                proxy_set_header    Cookie $http_cookie;
                proxy_set_header    Host $host:$server_port;
                proxy_set_header    X-Real-IP $remote_addr;
                proxy_set_header    X-Forwarded-For $proxy_add_x_forwarded_for;
                client_max_body_size        64m;
                client_body_buffer_size     128k;
      
        }
}
```

#### 2. 重定向

- 302
```
# 可以根据情况放到不同context 下，server, http 或者 location
if ($scheme = http ) {
  rewrite "^(.*)" https://$host$1 redirect;
  break;
}
# 等价配置
location / {
  return 302      https://$host$request_uri;
}
```
- 301
```
# 可以根据情况放到不同context 下，server, http 或者 location        
if ($scheme = http ) {
  rewrite "^(.*)" https://$host$1 permanent;
  break;
}
# 等价配置
location / {
  return 301  https://$host$request_uri;
}
```
- 302 vs 301

```
http.status_code 为301,则浏览器端会缓存这个结果，但浏览器再次发起请求时，不会再向服务端请求原来的URL，容易导致失控；
优点就是变成了一次请求
```

- A域名跳转到B域名
```
if ($host != 'www.xifo.com' ) {
    rewrite ^/(.*)$ https://www.xifo.com/$1 permanent;
    break;
}
# 两种配置，根据不同需求配置
if ($host == 'xifo.com' ) {
    rewrite ^/(.*)$ https://www.xifo.com/$1 permanent;
    break;
}
```

#### 3. 动态DNS 配置

- 有些场景是a.cn 域名转发到 b.cn 下，b.cn 的域名解析可能会变化，假如直接配置proxy_pass http://b.cn, 则nginx 将缓存第一次的DNS 解析结果，DNS 变更时，无法感知；如下是设置为动态解析

1. https://$backends;  backends 后面不能有url
2. host header 要手动设置 "b.cn"
3. 把ipv6 关闭，当前nginx 在dns 解析时，ipv6 打开的情况下，b.cn 有CNAME 解析时，DNS会解析成功,ipv6-AAAA 解析失败，nginx 会认为DNS 解析失败，而导致502；
```
location  /after {
                resolver  100.100.2.136 valid=30s ipv6=off;
                set $backends "b.cn";
                proxy_pass https://$backends;
                proxy_redirect                  off;
                proxy_http_version              1.1;
                proxy_set_header                Connection "";
                proxy_set_header                Cookie $http_cookie;
                proxy_set_header                Host "b.cn";
                proxy_set_header                X-Real-IP $remote_addr;
                proxy_set_header                X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Proto $scheme;
                client_max_body_size            100m;
                client_body_buffer_size         128k;
                proxy_connect_timeout           1800;
                proxy_send_timeout              180;
                proxy_read_timeout              180;
                proxy_buffer_size               16k;
                proxy_buffers                   16      64k;
                proxy_busy_buffers_size         128k;
                proxy_temp_file_write_size      128k;
        }


```
#### 4. gzip 压缩配置

- gzip压缩, 根据域名转发需求做压缩
```
        gzip on;
        gzip_min_length 1024;   
        gzip_buffers 4 16k;
        gzip_comp_level 3;    
        gzip_types application/x-javascript application/javascript text/css;
        gzip_vary on;
```

#### 5. 跨域配置
- 允许某些方法或者某些origin的请求跨域, 允许所有域名则 Access-Control-Allow-Origin *
```
if ($request_method = OPTIONS) {
          add_header Access-Control-Allow-Origin "*";
          add_header Access-Control-Allow-Methods "GET,OPTIONS";
          add_header Content-Length 0;
          add_header Content-Type text/plain;
          add_header Access-Control-Allow-Headers "*";
          return 200;
  }
  add_header Access-Control-Allow-Origin "*";
  add_header Access-Control-Allow-Methods "GET,OPTIONS";
  add_header Access-Control-Allow-Credentials "true";
  add_header Access-Control-Allow-Headers "*";
```

#### 6. 访问限制allow 配置

- 只允许某些IP访问
```
cat << EOF >> inner-outer-limits.ip
  allow 10.0.0.0/8;
  allow 183.63.127.82 ;
  allow 121.8.170.250 ;
  allow 121.8.148.66;
  allow 221.4.55.68;
  allow 39.155.212.2;
  deny all;
EOF
# 在nginx 相应context中引入这个文件 (http, server, location)
include /usr/local/nginx/conf.d/inner-outer-limits.ip;
```
- 只禁止某些IP访问，一般用来拒绝来自某些ip的攻击
```
  deny 10.0.0.0/8;
  deny 183.63.127.82 ;
  allow all;
```

- Basic Auth限制访问
```
  location / {
    auth_basic "auth";
    auth_basic_user_file /etc/nginx/htpasswd/auth_basic.htpasswd;
    proxy_pass http://rabbitmq/;
  }
```
- auth_basic.htpasswd 生成
``` bash
yum install httpd-tools -y
sudo htpasswd  -c auth_basic.htpasswd psd
# 输入预设值的密码，用户是psd
auth_basic           "auth Area";
auth_basic_user_file /etc/nginx/htpasswd/auth_basic.htpasswd;
# Context:	http, server, location, limit_except
其中/etc/nginx/htpasswd/auth_basic.htpasswd 文件存有用户名和一段hash后的密码，可以用工具生成
http://tool.oschina.net/htpasswd
```

#### 7. header 来进行分流

- 通过自定义header x-app-type 来控制版本转发

```
# 说明 easinote-web 和 easinote-web-newversion 是两个不同版本的upstream
set $xapptype $http_x_app_type;
set $xen_upstreams easinote-web;
if ( $xapptype = "newApp" ) {
        set $xen_upstreams  easinote-web-newversion;

}
location / {
        proxy_pass http://$xen_upstreams;
        proxy_redirect      off;
        proxy_set_header                X-Forwarded-Url         "$scheme://$host$request_uri";
        proxy_http_version  1.1;
        proxy_set_header    Connection "";
        proxy_set_header    Cookie $http_cookie;
        proxy_set_header    Host $host;
        proxy_set_header    X-Real-IP $remote_addr;
        proxy_set_header    X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        client_max_body_size            512M;
        client_body_buffer_size         64M;
        proxy_connect_timeout           900;
        proxy_send_timeout              900;
        proxy_read_timeout              900;
        proxy_buffer_size               64M;
        proxy_buffers                   64 16M;
        proxy_busy_buffers_size         256M;
        proxy_temp_file_write_size      512M;
}

```

#### 8. 变量配置+IF 来做转发逻辑

- `背景，node 端是通过X-Forwarded-Url 这个变量或者protocol 信息的，在nginx 二层代理中会丢掉https，因为 X-Forwarded-Url 在第二层代理中被第二次set，导致https丢失；`

```
# 参考如下做法；解决第二层nginx 既可以作为第一层，也可以作为第二层
set $xforwarded_url "$scheme://$host$request_uri";
if ( $http_x_forwarded_url != "" ) {
        set $xforwarded_url $http_x_forwarded_url;
}
proxy_set_header                X-Forwarded-Url         $xforwarded_url; 
```

#### 9. 正则匹配


- [参考这篇文章](http://seanlook.com/2015/05/17/nginx-location-rewrite/)

```
# 根据文件类型转发
location ~* \.(css)$  {
  proxy_pass http://nginx-prd-1-ccloud-ali-v1;
  expires -1;
  add_header Pragma "no-cache";
  add_header Cache-Control "public,max-age=0";
}
```

- nginx 是最长匹配
```
location /images/ {

}

location /images/abc {

}
```

#### 10. nginx 浏览目录&文件

```
server {
	listen      80;
	server_name  swdl.gz.cvte.cn;

	location / {
		root /data/WebRoot/download/;
		autoindex on;
		autoindex_exact_size off;
		autoindex_localtime on;
	}
	# 点击浏览的文件时，不会变为下载，而是浏览器直接显示
	location ~ .*\.conf$ {
		proxy_hide_header Content-Type;
		add_header Content-Type text/plain;
		root /data/WebRoot/download/;
	}
}
```

#### 11. upsync + consul 实现高可用


#### 12.限制连接和速度

```
# 限制连接,context: http
# 根据IP 和 域名限制连接数，使用10m 的共享内存
limit_conn_zone $binary_remote_addr zone=conn-perip:10m;
# 限制server_name 的连接数，
limit_conn_zone $server_name zone=conn-perserver:10m;
location / {
# context: http, server, location
limit_conn conn-perip 3;
limit_conn conn-perserver 100;
}


# 限制QPS, context: http
limit_req_zone $binary_remote_addr zone=conn-perip:10m rate=5r/s;
limit_req_zone $server_name zone=conn-perserver:10m rate=10r/s;
location / {
# context: http, server, location
limit_req zone=conn-perip burst=5 nodelay;
limit_req zone=conn-perserver burst=10;
}
limit_req zone=one burst=10;
limit_req_status 503;  # default value anyway
# 5 requests/second from 1 single IP address are allowed.
# Between 5 and 10 requests/sec all new incoming requests are delayed.
# Over 10 requests/sec all new incoming requests are rejected with the status code set in limit_req_status

----------
# 直接限制访问速度
limit_rate 2000k;
```

### 常见问题

#### a. 重定向 302 & 301
- 影响 
  + 浏览器影响
  + 服务端影响（node&java 不会追302&301）

##### b. 缓存的配置

```
add_header    Cache-Control  max-age=5184000; # 缓存2个月
```
- 在缓存时间内，浏览器将不再请求服务端， 采用缓存的静态文件css,js 等一定要hash，否则将失控；

##### c. header 中的特殊字符问题
- 1.0.1 自定义header
  1. underscores_in_headers default off. 不允许带下划线；
  2. 通常自定义头部使用x-access-token； 以x- 开头
  3. header 中带有特殊字符如x_access_token ，请求将被nginx 丢弃，不会转发到后端
- 是否会被转发

##### d. 动态DNS 配置

- resolver  100.100.2.136 valid=30s ipv6=off; # dns server, 缓存时间，关闭ipv6
- resolver ipv4 vs ipv6
  + 并不是所有DNS SERVER 都是支持ipv6 解析的，不支持将可能导致解析失败

##### e. 静态资源 NFS

- NFS server 挂掉
- nginx 作为访问本地文件系统时，hang 住，连接数很快就会耗尽

##### f. open file 限制

- nginx 可以打开的文件数:
  + nginx 启动方式,systemd 限制
  + nginx 运行用户有关,  /etc/security/limits.conf 和 /etc/security/limits.d/20-nproc.conf
  - nginx.conf: worker_rlimit_nofile 100000;

- 以下以systemd 启动为例，
```
cat /usr/lib/systemd/system/nginx.service
[Unit]
Description=nginx - high performance web server
Documentation=http://nginx.org/en/docs/
After=network.target remote-fs.target nss-lookup.target

[Service]
LimitCORE=infinity
LimitNOFILE=100000
LimitNPROC=100000
Type=forking
PIDFile=/var/run/nginx.pid
ExecStartPre=/usr/sbin/nginx -t -c /etc/nginx/nginx.conf
ExecStart=/usr/sbin/nginx -c /etc/nginx/nginx.conf
ExecReload=/bin/kill -s HUP $MAINPID
ExecStop=/bin/kill -s QUIT $MAINPID
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

- 请手动验证最后受哪个影响

##### g. nginx reload 失败

- yum install strace -y
- strace nginx -t

##### h. 状态码异常

- 413 会转发到后端吗？
- 502 服务挂了
- 404 谁返回的404， nginx？ 后端？


### 常用问题分析手段

- 转发异常
  + tcpdump
  + access_log
  + error_log
- 崩溃
  + nginx.conf: worker_rlimit_core 500m,working_directory /var/log/nginx;
  + error_log 监控崩溃 signal 
  


