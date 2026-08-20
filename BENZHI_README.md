# cleanroom-monitor Docker 打包说明

本镜像为 GMP 洁净车间环境监测系统（cleanroom-monitor）的容器化打包产物。

## 镜像名

`benzhi/cleanroom-monitor-bug-<N>:latest`（N 为 bug 分支编号，如 bug-1）。

## 构建

```bash
./build_benzhi_docker.sh cleanroom-monitor-bug-1 linux/amd64
./build_benzhi_docker.sh cleanroom-monitor-bug-1 linux/arm64
```

## 运行

```bash
docker run --rm -p 8080:8080 -e CR_DB_PATH=/data/cleanroom.db \
  -v $PWD/data:/data benzhi/cleanroom-monitor-bug-1:latest
```

访问 http://localhost:8080/web/index.html 查看监控看板。

## 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| CR_PORT | 8080 | HTTP 端口 |
| CR_DB_PATH | cleanroom.db | SQLite 数据库文件路径（落盘） |
| CR_WEB_DIR | web | 前端静态资源目录 |
| CR_COLLECTOR_INTERVAL | 5s | 采集器轮询间隔 |
| CR_OFFLINE_AFTER | 90s | 传感器离线判定时长 |
| CR_EVAL_WORKERS | 4 | 告警评估 worker 数 |
| CR_SIMULATOR | true | 是否启用模拟器 |
| CR_SIM_INTERVAL | 3s | 模拟器生成间隔 |
| CR_SNAPSHOT_REFRESH | 5s | 看板快照刷新间隔 |