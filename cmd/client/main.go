package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	HeaderLen  = 8
	MaxBodyLen = 65535
)

const (
	TypePing            uint16 = 1000
	TypePong            uint16 = 1001
	TypeLogin           uint16 = 2001
	TypeLoginResp       uint16 = 2002
	TypeRegister        uint16 = 2003
	TypeRegisterResp    uint16 = 2004
	TypeCreateRoom      uint16 = 3001
	TypeCreateRoomResp  uint16 = 3011
	TypeJoinRoom        uint16 = 3002
	TypeJoinRoomResp    uint16 = 3012
	TypeLeaveRoom       uint16 = 3003
	TypeLeaveRoomResp   uint16 = 3013
	TypeRoomList        uint16 = 3004
	TypeRoomListResp    uint16 = 3014
	TypePlayerJoin      uint16 = 3015
	TypePlayerLeave     uint16 = 3016
	TypeMove            uint16 = 4001
	TypeMoveResp        uint16 = 4002
	TypeGameOver        uint16 = 4003
	TypeGameStart       uint16 = 4004
	TypeBoardUpdate     uint16 = 4005
	TypeForfeitReq      uint16 = 4006
	TypeForfeitResp     uint16 = 4007
	TypeLeaderboardReq  uint16 = 5001
	TypeLeaderboardResp uint16 = 5002
	TypeUserStatsReq    uint16 = 5003
	TypeUserStatsResp   uint16 = 5004
)

type Packet struct {
	Len     uint32
	Type    uint16
	Seq     uint16
	Payload []byte
}

type Client struct {
	conn     net.Conn
	reader   *bufio.Reader
	seq      uint16
	mu       sync.Mutex
	userID   int64
	username string
	token    string
	roomID   int64
}

func NewClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) nextSeq() uint16 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seq++
	return c.seq
}

func (c *Client) send(msgType uint16, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	pkt := &Packet{
		Len:     uint32(HeaderLen + len(data)),
		Type:    msgType,
		Seq:     c.nextSeq(),
		Payload: data,
	}

	buf := make([]byte, HeaderLen+len(data))
	binary.BigEndian.PutUint32(buf[0:4], pkt.Len)
	binary.BigEndian.PutUint16(buf[4:6], pkt.Type)
	binary.BigEndian.PutUint16(buf[6:8], pkt.Seq)
	copy(buf[HeaderLen:], pkt.Payload)

	_, err = c.conn.Write(buf)
	return err
}

func (c *Client) recv() (*Packet, error) {
	header := make([]byte, HeaderLen)
	if _, err := c.reader.Read(header); err != nil {
		return nil, err
	}

	pkt := &Packet{
		Len:  binary.BigEndian.Uint32(header[0:4]),
		Type: binary.BigEndian.Uint16(header[4:6]),
		Seq:  binary.BigEndian.Uint16(header[6:8]),
	}

	bodyLen := pkt.Len - HeaderLen
	if bodyLen > 0 {
		pkt.Payload = make([]byte, bodyLen)
		if _, err := c.reader.Read(pkt.Payload); err != nil {
			return nil, err
		}
	}

	return pkt, nil
}

func (c *Client) recvLoop() {
	for {
		pkt, err := c.recv()
		if err != nil {
			fmt.Println("\n[Disconnected from server]")
			os.Exit(0)
		}
		c.handlePacket(pkt)
	}
}

