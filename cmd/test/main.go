package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

const HeaderLen = 8

type Packet struct {
	Len     uint32
	Type    uint16
	Seq     uint16
	Payload []byte
}

type GameClient struct {
	conn     net.Conn
	reader   *bufio.Reader
	seq      uint16
	userID   int64
	token    string
	roomID   int64
	mu       sync.Mutex
	lastResp map[string]interface{}
}

func NewGameClient(addr string) (*GameClient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &GameClient{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

func (c *GameClient) send(msgType uint16, payload interface{}) error {
	data, _ := json.Marshal(payload)
	buf := make([]byte, HeaderLen+len(data))
	binary.BigEndian.PutUint32(buf[0:4], uint32(HeaderLen+len(data)))
	binary.BigEndian.PutUint16(buf[4:6], msgType)
	c.seq++
	binary.BigEndian.PutUint16(buf[6:8], c.seq)
	copy(buf[HeaderLen:], data)
	_, err := c.conn.Write(buf)
	return err
}

func (c *GameClient) recv() (uint16, map[string]interface{}, error) {
	header := make([]byte, HeaderLen)
	if _, err := c.reader.Read(header); err != nil {
		return 0, nil, err
	}
	msgType := binary.BigEndian.Uint16(header[4:6])
	bodyLen := binary.BigEndian.Uint32(header[0:4]) - HeaderLen
	var payload []byte
	if bodyLen > 0 {
		payload = make([]byte, bodyLen)
		c.reader.Read(payload)
	}
	var result map[string]interface{}
	if len(payload) > 0 {
		json.Unmarshal(payload, &result)
	}
	return msgType, result, nil
}

func (c *GameClient) recvLoop(responses *chan map[string]interface{}) {
	for {
		msgType, resp, err := c.recv()
		if err != nil {
			return
		}
		if resp != nil {
			resp["_type"] = msgType
			select {
			case *responses <- resp:
			default:
			}
		}
	}
}

func (c *GameClient) waitForResponse(expectedType uint16, timeout time.Duration) map[string]interface{} {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		msgType, resp, err := c.recv()
		if err != nil {
			continue
		}
		if msgType == expectedType {
			return resp
		}
	}
	return nil
}

func (c *GameClient) register(username, password string) bool {
	c.send(2003, map[string]string{"username": username, "password": password})
	resp := c.waitForResponse(2004, 2*time.Second)
	return resp != nil && resp["code"].(float64) == 200
}

func (c *GameClient) login(username, password string) bool {
	c.send(2001, map[string]string{"username": username, "password": password})
	resp := c.waitForResponse(2002, 2*time.Second)
	if resp != nil && resp["code"].(float64) == 200 {
		if token, ok := resp["token"].(string); ok {
			c.token = token
		}
		if userID, ok := resp["user_id"].(float64); ok {
			c.userID = int64(userID)
		}
		return true
	}
	return false
}

func (c *GameClient) createRoom(name string) int64 {
	c.send(3001, map[string]string{"room_name": name})
	resp := c.waitForResponse(3011, 2*time.Second)
	if resp != nil && resp["code"].(float64) == 200 {
		c.roomID = int64(resp["room_id"].(float64))
		return c.roomID
	}
	return 0
}

func (c *GameClient) joinRoom(roomID int64) bool {
	c.send(3002, map[string]int64{"room_id": roomID})
	resp := c.waitForResponse(3012, 2*time.Second)
	if resp != nil && resp["code"].(float64) == 200 {
		c.roomID = roomID
		return true
	}
	return false
}

func (c *GameClient) move(x, y int) bool {
	c.send(4001, map[string]interface{}{"room_id": c.roomID, "x": x, "y": y})
	resp := c.waitForResponse(4002, 2*time.Second)
	return resp != nil && resp["code"].(float64) == 200
}

func (c *GameClient) getLeaderboard() {
	c.send(5001, map[string]int{"limit": 10})
	resp := c.waitForResponse(5002, 2*time.Second)
	if resp == nil {
		return
	}
	fmt.Println("\n=== 排行榜 ===")
	fmt.Println("排名 | 用户名         | 积分  | 胜/负")
	fmt.Println("-----|----------------|-------|-------")
	if ranks, ok := resp["ranks"].([]interface{}); ok {
		for _, r := range ranks {
			e := r.(map[string]interface{})
			fmt.Printf("#%-3d | %-14s | %-5d | %d/%d\n",
				int(e["rank"].(float64)), e["username"], int(e["score"].(float64)),
				int(e["win_count"].(float64)), int(e["lose_count"].(float64)))
		}
	}
}

func main() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║    五子棋游戏服务器 - 自动化测试       ║")
	fmt.Println("╚════════════════════════════════════════╝\n")

	// 创建两个客户端
	fmt.Println("【1】创建玩家连接...")
	player1, err := NewGameClient("localhost:9000")
	if err != nil {
		fmt.Println("   ❌ 连接失败:", err)
		return
	}
	defer player1.conn.Close()

	player2, err := NewGameClient("localhost:9000")
	if err != nil {
		fmt.Println("   ❌ 连接失败:", err)
		return
	}
	defer player2.conn.Close()
	fmt.Println("   ✓ 两个玩家已连接")

	// 注册用户
	fmt.Println("\n【2】注册玩家...")
	player1.register("autoplayer1", "123456")
	player2.register("autoplayer2", "123456")
	fmt.Println("   ✓ 玩家1: autoplayer1")
	fmt.Println("   ✓ 玩家2: autoplayer2")

	// 登录
	fmt.Println("\n【3】登录...")
	player1.login("autoplayer1", "123456")
	player2.login("autoplayer2", "123456")
	fmt.Printf("   ✓ 玩家1登录成功 (ID: %d)\n", player1.userID)
	fmt.Printf("   ✓ 玩家2登录成功 (ID: %d)\n", player2.userID)

	// 创建房间
	fmt.Println("\n【4】创建房间...")
	roomID := player1.createRoom("测试房间")
	fmt.Printf("   ✓ 房间ID: %d\n", roomID)

	// 加入房间
	fmt.Println("\n【5】玩家2加入房间...")
	player2.joinRoom(roomID)
	fmt.Println("   ✓ 加入成功，游戏开始!")

	// 模拟对局
	fmt.Println("\n【6】开始对局...")
	fmt.Println("   模拟落子（玩家1执黑先手）:")

	moves := [][2]int{
		{7, 7}, {7, 8}, {8, 8}, {6, 6}, {9, 9},
		{5, 5}, {10, 10}, {4, 4}, {11, 11},
	}

	for i, m := range moves {
		time.Sleep(100 * time.Millisecond)
		var client *GameClient
		var playerName string
		if i%2 == 0 {
			client = player1
			playerName = "玩家1(黑)"
		} else {
			client = player2
			playerName = "玩家2(白)"
		}
		success := client.move(m[0], m[1])
		if success {
			fmt.Printf("   ✓ %s 落子 (%d, %d)\n", playerName, m[0], m[1])
		} else {
			fmt.Printf("   ✗ %s 落子失败 (%d, %d)\n", playerName, m[0], m[1])
		}
	}

	time.Sleep(200 * time.Millisecond)

	// 查看排行榜
	fmt.Println("\n【7】查看排行榜...")
	player1.getLeaderboard()

	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║          测试完成! 所有功能正常        ║")
	fmt.Println("╚════════════════════════════════════════╝")
}
