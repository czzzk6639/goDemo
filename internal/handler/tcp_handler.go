package handler

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"game-server/internal/model"
	"game-server/internal/service"
	"game-server/pkg/protocol"
)

type TCPHandler struct {
	userService    *service.UserService
	sessionService *service.SessionService
	roomService    *service.RoomService
	gameService    *service.GameService
	rankService    *service.RankService
	codec          *protocol.Codec
	clients        map[int64]*Client
	mu             sync.RWMutex
	seqCounter     uint64
}

type Client struct {
	Conn       net.Conn
	UserID     int64
	Token      string
	Username   string
	LastActive time.Time
	RoomID     int64
}

func NewTCPHandler() *TCPHandler {
	return &TCPHandler{
		userService:    service.NewUserService(),
		sessionService: service.NewSessionService(),
		roomService:    service.NewRoomService(),
		gameService:    service.NewGameService(),
		rankService:    service.NewRankService(),
		codec:          protocol.NewCodec(),
		clients:        make(map[int64]*Client),
	}
}

func (h *TCPHandler) HandleConn(conn net.Conn) {
	defer conn.Close()

	log.Printf("New connection from %s", conn.RemoteAddr())

	var client *Client

	for {
		pkt, err := protocol.ReadPacket(conn)
		if err != nil {
			log.Printf("Read packet error: %v", err)
			if client != nil {
				h.RemoveClient(client.UserID)
				h.sessionService.SetUserOffline(client.UserID)
				if client.RoomID != 0 {
					h.handleDisconnect(client)
				}
			}
			return
		}

		msg, err := h.codec.Decode(pkt)
		if err != nil {
			log.Printf("Decode error: %v", err)
			h.sendError(conn, pkt.Seq, 400, err.Error())
			continue
		}

		if client != nil {
			client.LastActive = time.Now()
		}

		switch m := msg.(type) {
		case *protocol.PingReq:
			h.handlePing(conn, pkt.Seq, client)
		case *protocol.LoginReq:
			client = h.handleLogin(conn, pkt.Seq, m)
		case *protocol.RegisterReq:
			h.handleRegister(conn, pkt.Seq, m)
		default:
			if client == nil {
				h.sendError(conn, pkt.Seq, 401, "please login first")
				continue
			}
			h.handleAuthMessage(conn, pkt.Seq, client, msg)
		}
	}
}

func (h *TCPHandler) handleAuthMessage(conn net.Conn, seq uint16, client *Client, msg protocol.Message) {
	switch m := msg.(type) {
	case *protocol.CreateRoomReq:
		h.handleCreateRoom(conn, seq, client, m)
	case *protocol.JoinRoomReq:
		h.handleJoinRoom(conn, seq, client, m)
	case *protocol.LeaveRoomReq:
		h.handleLeaveRoom(conn, seq, client, m)
	case *protocol.RoomListReq:
		h.handleRoomList(conn, seq, client, m)
	case *protocol.MoveReq:
		h.handleMove(conn, seq, client, m)
	case *protocol.ForfeitReq:
		h.handleForfeit(conn, seq, client, m)
	case *protocol.LeaderboardReq:
		h.handleLeaderboard(conn, seq, client, m)
	case *protocol.UserStatsReq:
		h.handleUserStats(conn, seq, client, m)
	default:
		log.Printf("Unhandled message type: %T", m)
		h.sendError(conn, seq, 400, "unknown message type")
	}
}

func (h *TCPHandler) handlePing(conn net.Conn, seq uint16, client *Client) {
	if client != nil && client.Token != "" {
		h.sessionService.Refresh(client.Token)
	}
	resp := &protocol.PongResp{}
	h.sendMessage(conn, seq, resp)
}