func (c *Client) handlePacket(pkt *Packet) {
	switch pkt.Type {
	case TypePong:
		fmt.Println("\n[Pong received]")
	case TypeLoginResp:
		var resp map[string]interface{}
		json.Unmarshal(pkt.Payload, &resp)
		if resp["code"].(float64) == 200 {
			c.token = resp["token"].(string)
			c.userID = int64(resp["user_id"].(float64))
			fmt.Printf("\n[Login success] UserID: %d, Token: %s\n", c.userID, c.token[:16]+"...")
		} else {
			fmt.Printf("\n[Login failed] %s\n", resp["message"])
		}
	case TypeRegisterResp:
		var resp map[string]interface{}
		json.Unmarshal(pkt.Payload, &resp)
		if resp["code"].(float64) == 200 {
			fmt.Printf("\n[Register success] UserID: %d\n", int64(resp["user_id"].(float64)))
		} else {
			fmt.Printf("\n[Register failed] %s\n", resp["message"])
		}
	case TypeCreateRoomResp:
		var resp map[string]interface{}
		json.Unmarshal(pkt.Payload, &resp)
		if resp["code"].(float64) == 200 {
			c.roomID = int64(resp["room_id"].(float64))
			fmt.Printf("\n[Room created] RoomID: %d\n", c.roomID)
		} else {
			fmt.Printf("\n[Create room failed] %s\n", resp["message"])
		}
	case TypeJoinRoomResp:
		var resp map[string]interface{}
		json.Unmarshal(pkt.Payload, &resp)
		if resp["code"].(float64) == 200 {
			c.roomID = int64(resp["room_id"].(float64))
			fmt.Printf("\n[Joined room] RoomID: %d\n", c.roomID)
		} else {
			fmt.Printf("\n[Join room failed] %s\n", resp["message"])
		}
	case TypeLeaveRoomResp:
		var resp map[string]interface{}
		json.Unmarshal(pkt.Payload, &resp)
		c.roomID = 0
		fmt.Printf("\n[Left room] %s\n", resp["message"])
	case TypeRoomListResp:
		var resp map[string]interface{}
		json.Unmarshal(pkt.Payload, &resp)
		rooms := resp["rooms"].([]interface{})
		fmt.Printf("\n[Room List] %d rooms\n", len(rooms))
		for _, r := range rooms {
			room := r.(map[string]interface{})
			fmt.Printf("  Room %d: %s (Players: %d, Status: %d)\n",
				int64(room["room_id"].(float64)),
				room["room_name"],
				len(room["players"].([]interface{})),
				int(room["status"].(float64)))
		}
	case TypePlayerJoin:
		var msg map[string]interface{}
		json.Unmarshal(pkt.Payload, &msg)
		fmt.Printf("\n[Player joined] %s (ID: %d)\n", msg["username"], int64(msg["user_id"].(float64)))
	case TypePlayerLeave:
		var msg map[string]interface{}
		json.Unmarshal(pkt.Payload, &msg)
		fmt.Printf("\n[Player left] ID: %d, Reason: %s\n", int64(msg["user_id"].(float64)), msg["reason"])
	case TypeGameStart:
		var msg map[string]interface{}
		json.Unmarshal(pkt.Payload, &msg)
		players := msg["players"].([]interface{})
		first := int64(msg["first_player"].(float64))
		fmt.Printf("\n[Game started!] Players: %v, First: %d\n", players, first)
		fmt.Println("Use 'move x y' to place your piece (0-14)")
	case TypeBoardUpdate:
		var msg map[string]interface{}
		json.Unmarshal(pkt.Payload, &msg)
		board := msg["board"].([]interface{})
		current := int64(msg["current_player"].(float64))
		c.printBoard(board)
		fmt.Printf("\n[Current turn: Player %d]\n", current)
	case TypeMoveResp:
		var resp map[string]interface{}
		json.Unmarshal(pkt.Payload, &resp)
		if resp["code"].(float64) == 200 {
			fmt.Printf("\n[Move success] (%d, %d)\n", int(resp["x"].(float64)), int(resp["y"].(float64)))
		} else {
			fmt.Printf("\n[Move failed] %s\n", resp["message"])
		}
	case TypeGameOver:
		var msg map[string]interface{}
		json.Unmarshal(pkt.Payload, &msg)
		winner := int64(msg["winner"].(float64))
		if winner == c.userID {
			fmt.Printf("\n*** YOU WIN! ***\n")
		} else if winner != 0 {
			fmt.Printf("\n*** YOU LOSE! Winner: %d ***\n", winner)
		} else {
			fmt.Println("\n*** DRAW! ***")
		}
		c.roomID = 0
	case TypeLeaderboardResp:
		var resp map[string]interface{}
		json.Unmarshal(pkt.Payload, &resp)
		ranks := resp["ranks"].([]interface{})
		fmt.Printf("\n[Leaderboard] Top %d\n", len(ranks))
		fmt.Println("Rank | Username       | Score | W/L    | WinRate")
		fmt.Println("-----|----------------|-------|--------|--------")
		for _, r := range ranks {
			entry := r.(map[string]interface{})
			fmt.Printf("%4d | %-14s | %5d | %d/%-3d | %s\n",
				int(entry["rank"].(float64)),
				entry["username"],
				int(entry["score"].(float64)),
				int(entry["win_count"].(float64)),
				int(entry["lose_count"].(float64)),
				entry["win_rate"])
		}
	case TypeUserStatsResp:
		var resp map[string]interface{}
		json.Unmarshal(pkt.Payload, &resp)
		if resp["code"].(float64) == 200 {
			fmt.Printf("\n[User Stats]\n")
			fmt.Printf("  Username: %s\n", resp["username"])
			fmt.Printf("  Score: %d\n", int(resp["score"].(float64)))
			fmt.Printf("  W/L: %d/%d\n", int(resp["win_count"].(float64)), int(resp["lose_count"].(float64)))
			fmt.Printf("  Win Rate: %s\n", resp["win_rate"])
			fmt.Printf("  Rank: #%d\n", int(resp["rank"].(float64)))
		} else {
			fmt.Printf("\n[Get stats failed] %s\n", resp["message"])
		}
	default:
		fmt.Printf("\n[Unknown message type: %d] %s\n", pkt.Type, string(pkt.Payload))
	}
	printPrompt()
}

