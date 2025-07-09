# RocketMQ 测试环境

此目录包含用于搭建本地 RocketMQ 测试环境的 Docker Compose 配置。

## ✨ 特性

- 🚀 **简化配置** - 无数据持久化，适合测试环境
- 🔧 **内存优化** - 降低内存使用，适合本地开发
- 🌐 **管理界面** - 包含 RocketMQ Dashboard
- 📦 **开箱即用** - 自动配置，无需手动设置

## 📦 包含服务

| 服务 | 端口 | 容器名 | 描述 |
|------|------|--------|------|
| **NameServer** | 9876 | rmqnamesrv | 服务注册与发现 |
| **Broker** | 10911 | rmqbroker | 消息代理服务 |
| **Dashboard** | 8080 | rmqdashboard | Web管理界面 |

## 🚀 快速开始

### 方法1：使用启动脚本（推荐）

```bash
# 运行启动脚本
./start.sh
```

### 方法2：手动启动

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 停止服务

```bash
# 停止服务
docker-compose down

# 停止并清理所有数据
docker-compose down --volumes --remove-orphans
```

## 🔗 连接信息

### MQ Toolkit 配置

在 MQ Toolkit 中创建 RocketMQ 连接时使用以下配置：

```
连接类型: RocketMQ
主机: localhost
端口: 9876
用户名: (留空)
密码: (留空)
```

### 管理界面

- **RocketMQ Dashboard**: http://localhost:8080
  - 查看主题、消息、消费者状态
  - 监控集群健康状况
  - 💡 **提示**: 如果您的RocketMQ部署在其他主机上，请将localhost替换为实际的主机地址

## 🔍 故障排除

### 检查服务状态

```bash
# 查看容器状态
docker ps --filter "name=rmq"

# 查看服务日志
docker logs rmqnamesrv
docker logs rmqbroker
docker logs rmqdashboard
```

### 测试连接

```bash
# 测试 NameServer 端口
telnet localhost 9876

# 测试 Broker 端口
telnet localhost 10911
```

### 常见问题

1. **端口被占用**
   ```bash
   # 检查端口占用
   lsof -i :9876
   lsof -i :10911
   lsof -i :8080
   ```

2. **内存不足**
   - 确保 Docker 分配至少 2GB 内存
   - 可以在 docker-compose.yml 中进一步减少内存配置

3. **连接失败**
   - 确保防火墙允许相关端口
   - 检查 Docker 网络配置

## 📝 配置说明

### 内存配置

- **NameServer**: 256MB (JAVA_OPT_EXT=-Xms256m -Xmx256m)
- **Broker**: 512MB (JAVA_OPT_EXT=-Xms512m -Xmx512m)

### 网络配置

- 使用自定义网络 `rocketmq-network`
- Broker IP 设置为 `localhost` 以支持外部连接

### 数据持久化

- ❌ **无数据持久化** - 容器重启后数据清空
- ✅ **适合测试** - 每次启动都是干净环境
- ✅ **节省空间** - 不占用宿主机存储

## 🔧 自定义配置

如需修改配置，可以编辑 `docker-compose.yml` 文件：

```yaml
# 修改内存配置
environment:
  - JAVA_OPT_EXT=-Xms128m -Xmx128m  # 减少内存使用

# 修改端口映射
ports:
  - "19876:9876"  # 使用不同的宿主机端口
```

## 📚 相关链接

- [RocketMQ 官方文档](https://rocketmq.apache.org/)
- [RocketMQ Docker 镜像](https://hub.docker.com/r/apache/rocketmq)
- [RocketMQ Dashboard](https://github.com/apache/rocketmq-dashboard)