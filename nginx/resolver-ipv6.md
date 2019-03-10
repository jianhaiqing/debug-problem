### 问题描述

> 由于某些场景，我们需要做代理转发服务（比如，服务器无法访问外网；如被调用的服务的IP经常变化，在代码中直接访问域名可能会有DNS 缓存，导致服务IP变化时，导致DNS未更新而，访问失败）；
> 在生产环境中，我们的服务器无法访问外网，所以服务端配置了host 到能否访问外网的nginx服务器上，做代理转发； 如下为nginx配置

```conf
server {
        listen  80;
        server_name  jianhaiqing.test.seewo.com;
        resolver 114.114.114.114 valid=30s;
	set $backends "care.test.seewo.com";
        location / {
                proxy_pass http://$backends;
                proxy_http_version 1.1;
                proxy_set_header   Host "care.test.seewo.com";
        }
}
```

> 问题： 某天业务反馈，服务端调用此服务器，一直502，是不是线上网路有问题；

### 问题分析

> 从kibana（由于有多台nginx，为了便于查看和分析日志，我们将nginx 日志收集到了kibana 上）上看到，nginx 并没有找到出去，error 日志看是 care.test.seewo.com 的域名解析失败导致的；
> dig care.test.seewo.com @114.114.114.114 +short 能否正常解析；（一层的CNAME 解析）
``` bash
$ dig care.test.seewo.com @114.114.114.114 +short
gtm.test.seewo.com.
121.8.148.91
```
> 将服务端抓包，发现 nginx 向114.114.114.114 上查看 care.test.seewo.com 的ipv6 地址失败导致的；
![AAAA ipv6 解析](https://ws2.sinaimg.cn/large/006tKfTcgy1g0wl7z98l6j31h609r7b6.jpg)
> dig -6 care.test.seewo.com @114.114.114.114, DNS 服务器正常返回
``` bash
$ dig -6 care.test.seewo.com @114.114.114.114  +short
gtm.test.seewo.com.
121.8.148.91
```
> 使用公司的DNS 服务器解析 `resolver 172.17.82.12 valid=30s`; 公司对这个域名做了内外网分析解析，
``` bash
$ dig -6 care.test.seewo.com @172.17.82.12  +short
gtm.test.seewo.com.
10.10.14.60
```
![inner-dns ipv6](https://ws4.sinaimg.cn/large/006tKfTcgy1g0wqfmzobuj32y70o4x6p.jpg)

> 将care.test.seewo.com 改为A 记录，可以正常代理转发；

### 总结

+ 使用公网DNS 服务器114.114.114.114 ，域名是CNAME时, nginx ipv6 解析失败（估计是返回的包不一样，）
+ 使用公网DNS 服务器114.114.114.114 ，域名是CNAME时，resolver 114.114.114.114 valid=30s ipv6=off;关闭ipv6 解析的情况下，代理能正常转发
+ 使用公网DNS 服务器114.114.114.114 ，域名是A 时,代理能正常转发


### 备注

+ nginx版本 如下
``` bash
$ nginx -V
nginx version: openresty/1.13.6.1
built by gcc 4.8.5 20150623 (Red Hat 4.8.5-16) (GCC)
built with OpenSSL 1.0.2k-fips  26 Jan 2017
TLS SNI support enabled
configure arguments: --prefix=/usr/local/openresty/nginx --with-cc-opt='-O2 -O2 -g -pipe -Wp,-D_FORTIFY_SOURCE=2 -fexceptions -fstack-protector --param=ssp-buffer-size=4 -m64 -mtune=generic' --add-module=../ngx_devel_kit-0.3.0 --add-module=../echo-nginx-module-0.61 --add-module=../xss-nginx-module-0.05 --add-module=../ngx_coolkit-0.2rc3 --add-module=../set-misc-nginx-module-0.31 --add-module=../form-input-nginx-module-0.12 --add-module=../encrypted-session-nginx-module-0.07 --add-module=../srcache-nginx-module-0.31 --add-module=../ngx_lua-0.10.11 --add-module=../ngx_lua_upstream-0.07 --add-module=../headers-more-nginx-module-0.33 --add-module=../array-var-nginx-module-0.05 --add-module=../memc-nginx-module-0.18 --add-module=../redis2-nginx-module-0.14 --add-module=../redis-nginx-module-0.3.7 --add-module=../rds-json-nginx-module-0.15 --add-module=../rds-csv-nginx-module-0.08 --add-module=../ngx_stream_lua-0.0.3 --with-ld-opt=-Wl,-rpath,/usr/local/openresty/luajit/lib --conf-path=/usr/local/nginx/nginx.conf --error-log-path=/usr/local/nginx/log/error.log --http-log-path=/usr/local/nginx/log/access.log --pid-path=/var/run/nginx.pid --lock-path=/var/run/nginx.lock --http-client-body-temp-path=/var/cache/nginx/client_temp --http-proxy-temp-path=/var/cache/nginx/proxy_temp --http-fastcgi-temp-path=/var/cache/nginx/fastcgi_temp --http-uwsgi-temp-path=/var/cache/nginx/uwsgi_temp --http-scgi-temp-path=/var/cache/nginx/scgi_temp --user=nginx --group=nginx --with-http_ssl_module --with-http_realip_module --with-http_addition_module --with-http_sub_module --with-http_dav_module --with-http_flv_module --with-http_mp4_module --with-http_gunzip_module --with-http_gzip_static_module --with-http_random_index_module --with-http_secure_link_module --with-http_stub_status_module --with-http_auth_request_module --with-mail --with-stream --with-mail_ssl_module --with-file-aio --with-ipv6 --add-module=/usr/local/src/nginx-upsync/nginx-upsync-module --add-module=/usr/local/src/ngx_cache_purge-2.3 --add-module=/usr/local/src/nginx-goodies-nginx-sticky-module-ng-08a395c66e42 --with-http_v2_module --with-stream_ssl_module --with-stream --with-stream_ssl_module
```

+ 测试脚本
```
curl -v jianhaiqing.test.seewo.com --resolve jianhaiqing.test.seewo.com:80:10.21.17.36
tcpdump  -i any  'port 80 or port 53' -w resolver-ipv6.pcap
```
+ 详细的tcpdump 包信息, refer to https://github.com/jianhaiqing/debug-problem/blob/master/nginx/resolver-ipv6.pcap

