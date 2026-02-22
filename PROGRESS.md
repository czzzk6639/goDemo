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
│   ├── handler/
│   │   ├── http_handler.go     # HTTP处理器
│   │   └── tcp_handler.go      # TCP处理器
│   ├── model/
│   │   ├── user.go             # 用户模型
│   │   ├── room.go             # 房间模型
│   │   └── game.go             # 游戏模型
│   ├── repository/
│   │   ├── db.go               # 数据库连接
│   │   ├── user_repo.go        # 用户数据访问
│   │   └── game_repository.go  # 游戏记录数据访问
│   ├── router/router.go        # HTTP路由
│   └── service/
│       ├── user_service.go     # 用户服务
│       ├── session.go          # 会话服务
│       ├── room_service.go     # 房间服务
│       ├── game_service.go     # 游戏服务
│       └── rank_service.go     # 排行榜服务
├── pkg/
│   ├── protocol/               # 消息协议
│   │   ├── packet.go           # 消息包结构
│   │   ├── codec.go            # 编解码器
│   │   └── message.go          # 消息类型
│   └── redis/redis.go          # Redis连接池
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

### ✅ 第二周已完成
- [x] TCP服务器实现 (端口9000)
- [x] 自定义二进制消息协议
- [x] 消息编解码器 (JSON序列化)
- [x] 消息类型定义 (登录、注册、房间、游戏等)
- [x] TCP消息处理 (Ping/Pong、Login、Register)
- [x] 客户端连接管理

### ✅ 第三周已完成
- [x] Redis连接池 (go-redis/v9)
- [x] 会话管理服务 (Create, Get, Validate, Delete, Refresh)
- [x] Token验证机制 (登录时创建会话，心跳时刷新会话)
- [x] 在线状态管理 (SetUserOnline/Offline)
- [x] 会话超时处理 (24小时TTL)

### ✅ 第四周已完成
- [x] 房间模型 (Room, RoomStatus)
- [x] 房间服务 (Create, Join, Leave, List, Delete)
- [x] 房间消息处理 (CreateRoom, JoinRoom, LeaveRoom, RoomList)
- [x] 房间广播消息 (PlayerJoin, PlayerLeave)
- [x] 断线自动离开房间

### ✅ 第五周已完成
- [x] 游戏模型 (Game, GameState, 15x15棋盘)
- [x] 游戏服务 (StartGame, MakeMove, Forfeit, EndGame)
- [x] 落子验证 (位置验证、回合验证、占用验证)
- [x] 五子棋胜负判定算法 (横、竖、斜四个方向)
- [x] 游戏消息处理 (Move, Forfeit)
- [x] 游戏开始通知 (GameStart)
- [x] 棋盘更新广播 (BoardUpdate)
- [x] 游戏结束通知 (GameOver)
- [x] 断线自动认输

### ✅ 第六周已完成
- [x] 游戏记录存储 (game_repository.go)
- [x] 用户积分更新 (UpdateScore)
- [x] 排行榜服务 (rank_service.go)
- [x] 排行榜查询 (LeaderboardReq/Resp)
- [x] 用户统计查询 (UserStatsReq/Resp)
- [x] 游戏结束自动更新积分 (胜+25, 负-20)
- [x] 项目全部功能完成

### API 接口
```
POST /api/register    # 用户注册
POST /api/login       # 用户登录 → 返回Token
GET  /api/user/{id}   # 查询用户信息
```

## 下一步任务

### 项目已完成所有核心功能

后续可优化方向：
1. 添加单元测试和集成测试
2. 实现游戏回放功能
3. 添加好友系统
4. 实现聊天功能
5. 添加观战功能
6. 性能优化和压力测试

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

服务启动后会同时监听：
- HTTP服务：端口8080
- TCP服务：端口9000

### 4. 测试HTTP API
```bash
curl -X POST http://localhost:8080/api/register -d "{\"username\":\"test\",\"password\":\"123456\"}"
curl -X POST http://localhost:8080/api/login -d "{\"username\":\"test\",\"password\":\"123456\"}"
curl http://localhost:8080/api/user/1
```

### 5. TCP消息协议
```
消息头 (8字节):
| Len(4字节) | Type(2字节) | Seq(2字节) |

消息体:
| Payload(JSON格式) |
```

消息类型：
- 1000: Ping / 1001: Pong
- 2001: LoginReq / 2002: LoginResp
- 2003: RegisterReq / 2004: RegisterResp
- 3001: CreateRoomReq / 3011: CreateRoomResp
- 3002: JoinRoomReq / 3012: JoinRoomResp
- 3003: LeaveRoomReq / 3013: LeaveRoomResp
- 3004: RoomListReq / 3014: RoomListResp
- 3015: PlayerJoin / 3016: PlayerLeave
- 4001: MoveReq / 4002: MoveResp
- 4003: GameOver / 4004: GameStart
- 4005: BoardUpdate / 4006: ForfeitReq / 4007: ForfeitResp
- 5001: LeaderboardReq / 5002: LeaderboardResp
- 5003: UserStatsReq / 5004: UserStatsResp

### 6. Redis数据结构
```
会话存储:
  Key: session:{token}
  Value: {user_id, username, token, created_at}
  TTL: 24小时

在线状态:
  Key: online:{user_id}
  Value: token
  TTL: 24小时
```

### 7. 游戏数据结构
```
棋盘: 15x15 二维数组
玩家: 2人对战
回合: 轮流落子
判定: 横/竖/斜连续5子获胜
```

### 8. 积分系统
```
初始积分: 1000分
获胜奖励: +25分
失败扣分: -20分
排行榜: 按积分降序排列
```

## 项目进度

| 周 | 任务 | 状态 |
|---|------|------|
| W1 | HTTP用户API + MySQL | ✅ 完成 |
| W2 | TCP服务器 + 消息协议 | ✅ 完成 |
| W3 | Redis会话 + Token验证 | ✅ 完成 |
| W4 | 房间系统 | ✅ 完成 |
| W5 | 游戏逻辑（落子、胜负判定） | ✅ 完成 |
| W6 | 排行榜 + 积分系统 | ✅ 完成 |

## 关键技术点
- 密码加密：bcrypt
- 配置解析：gopkg.in/yaml.v3
- 数据库：go-sql-driver/mysql
- Redis：go-redis/v9
- Token生成：crypto/rand + hex编码
- 会话存储：Redis (key: session:{token})
- 在线状态：Redis (key: online:{userID})
- 房间管理：内存存储 + sync.RWMutex
- 广播消息：遍历房间玩家发送
- 游戏逻辑：15x15棋盘、五子连珠判定
- 胜负算法：横、竖、正斜、反斜四个方向检测
- 积分系统：胜+25分，负-20分
- 排行榜：按积分降序排序