func (h *TCPHandler) handleLogin(conn net.Conn, seq uint16, req *protocol.LoginReq) *Client {
	resp := &protocol.LoginResp{}

	if req.Token != "" {
		sess, err := h.sessionService.Validate(req.Token)
		if err != nil {
			resp.Code = 401
			resp.Message = "invalid or expired token"
			h.sendMessage(conn, seq, resp)
			return nil
		}

		resp.Code = 200
		resp.Message = "login success via token"
		resp.Token = sess.Token
		resp.UserID = sess.UserID

		client := &Client{
			Conn:       conn,
			UserID:     sess.UserID,
			Token:      sess.Token,
			Username:   sess.Username,
			LastActive: time.Now(),
		}

		h.mu.Lock()
		h.clients[sess.UserID] = client
		h.mu.Unlock()

		h.sessionService.SetUserOnline(sess.UserID, sess.Token)
		h.sendMessage(conn, seq, resp)
		log.Printf("User %d logged in via token", sess.UserID)
		return client
	}

	user, token, err := h.userService.Login(&model.UserLoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		resp.Code = 401
		resp.Message = err.Error()
		h.sendMessage(conn, seq, resp)
		return nil
	}

	if err := h.sessionService.Create(user.ID, user.Username, token); err != nil {
		resp.Code = 500
		resp.Message = "failed to create session"
		h.sendMessage(conn, seq, resp)
		return nil
	}

	resp.Code = 200
	resp.Message = "login success"
	resp.Token = token
	resp.UserID = user.ID

	client := &Client{
		Conn:       conn,
		UserID:     user.ID,
		Token:      token,
		Username:   user.Username,
		LastActive: time.Now(),
	}

	h.mu.Lock()
	h.clients[user.ID] = client
	h.mu.Unlock()

	h.sessionService.SetUserOnline(user.ID, token)
	h.sendMessage(conn, seq, resp)
	log.Printf("User %d logged in", user.ID)
	return client
}

func (h *TCPHandler) handleRegister(conn net.Conn, seq uint16, req *protocol.RegisterReq) {
	resp := &protocol.RegisterResp{}

	user, err := h.userService.Register(&model.UserRegisterRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		resp.Code = 400
		resp.Message = err.Error()
		h.sendMessage(conn, seq, resp)
		return
	}

	resp.Code = 200
	resp.Message = "register success"
	resp.UserID = user.ID

	h.sendMessage(conn, seq, resp)
	log.Printf("User %d registered", user.ID)
}

func (h *TCPHandler) sendMessage(conn net.Conn, seq uint16, msg protocol.Message) {
	data, err := h.codec.Encode(msg, seq)
	if err != nil {
		log.Printf("Encode error: %v", err)
		return
	}

	if _, err := conn.Write(data); err != nil {
		log.Printf("Write error: %v", err)
	}
}

func (h *TCPHandler) sendError(conn net.Conn, seq uint16, code int, message string) {
	resp := &protocol.ErrorResp{
		Code:    code,
		Message: message,
	}
	h.sendMessage(conn, seq, resp)
}

func (h *TCPHandler) nextSeq() uint16 {
	return uint16(atomic.AddUint64(&h.seqCounter, 1))
}

func (h *TCPHandler) RemoveClient(userID int64) {
	h.mu.Lock()
	delete(h.clients, userID)
	h.mu.Unlock()
}

func (h *TCPHandler) GetClient(userID int64) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[userID]
}

func (h *TCPHandler) GetOnlineUsers() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]int64, 0, len(h.clients))
	for id := range h.clients {
		users = append(users, id)
	}
	return users
}

func (h *TCPHandler) handleCreateRoom(conn net.Conn, seq uint16, client *Client, req *protocol.CreateRoomReq) {
	resp := &protocol.CreateRoomResp{}

	if client.RoomID != 0 {
		resp.Code = 400
		resp.Message = "already in a room, please leave first"
		h.sendMessage(conn, seq, resp)
		return
	}

	roomName := req.RoomName
	if roomName == "" {
		roomName = client.Username + "'s room"
	}

	room, err := h.roomService.CreateRoom(roomName, client.UserID)
	if err != nil {
		resp.Code = 500
		resp.Message = err.Error()
		h.sendMessage(conn, seq, resp)
		return
	}

	client.RoomID = room.ID

	resp.Code = 200
	resp.Message = "room created"
	resp.RoomID = room.ID

	h.sendMessage(conn, seq, resp)
	log.Printf("User %d created room %d", client.UserID, room.ID)
}

