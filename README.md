# 五子棋游戏服务器

一个基于 Go 语言开发的五子棋在线对战游戏服务器，支持 TCP 长连接、用户认证、房间匹配、实时对战和排行榜功能。

## 技术栈

- **语言**: Go 1.21+
- **数据库**: MySQL 8.0
- **缓存**: Redis 7.0
- **网络协议**: HTTP + TCP 长连接

## 功能特性

- 用户注册/登录 (HTTP API)
- TCP 长连接通信
- 自定义二进制消息协议
- Token 会话管理
- 房间创建/加入/离开
- 实时五子棋对战
- 胜负判定算法
- 积分系统
- 排行榜

## 项目结构

```
.
├── cmd/
│   ├── server/main.go       # 服务端入口
│   └── client/main.go       # 测试客户端
├── configs/config.yaml      # 配置文件
├── internal/
│   ├── config/              # 配置读取
│   ├── handler/             # HTTP/TCP 处理器
│   ├── model/               # 数据模型
│   ├── repository/          # 数据访问层
│   ├── router/              # HTTP 路由
│   └── service/             # 业务逻辑层
├── pkg/
│   ├── protocol/            # 消息协议
│   └── redis/               # Redis 连接
└── scripts/init.sql         # 数据库初始化脚本
```

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/czzzk6639/goDemo.git
cd goDemo
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 初始化数据库

```bash
mysql -u root -p < scripts/init.sql
```

### 4. 配置文件

编辑 `configs/config.yaml`，配置 MySQL 和 Redis 连接信息：

```yaml
server:
  http_port: 8080
  tcp_port: 9000

mysql:
  host: 127.0.0.1
  port: 3306
  username: root
  password: your_password
  database: gomoku

redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0
```

### 5. 启动服务

```bash
go run ./cmd/server
```

服务启动后监听：
- HTTP: `http://localhost:8080`
- TCP: `localhost:9000`

## HTTP API

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/register | 用户注册 |
| POST | /api/login | 用户登录 |
| GET | /api/user/:id | 查询用户信息 |

### 示例

```bash
# 注册
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"player1","password":"123456"}'

# 登录
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"player1","password":"123456"}'

# 查询用户
curl http://localhost:8080/api/user/1
```

## TCP 消息协议

### 消息格式

```
| Len(4B) | Type(2B) | Seq(2B) | Payload(JSON) |
```

### 消息类型

| 类型码 | 名称 | 描述 |
|--------|------|------|
| 1000/1001 | Ping/Pong | 心跳 |
| 2001/2002 | LoginReq/Resp | 登录 |
| 2003/2004 | RegisterReq/Resp | 注册 |
| 3001/3011 | CreateRoomReq/Resp | 创建房间 |
| 3002/3012 | JoinRoomReq/Resp | 加入房间 |
| 3003/3013 | LeaveRoomReq/Resp | 离开房间 |
| 3004/3014 | RoomListReq/Resp | 房间列表 |
| 4001/4002 | MoveReq/Resp | 落子 |
| 4003 | GameOver | 游戏结束 |
| 4004 | GameStart | 游戏开始 |
| 4005 | BoardUpdate | 棋盘更新 |
| 5001/5002 | LeaderboardReq/Resp | 排行榜 |
| 5003/5004 | UserStatsReq/Resp | 用户统计 |

## 游戏规则

- 棋盘大小: 15 x 15
- 对战模式: 1v1
- 获胜条件: 横/竖/斜连续 5 子

## 积分系统

| 结果 | 积分变化 |
|------|----------|
| 胜利 | +25 |
| 失败 | -20 |
| 初始积分 | 1000 |

## 测试客户端

项目包含一个测试客户端：

```bash
go run ./cmd/client
```

## 依赖

- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
- [go-redis/v9](https://github.com/redis/go-redis)
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)

## License

MIT
