# RocketMQ 测试环境

此目录包含用于搭建本地 RocketMQ 测试环境的 `docker-compose.yml` 文件。

这个配置经过了简化和修复，以确保在本地开发环境中可以从外部（如桌面应用）正确连接。

## 包含服务

- **namesrv**: RocketMQ 的名字服务。
- **broker**: RocketMQ 的消息代理服务。
- **dashboard**: RocketMQ 的管理后台，方便查看消息和状态。

## 使用方法

### 启动环境

在当前目录下运行以下命令来启动所有 RocketMQ 服务：

```bash
docker-compose up -d
```

### 停止环境

如果想停止服务但保留数据，请运行：

```bash
docker-compose down
```

### 访问信息

- **NameServer 地址**: `localhost:9876`
- **Broker (对外) 地址**: `host.docker.internal:10911` (请在应用中用此地址连接)
- **管理后台**: [http://localhost:8080](http://localhost:8080)

### 清理数据

如果需要停止服务并删除所有相关的数据卷（此操作会删除所有消息和日志），请运行：

```bash
docker-compose down -v
```