func (h *TCPHandler) handleJoinRoom(conn net.Conn, seq uint16, client *Client, req *protocol.JoinRoomReq) {
	resp := &protocol.JoinRoomResp{}

	if client.RoomID != 0 {
		resp.Code = 400
		resp.Message = "already in a room, please leave first"
		h.sendMessage(conn, seq, resp)
		return
	}

	room, err := h.roomService.GetRoom(req.RoomID)
	if err != nil {
		resp.Code = 404
		resp.Message = err.Error()
		h.sendMessage(conn, seq, resp)
		return
	}

	if err := h.roomService.JoinRoom(req.RoomID, client.UserID); err != nil {
		resp.Code = 400
		resp.Message = err.Error()
		h.sendMessage(conn, seq, resp)
		return
	}

	client.RoomID = room.ID

	resp.Code = 200
	resp.Message = "joined room"
	resp.RoomID = room.ID

	h.sendMessage(conn, seq, resp)

	h.broadcastToRoom(room.ID, &protocol.PlayerJoin{
		RoomID:   room.ID,
		UserID:   client.UserID,
		Username: client.Username,
	}, client.UserID)

	log.Printf("User %d joined room %d", client.UserID, room.ID)

	room, _ = h.roomService.GetRoom(room.ID)
	if room != nil && room.IsFull() {
		h.startGame(room)
	}
}

func (h *TCPHandler) handleLeaveRoom(conn net.Conn, seq uint16, client *Client, req *protocol.LeaveRoomReq) {
	resp := &protocol.LeaveRoomResp{}

	if client.RoomID == 0 {
		resp.Code = 400
		resp.Message = "not in any room"
		h.sendMessage(conn, seq, resp)
		return
	}

	roomID := client.RoomID

	_, err := h.roomService.GetRoom(roomID)
	if err != nil {
		client.RoomID = 0
		resp.Code = 200
		resp.Message = "left room"
		h.sendMessage(conn, seq, resp)
		return
	}

	h.broadcastToRoom(roomID, &protocol.PlayerLeave{
		RoomID: roomID,
		UserID: client.UserID,
		Reason: "player left",
	}, 0)

	if err := h.roomService.LeaveRoom(roomID, client.UserID); err != nil {
		resp.Code = 400
		resp.Message = err.Error()
		h.sendMessage(conn, seq, resp)
		return
	}

	client.RoomID = 0

	resp.Code = 200
	resp.Message = "left room"

	h.sendMessage(conn, seq, resp)
	log.Printf("User %d left room %d", client.UserID, roomID)
}

func (h *TCPHandler) handleRoomList(conn net.Conn, seq uint16, client *Client, req *protocol.RoomListReq) {
	resp := &protocol.RoomListResp{}

	rooms := h.roomService.ListWaitingRooms()

	roomInfos := make([]*protocol.RoomInfo, 0, len(rooms))
	for _, room := range rooms {
		roomInfos = append(roomInfos, &protocol.RoomInfo{
			RoomID:    room.ID,
			RoomName:  room.Name,
			Players:   room.Players,
			CreatorID: room.CreatorID,
			Status:    int(room.Status),
		})
	}

	resp.Code = 200
	resp.Message = "success"
	resp.Rooms = roomInfos

	h.sendMessage(conn, seq, resp)
}

func (h *TCPHandler) broadcastToRoom(roomID int64, msg protocol.Message, excludeUserID int64) {
	players, err := h.roomService.GetRoomPlayers(roomID)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, playerID := range players {
		if playerID == excludeUserID {
			continue
		}
		if client, ok := h.clients[playerID]; ok {
			h.sendMessage(client.Conn, h.nextSeq(), msg)
		}
	}
}

