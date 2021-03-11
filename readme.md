# 参考资料

[源代码构建 Caddy](https://caddyserver.com/docs/build)

[Caddy 在线手册](https://caddyserver.com/docs)

# Docker 安装

```sh
# 构建 docker 镜像
docker build -t caddy2-markdown .
```

```sh
cd test
xcaddy build v2.3.0 --with github.com/maq128/caddy-markdown@v0.0.1=..
caddy run

http://localhost/
```

# Windows 系统中为 .md 文件指定 MIME type

在系统注册表中添加如下内容：
```
[HKEY_LOCAL_MACHINE\SOFTWARE\Classes\.md]
"Content Type"="text/markdown"
```