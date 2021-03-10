# 参考资料

[源代码构建 Caddy](https://caddyserver.com/docs/build)

[Caddy 在线手册](https://caddyserver.com/docs)

# Docker 安装

```sh
# 构建 docker 镜像
docker build -t caddy2-markdown .
```

```sh
xcaddy build v2.3.0 --with github.com/maq128/caddy-markdown@v0.0.1=.

cd test
caddy run

http://localhost/
http://localhost/index.md
http://localhost/simple.md
http://localhost/normal.md
http://localhost/github.md
```