func (h *TCPHandler) broadcastToAll(msg protocol.Message, excludeUserID int64) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for userID, client := range h.clients {
		if userID == excludeUserID {
			continue
		}
		h.sendMessage(client.Conn, h.nextSeq(), msg)
	}
}

func (h *TCPHandler) handleDisconnect(client *Client) {
	if client.RoomID == 0 {
		return
	}

	roomID := client.RoomID

	game, _ := h.gameService.GetGame(roomID)
	if game != nil && !game.IsFinished() {
		winner, _ := h.gameService.Forfeit(roomID, client.UserID)
		if winner != 0 {
			h.broadcastToRoom(roomID, &protocol.GameOver{
				RoomID: roomID,
				Winner: winner,
			}, 0)
		}
		h.gameService.EndGame(roomID)
	}

	h.broadcastToRoom(roomID, &protocol.PlayerLeave{
		RoomID: roomID,
		UserID: client.UserID,
		Reason: "player disconnected",
	}, 0)

	h.roomService.LeaveRoom(roomID, client.UserID)
	client.RoomID = 0

	log.Printf("User %d disconnected from room %d", client.UserID, roomID)
}

func (h *TCPHandler) startGame(room *model.Room) {
	game, err := h.gameService.StartGame(room.ID, room.Players)
	if err != nil {
		log.Printf("Failed to start game: %v", err)
		return
	}

	h.roomService.SetRoomStatus(room.ID, model.RoomStatusPlaying)

	gameStart := &protocol.GameStart{
		RoomID:      room.ID,
		Players:     room.Players,
		FirstPlayer: game.CurrentPlayer(),
	}

	h.broadcastToRoom(room.ID, gameStart, 0)
	log.Printf("Game started in room %d, first player: %d", room.ID, game.CurrentPlayer())
}

func (h *TCPHandler) handleMove(conn net.Conn, seq uint16, client *Client, req *protocol.MoveReq) {
	resp := &protocol.MoveResp{}

	if client.RoomID == 0 {
		resp.Code = 400
		resp.Message = "not in any room"
		h.sendMessage(conn, seq, resp)
		return
	}

	roomID := client.RoomID

	game, err := h.gameService.GetGame(roomID)
	if err != nil {
		resp.Code = 404
		resp.Message = "game not found"
		h.sendMessage(conn, seq, resp)
		return
	}

	if game.IsFinished() {
		resp.Code = 400
		resp.Message = "game already finished"
		h.sendMessage(conn, seq, resp)
		return
	}

	if err := h.gameService.MakeMove(roomID, client.UserID, req.X, req.Y); err != nil {
		resp.Code = 400
		resp.Message = err.Error()
		h.sendMessage(conn, seq, resp)
		return
	}

	resp.Code = 200
	resp.Message = "move success"
	resp.X = req.X
	resp.Y = req.Y
	resp.Player = client.UserID

	h.sendMessage(conn, seq, resp)

	game, _ = h.gameService.GetGame(roomID)
	if game == nil {
		return
	}

	boardUpdate := &protocol.BoardUpdate{
		RoomID:        roomID,
		Board:         game.GetBoardCopy(),
		LastX:         req.X,
		LastY:         req.Y,
		LastPlayer:    client.UserID,
		CurrentPlayer: game.CurrentPlayer(),
	}

	h.broadcastToRoom(roomID, boardUpdate, 0)

	if game.IsFinished() {
		gameOver := &protocol.GameOver{
			RoomID:  roomID,
			Winner:  game.Winner,
			WinLine: game.WinLine,
		}
		h.broadcastToRoom(roomID, gameOver, 0)
		h.gameService.EndGame(roomID)
		h.roomService.SetRoomStatus(roomID, model.RoomStatusFinished)

		h.updateGameResult(roomID, game.Players, game.Winner)

		log.Printf("Game finished in room %d, winner: %d", roomID, game.Winner)
	}
}

