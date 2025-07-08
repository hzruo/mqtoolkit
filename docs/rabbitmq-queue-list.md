# RabbitMQ 队列列表功能

## 📋 概述

MQ Toolkit 现在支持显示 RabbitMQ 的队列列表。由于 RabbitMQ 的 AMQP 协议没有直接的 API 来列出所有队列，我们使用 RabbitMQ 的 HTTP 管理 API 来实现这个功能。

## 🔧 实现原理

### 技术方案
- **AMQP 连接**: 用于消息发送和接收
- **HTTP 管理 API**: 用于获取队列列表和管理操作
- **双重认证**: AMQP 和 HTTP API 使用相同的用户名密码

### API 端点
```
GET http://localhost:15672/api/queues/{vhost}
```

### 认证方式
- HTTP Basic Authentication
- 用户名/密码与 AMQP 连接相同
- 默认: guest/guest

## 📦 前置要求

### 1. RabbitMQ 管理插件
管理插件必须启用才能使用 HTTP API：

```bash
# 启用管理插件
rabbitmq-plugins enable rabbitmq_management

# 检查插件状态
rabbitmq-plugins list
```

### 2. 管理界面访问
- **URL**: http://localhost:15672
- **默认端口**: 15672
- **用户名**: guest
- **密码**: guest

### 3. 防火墙设置
确保端口 15672 可访问：

```bash
# Ubuntu/Debian
sudo ufw allow 15672

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=15672/tcp
sudo firewall-cmd --reload
```

## 🚀 使用方法

### 1. 配置连接
在 MQ Toolkit 中创建 RabbitMQ 连接：
- **主机**: localhost
- **端口**: 5672 (AMQP)
- **用户名**: guest (或留空使用默认值)
- **密码**: guest (或留空使用默认值)
- **VHost**: / (或留空使用默认值)

### 2. 查看队列列表
1. 选择 RabbitMQ 连接
2. 进入"主题/队列"标签页
3. 队列列表会自动加载显示

### 3. 队列信息
显示的队列信息包括：
- **队列名称**
- **分区数**: 固定为 1 (RabbitMQ 没有分区概念)
- **副本数**: 固定为 1 (RabbitMQ 没有副本概念)

## 🔍 故障排除

### 问题1: 队列列表为空
**可能原因**:
- RabbitMQ 管理插件未启用
- 管理界面端口 (15672) 不可访问
- 认证失败

**解决方案**:
```bash
# 检查管理插件
rabbitmq-plugins list | grep management

# 启用管理插件
rabbitmq-plugins enable rabbitmq_management

# 重启 RabbitMQ
sudo systemctl restart rabbitmq-server

# 测试管理界面
curl -u guest:guest http://localhost:15672/api/queues/%2F
```

### 问题2: 认证失败
**可能原因**:
- 用户名密码错误
- 用户权限不足

**解决方案**:
```bash
# 检查用户列表
rabbitmqctl list_users

# 创建管理员用户
rabbitmqctl add_user admin admin
rabbitmqctl set_user_tags admin administrator
rabbitmqctl set_permissions -p / admin ".*" ".*" ".*"
```

### 问题3: 端口访问问题
**可能原因**:
- 防火墙阻止
- RabbitMQ 配置问题

**解决方案**:
```bash
# 检查端口监听
netstat -ln | grep 15672

# 检查防火墙
sudo ufw status

# 测试连接
telnet localhost 15672
```