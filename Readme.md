一个简单的运维调试工具，此工具旨在帮助运维定位真实ip，在一些复杂的场景中，一个域名要经过很多层转发才会到达后端服务，这个过程中ip地址因为转发会出现后端服务获取到的是内网ip的情况。  

同时支持自定义返回状态码，通过`?http_code=500`指定返回状态码，模拟后端服务异常情况下不同网关对于错误状态的处理和`Header`获取情况。

浏览器直接访问ip:port/?format=json，会返回如下json
```json
{
    "ClientIp": "::1",
    "IpAddress": "局域网地址",
    "RealIp": "::1"
}
```
- Client-Ip: 程序获取到的IP
- RealIp: 通过X-Forwarded-For X-Original-Forwarded-For ，由程序获取到真实IP
- IpAddress: GeoLite2-City.mmdb 获取到的ip地址


#### 返回Text信息
```
http://127.0.0.1:8080/
```

#### 指定返回状态码，测试不同状态码下网关处理逻辑
```shell
curl "http://127.0.0.1:8080?http_code=500" -I
curl "http://127.0.0.1:8080?http_code=400" -I
```
- http_code 任意http状态码值

#### 启动
```shell
./app -port :8081
```

#### docker运行
```shell
docker run -itd -p 8087:8080 --name getip  humorwang/getip:latest
```
realip库参考: https://github.com/tomasen/realip

更新数据： https://github.com/P3TERX/GeoLite.mmdb/releases