func (c *Client) printBoard(board []interface{}) {
	fmt.Println("\n   0 1 2 3 4 5 6 7 8 9 0 1 2 3 4")
	fmt.Println("   ------------------------------")
	for i, row := range board {
		fmt.Printf("%2d|", i)
		for _, cell := range row.([]interface{}) {
			val := int(cell.(float64))
			if val == 0 {
				fmt.Print(" .")
			} else if val == 1 {
				fmt.Print(" X")
			} else {
				fmt.Print(" O")
			}
		}
		fmt.Println()
	}
}

func printPrompt() {
	fmt.Print("> ")
}

func printHelp() {
	fmt.Println(`
Commands:
  register <username> <password>  - Register new user
  login <username> <password>     - Login with username/password
  login-token <token>             - Login with token
  create [room_name]              - Create a room
  join <room_id>                  - Join a room
  leave                           - Leave current room
  rooms                           - List waiting rooms
  move <x> <y>                    - Make a move (0-14)
  forfeit                         - Forfeit current game
  leaderboard [limit]             - Show leaderboard
  stats [user_id]                 - Show user stats
  ping                            - Send ping
  help                            - Show this help
  quit                            - Exit

Game flow:
  1. register user1 123456
  2. login user1 123456
  3. create MyRoom
  4. (another client) join <room_id>
  5. Game starts automatically
  6. move 7 7  (place at center)
`)
}

func main() {
	addr := "localhost:9000"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	fmt.Printf("Connecting to %s...\n", addr)
	client, err := NewClient(addr)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Connected! Type 'help' for commands.")
	go client.recvLoop()

	scanner := bufio.NewScanner(os.Stdin)
	printPrompt()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			printPrompt()
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return
		case "help":
			printHelp()
		case "ping":
			client.send(TypePing, struct{}{})
		case "register":
			if len(args) < 2 {
				fmt.Println("Usage: register <username> <password>")
			} else {
				client.send(TypeRegister, map[string]string{
					"username": args[0],
					"password": args[1],
				})
			}
		case "login":
			if len(args) < 2 {
				fmt.Println("Usage: login <username> <password>")
			} else {
				client.username = args[0]
				client.send(TypeLogin, map[string]string{
					"username": args[0],
					"password": args[1],
				})
			}
		case "login-token":
			if len(args) < 1 {
				fmt.Println("Usage: login-token <token>")
			} else {
				client.send(TypeLogin, map[string]string{
					"token": args[0],
				})
			}
		case "create":
			name := "Game Room"
			if len(args) > 0 {
				name = strings.Join(args, " ")
			}
			client.send(TypeCreateRoom, map[string]string{
				"room_name": name,
			})
		case "join":
			if len(args) < 1 {
				fmt.Println("Usage: join <room_id>")
			} else {
				var roomID int64
				fmt.Sscanf(args[0], "%d", &roomID)
				client.send(TypeJoinRoom, map[string]int64{
					"room_id": roomID,
				})
			}
		case "leave":
			if client.roomID == 0 {
				fmt.Println("Not in a room")
			} else {
				client.send(TypeLeaveRoom, map[string]int64{
					"room_id": client.roomID,
				})
			}
		case "rooms":
			client.send(TypeRoomList, struct{}{})
		case "move":
			if len(args) < 2 {
				fmt.Println("Usage: move <x> <y>")
			} else {
				var x, y int
				fmt.Sscanf(args[0], "%d", &x)
				fmt.Sscanf(args[1], "%d", &y)
				client.send(TypeMove, map[string]interface{}{
					"room_id": client.roomID,
					"x":       x,
					"y":       y,
				})
			}
		case "forfeit":
			if client.roomID == 0 {
				fmt.Println("Not in a game")
			} else {
				client.send(TypeForfeitReq, map[string]int64{
					"room_id": client.roomID,
				})
			}
		case "leaderboard":
			limit := 10
			if len(args) > 0 {
				fmt.Sscanf(args[0], "%d", &limit)
			}
			client.send(TypeLeaderboardReq, map[string]int{
				"limit":  limit,
				"offset": 0,
			})
		case "stats":
			req := map[string]int64{}
			if len(args) > 0 {
				var userID int64
				fmt.Sscanf(args[0], "%d", &userID)
				req["user_id"] = userID
			} else if client.userID != 0 {
				req["user_id"] = client.userID
			}
			client.send(TypeUserStatsReq, req)
		default:
			fmt.Printf("Unknown command: %s\n", cmd)
		}
	}
}
