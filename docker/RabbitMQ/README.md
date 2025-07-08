# RabbitMQ 测试环境

此目录包含用于搭建本地 RabbitMQ 测试环境的 `docker-compose.yml` 文件。

## 包含服务

- **RabbitMQ**: RabbitMQ 消息代理服务，并已启用管理后台插件。

## 使用方法

### 启动环境

在当前目录下运行以下命令来启动 RabbitMQ 服务：

```bash
docker-compose up -d
```

### 停止环境

如果想停止服务但保留数据，请运行：

```bash
docker-compose down
```

### 访问信息

- **服务端口**: `5672`
- **管理后台**: [http://localhost:15672](http://localhost:15672)
- **默认凭证**:
  - **用户名**: `guest`
  - **密码**: `guest`

### 清理数据

如果需要停止服务并删除所有相关的数据卷（此操作会删除所有消息和配置），请运行：

```bash
docker-compose down -v
```