func (h *TCPHandler) handleForfeit(conn net.Conn, seq uint16, client *Client, req *protocol.ForfeitReq) {
	resp := &protocol.ForfeitResp{}

	if client.RoomID == 0 {
		resp.Code = 400
		resp.Message = "not in any room"
		h.sendMessage(conn, seq, resp)
		return
	}

	roomID := client.RoomID

	winner, err := h.gameService.Forfeit(roomID, client.UserID)
	if err != nil {
		resp.Code = 400
		resp.Message = err.Error()
		h.sendMessage(conn, seq, resp)
		return
	}

	resp.Code = 200
	resp.Message = "forfeit success"
	resp.Winner = winner

	h.sendMessage(conn, seq, resp)

	gameOver := &protocol.GameOver{
		RoomID: roomID,
		Winner: winner,
	}
	h.broadcastToRoom(roomID, gameOver, 0)

	h.gameService.EndGame(roomID)
	h.roomService.SetRoomStatus(roomID, model.RoomStatusFinished)

	game, _ := h.gameService.GetGame(roomID)
	if game != nil {
		h.updateGameResult(roomID, game.Players, winner)
	} else {
		players, _ := h.roomService.GetRoomPlayers(roomID)
		h.updateGameResult(roomID, players, winner)
	}

	log.Printf("User %d forfeited, winner: %d in room %d", client.UserID, winner, roomID)
}

func (h *TCPHandler) handleLeaderboard(conn net.Conn, seq uint16, client *Client, req *protocol.LeaderboardReq) {
	resp := &protocol.LeaderboardResp{}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	entries, err := h.rankService.GetLeaderboard(limit, req.Offset)
	if err != nil {
		resp.Code = 500
		resp.Message = err.Error()
		h.sendMessage(conn, seq, resp)
		return
	}

	ranks := make([]*protocol.RankEntry, 0, len(entries))
	for _, e := range entries {
		ranks = append(ranks, &protocol.RankEntry{
			UserID:    e.UserID,
			Username:  e.Username,
			Score:     e.Score,
			WinCount:  e.WinCount,
			LoseCount: e.LoseCount,
			WinRate:   e.WinRate,
			Rank:      e.Rank,
		})
	}

	resp.Code = 200
	resp.Message = "success"
	resp.Ranks = ranks

	h.sendMessage(conn, seq, resp)
}

func (h *TCPHandler) handleUserStats(conn net.Conn, seq uint16, client *Client, req *protocol.UserStatsReq) {
	resp := &protocol.UserStatsResp{}

	userID := req.UserID
	if userID == 0 {
		userID = client.UserID
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		resp.Code = 404
		resp.Message = "user not found"
		h.sendMessage(conn, seq, resp)
		return
	}

	rank, _ := h.rankService.GetUserRank(userID)
	score, winCount, loseCount, _ := h.rankService.GetUserStats(userID)

	winRate := "0.0%"
	total := winCount + loseCount
	if total > 0 {
		winRateVal := float64(winCount) / float64(total) * 100
		winRate = fmt.Sprintf("%.1f%%", winRateVal)
	}

	resp.Code = 200
	resp.Message = "success"
	resp.UserID = user.ID
	resp.Username = user.Username
	resp.Score = score
	resp.WinCount = winCount
	resp.LoseCount = loseCount
	resp.WinRate = winRate
	resp.Rank = rank

	h.sendMessage(conn, seq, resp)
}

func (h *TCPHandler) updateGameResult(roomID int64, players []int64, winner int64) {
	if len(players) < 2 {
		return
	}

	loser := int64(0)
	for _, p := range players {
		if p != winner {
			loser = p
			break
		}
	}

	if winner != 0 && loser != 0 {
		h.userService.UpdateScore(winner, 25, true)
		h.userService.UpdateScore(loser, -20, false)

		log.Printf("Score updated: winner %d (+25), loser %d (-20)", winner, loser)
	}
}
