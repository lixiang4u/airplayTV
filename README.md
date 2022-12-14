# airplayTV

## 项目说明
- 该项目是研究web安全衍生物
- 通过一系列技术获取视频网站的播放列表以及播放地址等数据
- 涉及：js加解密/数据加解密/使用websocket加解密数据/CORS跨域/COOKIE认证/人机认证等
- 请勿使用该项目用于任何非法用途！！！

## 配置前端代码
- 拉取前端代码`git clone git@github.com:lixiang4u/airplayTV-web.git`
- 找到文件`src/main.js` 替换 `axios.defaults.baseURL`后的服务器地址

## 前端构建
```code
npm install
npm run build
```

## 配置nginx
- 在`server`模块配置web根目录指向 `dist/` 目录
```code
root '/path/to/airplayTV-web/dist';
```

- 在`http`模块配置`websocket upgrade`和`upstream`

```code
#添加反向代理websocket相关map
map $http_upgrade $connection_upgrade {
    default upgrade;
    '' close;
}

# 定义反向的上游
upstream airplayTV {
    server 127.0.0.1:8089;
}
```

- 在`server`模块配置`location`的`接口、websocket、m3u8资源`请求转发

```code

# 添加默认请求文件
location / {
    try_files $uri $uri/ /index.html last;
}
# 代理/api/到后端
location /api/ {
    proxy_pass http://airplayTV;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header REMOTE-HOST $remote_addr;

    #Set Nginx Cache
    add_header Cache-Control no-cache;
    expires 12h;
}
# 代理websocket到后端
location /api/ws {
    proxy_pass http://airplayTV;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header REMOTE-HOST $remote_addr;

    # 升级websocket
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection $connection_upgrade;

    #Set Nginx Cache
    add_header Cache-Control no-cache;
    expires 12h;
}
# 代理资源文件到后端
location /m3u8/ {
    proxy_pass http://airplayTV;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header REMOTE-HOST $remote_addr;

    #Set Nginx Cache
    add_header Cache-Control no-cache;
    expires 12h;
}

```

## 启动api服务
- 拉取接口代码`git clone git@github.com:lixiang4u/airplayTV.git`
- 构建并启动服务
```code
go get
go build
chmod +x ./airplayTV
./airplayTV serve
```

## 访问nginx配置的域名测试

## 说明
- HLS.js更新到(V1.2.3)会导致部分视频无法播放
- 如果部分m3u8视频在本软件播放失败但是其他播放器可以播放，那么需要降级HLS.js更新到V12.0
- 编码时将第三方插件变量赋值给Vue的数据变量可能导致不可预期问题(Vue数据变量的Proxy机制引起)

## 免责声明

- 本软件仅作学习研究，请勿用于其它任何场景。作者不承担一切不正当使用本软件的责任。
