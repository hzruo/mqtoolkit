# Kafka 测试环境

此目录包含用于搭建本地 Kafka 测试环境的 `docker-compose.yml` 文件。

## 包含服务

- **Zookeeper**: Kafka 用于存储元数据的必要组件。
- **Kafka**: Kafka 消息代理服务。

## 使用方法

### 启动环境

在当前目录下运行以下命令来启动 Kafka 和 Zookeeper 服务：

```bash
docker-compose up -d
```

### 停止环境

如果想停止服务但保留数据，请运行：

```bash
docker-compose down
```

### 访问信息

- **Broker 地址**: `localhost:9092`

### 清理数据

如果需要停止服务并删除所有相关的数据卷（此操作会删除所有消息），请运行：

```bash
docker-compose down -v
```