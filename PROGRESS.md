# 五子棋游戏服务器 - 项目进度

## 项目概述
- 目标：为应聘游戏服务端工程师准备项目
- 技术栈：Go、MySQL、Redis、TCP长连接
- 核心能力展示：网络服务、数据存储、并发控制

## 项目结构
```
game-server/
├── cmd/server/main.go          # 程序入口
├── configs/config.yaml         # 配置文件
├── internal/
│   ├── config/config.go        # 配置读取
│   ├── handler/http_handler.go # HTTP处理器
│   ├── model/user.go           # 用户模型
│   ├── repository/             # 数据访问层
│   ├── router/router.go        # HTTP路由
│   └── service/user_service.go # 业务逻辑层
├── pkg/                        # 待开发：protocol、session、room
├── scripts/init.sql            # 数据库脚本
└── go.mod
```

## 当前进度

### ✅ 第一周已完成
- [x] 项目目录结构
- [x] 配置管理 (config.yaml + config.go)
- [x] 用户模型定义
- [x] MySQL 数据库连接
- [x] 用户 Repository (CreateUser, GetUserByUsername, GetUserByID)
- [x] 用户 Service (Register, Login, 密码bcrypt加密)
- [x] HTTP Handler (Register, Login, GetUser)
- [x] HTTP 路由配置
- [x] 主程序入口
- [x] 数据库初始化脚本 (users表、games表)
- [x] 项目编译成功

### API 接口
```
POST /api/register    # 用户注册
POST /api/login       # 用户登录 → 返回Token
GET  /api/user/{id}   # 查询用户信息
```

## 下一步任务：第二周

### 目标
- TCP服务器实现
- 自定义二进制消息协议
- 消息编解码器

### 具体任务
1. 创建 pkg/protocol/packet.go - 消息包结构定义
2. 创建 pkg/protocol/codec.go - 编解码器
3. 创建 pkg/protocol/message.go - 消息类型定义
4. 创建 internal/handler/tcp_handler.go - TCP消息处理
5. 修改 cmd/server/main.go - 启动TCP服务器

### TCP消息协议设计
```
消息头 (8字节):
| Len(4字节) | Type(2字节) | Seq(2字节) |

消息体:
| Payload(N字节) |
```

## 启动指南

### 1. 初始化数据库
```bash
mysql -u root -p < scripts/init.sql
```

### 2. 修改配置
编辑 configs/config.yaml，设置MySQL密码

### 3. 启动服务
```bash
go run ./cmd/server
# 或
./bin/server.exe
```

### 4. 测试API
```bash
curl -X POST http://localhost:8080/api/register -d "{\"username\":\"test\",\"password\":\"123456\"}"
curl -X POST http://localhost:8080/api/login -d "{\"username\":\"test\",\"password\":\"123456\"}"
curl http://localhost:8080/api/user/1
```

## 后续计划

| 周 | 任务 | 状态 |
|---|------|------|
| W1 | HTTP用户API + MySQL | ✅ 完成 |
| W2 | TCP服务器 + 消息协议 | ⏳ 待开始 |
| W3 | Redis会话 + Token验证 | ⏳ 待开始 |
| W4 | 房间系统 | ⏳ 待开始 |
| W5 | 游戏逻辑（落子、胜负判定） | ⏳ 待开始 |
| W6 | 排行榜 + 整合测试 | ⏳ 待开始 |

## 关键技术点
- 密码加密：bcrypt
- 配置解析：gopkg.in/yaml.v3
- 数据库：go-sql-driver/mysql
- Token生成：crypto/rand + hex编码
