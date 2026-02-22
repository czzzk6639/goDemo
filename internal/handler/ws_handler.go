package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"game-server/internal/model"
	"game-server/internal/service"
	"game-server/pkg/protocol"

	"github.com/gorilla/websocket"
)

type WSHandler struct {
	userService    *service.UserService
	sessionService *service.SessionService
	roomService    *service.RoomService
	gameService    *service.GameService
	rankService    *service.RankService
	clients        map[int64]*WSClient
	mu             sync.RWMutex
}

type WSClient struct {
	Conn       *websocket.Conn
	UserID     int64
	Token      string
	Username   string
	LastActive time.Time
	RoomID     int64
}

type WSMessage struct {
	Type    uint16          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type WSResponse struct {
	Type    uint16      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewWSHandler() *WSHandler {
	return &WSHandler{
		userService:    service.NewUserService(),
		sessionService: service.NewSessionService(),
		roomService:    service.NewRoomService(),
		gameService:    service.NewGameService(),
		rankService:    service.NewRankService(),
		clients:        make(map[int64]*WSClient),
	}
}

func (h *WSHandler) HandleWS(conn *websocket.Conn) {
	defer conn.Close()

	log.Printf("New WebSocket connection from %s", conn.RemoteAddr())

	var client *WSClient

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			if client != nil {
				h.RemoveClient(client.UserID)
				h.sessionService.SetUserOffline(client.UserID)
				if client.RoomID != 0 {
					h.handleDisconnect(client)
				}
			}
			return
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(data, &wsMsg); err != nil {
			h.sendError(conn, 400, "invalid message format")
			continue
		}

		if client != nil {
			client.LastActive = time.Now()
		}

		switch wsMsg.Type {
		case protocol.TypePing:
			h.handlePing(conn, client)
		case protocol.TypeLogin:
			client = h.handleLogin(conn, wsMsg.Payload)
		case protocol.TypeRegister:
			h.handleRegister(conn, wsMsg.Payload)
		default:
			if client == nil {
				h.sendError(conn, 401, "please login first")
				continue
			}
			h.handleAuthMessage(conn, client, wsMsg.Type, wsMsg.Payload)
		}
	}
}

func (h *WSHandler) handleAuthMessage(conn *websocket.Conn, client *WSClient, msgType uint16, payload json.RawMessage) {
	switch msgType {
	case protocol.TypeCreateRoom:
		h.handleCreateRoom(conn, client, payload)
	case protocol.TypeJoinRoom:
		h.handleJoinRoom(conn, client, payload)
	case protocol.TypeLeaveRoom:
		h.handleLeaveRoom(conn, client, payload)
	case protocol.TypeRoomList:
		h.handleRoomList(conn, client, payload)
	case protocol.TypeMove:
		h.handleMove(conn, client, payload)
	case protocol.TypeForfeitReq:
		h.handleForfeit(conn, client, payload)
	case protocol.TypeLeaderboardReq:
		h.handleLeaderboard(conn, client, payload)
	case protocol.TypeUserStatsReq:
		h.handleUserStats(conn, client, payload)
	default:
		h.sendError(conn, 400, "unknown message type")
	}
}

func (h *WSHandler) handlePing(conn *websocket.Conn, client *WSClient) {
	if client != nil && client.Token != "" {
		h.sessionService.Refresh(client.Token)
	}
	h.sendMessage(conn, protocol.TypePong, &protocol.PongResp{})
}

func (h *WSHandler) handleLogin(conn *websocket.Conn, payload json.RawMessage) *WSClient {
	var req protocol.LoginReq
	if err := json.Unmarshal(payload, &req); err != nil {
		h.sendMessage(conn, protocol.TypeLoginResp, &protocol.LoginResp{Code: 400, Message: "invalid payload"})
		return nil
	}

	resp := &protocol.LoginResp{}

	if req.Token != "" {
		sess, err := h.sessionService.Validate(req.Token)
		if err != nil {
			resp.Code = 401
			resp.Message = "invalid or expired token"
			h.sendMessage(conn, protocol.TypeLoginResp, resp)
			return nil
		}

		resp.Code = 200
		resp.Message = "login success via token"
		resp.Token = sess.Token
		resp.UserID = sess.UserID

		client := &WSClient{
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
		h.sendMessage(conn, protocol.TypeLoginResp, resp)
		log.Printf("WebSocket User %d logged in via token", sess.UserID)
		return client
	}

	user, token, err := h.userService.Login(&model.UserLoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		resp.Code = 401
		resp.Message = err.Error()
		h.sendMessage(conn, protocol.TypeLoginResp, resp)
		return nil
	}

	if err := h.sessionService.Create(user.ID, user.Username, token); err != nil {
		resp.Code = 500
		resp.Message = "failed to create session"
		h.sendMessage(conn, protocol.TypeLoginResp, resp)
		return nil
	}

	resp.Code = 200
	resp.Message = "login success"
	resp.Token = token
	resp.UserID = user.ID

	client := &WSClient{
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
	h.sendMessage(conn, protocol.TypeLoginResp, resp)
	log.Printf("WebSocket User %d logged in", user.ID)
	return client
}

func (h *WSHandler) handleRegister(conn *websocket.Conn, payload json.RawMessage) {
	var req protocol.RegisterReq
	if err := json.Unmarshal(payload, &req); err != nil {
		h.sendMessage(conn, protocol.TypeRegisterResp, &protocol.RegisterResp{Code: 400, Message: "invalid payload"})
		return
	}

	resp := &protocol.RegisterResp{}

	user, err := h.userService.Register(&model.UserRegisterRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		resp.Code = 400
		resp.Message = err.Error()
		h.sendMessage(conn, protocol.TypeRegisterResp, resp)
		return
	}

	resp.Code = 200
	resp.Message = "register success"
	resp.UserID = user.ID

	h.sendMessage(conn, protocol.TypeRegisterResp, resp)
	log.Printf("WebSocket User %d registered", user.ID)
}

func (h *WSHandler) handleCreateRoom(conn *websocket.Conn, client *WSClient, payload json.RawMessage) {
	var req protocol.CreateRoomReq
	json.Unmarshal(payload, &req)

	resp := &protocol.CreateRoomResp{}

	if client.RoomID != 0 {
		resp.Code = 400
		resp.Message = "already in a room, please leave first"
		h.sendMessage(conn, protocol.TypeCreateRoomResp, resp)
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
		h.sendMessage(conn, protocol.TypeCreateRoomResp, resp)
		return
	}

	client.RoomID = room.ID

	resp.Code = 200
	resp.Message = "room created"
	resp.RoomID = room.ID

	h.sendMessage(conn, protocol.TypeCreateRoomResp, resp)
	log.Printf("WebSocket User %d created room %d", client.UserID, room.ID)
}

func (h *WSHandler) handleJoinRoom(conn *websocket.Conn, client *WSClient, payload json.RawMessage) {
	var req protocol.JoinRoomReq
	json.Unmarshal(payload, &req)

	resp := &protocol.JoinRoomResp{}

	if client.RoomID != 0 {
		resp.Code = 400
		resp.Message = "already in a room, please leave first"
		h.sendMessage(conn, protocol.TypeJoinRoomResp, resp)
		return
	}

	room, err := h.roomService.GetRoom(req.RoomID)
	if err != nil {
		resp.Code = 404
		resp.Message = err.Error()
		h.sendMessage(conn, protocol.TypeJoinRoomResp, resp)
		return
	}

	if err := h.roomService.JoinRoom(req.RoomID, client.UserID); err != nil {
		resp.Code = 400
		resp.Message = err.Error()
		h.sendMessage(conn, protocol.TypeJoinRoomResp, resp)
		return
	}

	client.RoomID = room.ID

	resp.Code = 200
	resp.Message = "joined room"
	resp.RoomID = room.ID

	h.sendMessage(conn, protocol.TypeJoinRoomResp, resp)

	h.broadcastToRoom(room.ID, &protocol.PlayerJoin{
		RoomID:   room.ID,
		UserID:   client.UserID,
		Username: client.Username,
	}, client.UserID)

	log.Printf("WebSocket User %d joined room %d", client.UserID, room.ID)

	room, _ = h.roomService.GetRoom(room.ID)
	if room != nil && room.IsFull() {
		h.startGame(room)
	}
}

func (h *WSHandler) handleLeaveRoom(conn *websocket.Conn, client *WSClient, payload json.RawMessage) {
	resp := &protocol.LeaveRoomResp{}

	if client.RoomID == 0 {
		resp.Code = 400
		resp.Message = "not in any room"
		h.sendMessage(conn, protocol.TypeLeaveRoomResp, resp)
		return
	}

	roomID := client.RoomID

	_, err := h.roomService.GetRoom(roomID)
	if err != nil {
		client.RoomID = 0
		resp.Code = 200
		resp.Message = "left room"
		h.sendMessage(conn, protocol.TypeLeaveRoomResp, resp)
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
		h.sendMessage(conn, protocol.TypeLeaveRoomResp, resp)
		return
	}

	client.RoomID = 0

	resp.Code = 200
	resp.Message = "left room"

	h.sendMessage(conn, protocol.TypeLeaveRoomResp, resp)
	log.Printf("WebSocket User %d left room %d", client.UserID, roomID)
}

func (h *WSHandler) handleRoomList(conn *websocket.Conn, client *WSClient, payload json.RawMessage) {
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

	h.sendMessage(conn, protocol.TypeRoomListResp, resp)
}

func (h *WSHandler) handleMove(conn *websocket.Conn, client *WSClient, payload json.RawMessage) {
	var req protocol.MoveReq
	json.Unmarshal(payload, &req)

	resp := &protocol.MoveResp{}

	if client.RoomID == 0 {
		resp.Code = 400
		resp.Message = "not in any room"
		h.sendMessage(conn, protocol.TypeMoveResp, resp)
		return
	}

	roomID := client.RoomID

	game, err := h.gameService.GetGame(roomID)
	if err != nil {
		resp.Code = 404
		resp.Message = "game not found"
		h.sendMessage(conn, protocol.TypeMoveResp, resp)
		return
	}

	if game.IsFinished() {
		resp.Code = 400
		resp.Message = "game already finished"
		h.sendMessage(conn, protocol.TypeMoveResp, resp)
		return
	}

	if err := h.gameService.MakeMove(roomID, client.UserID, req.X, req.Y); err != nil {
		resp.Code = 400
		resp.Message = err.Error()
		h.sendMessage(conn, protocol.TypeMoveResp, resp)
		return
	}

	resp.Code = 200
	resp.Message = "move success"
	resp.X = req.X
	resp.Y = req.Y
	resp.Player = client.UserID

	h.sendMessage(conn, protocol.TypeMoveResp, resp)

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

func (h *WSHandler) handleForfeit(conn *websocket.Conn, client *WSClient, payload json.RawMessage) {
	var req protocol.ForfeitReq
	json.Unmarshal(payload, &req)

	resp := &protocol.ForfeitResp{}

	if client.RoomID == 0 {
		resp.Code = 400
		resp.Message = "not in any room"
		h.sendMessage(conn, protocol.TypeForfeitResp, resp)
		return
	}

	roomID := client.RoomID

	winner, err := h.gameService.Forfeit(roomID, client.UserID)
	if err != nil {
		resp.Code = 400
		resp.Message = err.Error()
		h.sendMessage(conn, protocol.TypeForfeitResp, resp)
		return
	}

	resp.Code = 200
	resp.Message = "forfeit success"
	resp.Winner = winner

	h.sendMessage(conn, protocol.TypeForfeitResp, resp)

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

	log.Printf("WebSocket User %d forfeited, winner: %d in room %d", client.UserID, winner, roomID)
}

func (h *WSHandler) handleLeaderboard(conn *websocket.Conn, client *WSClient, payload json.RawMessage) {
	var req protocol.LeaderboardReq
	json.Unmarshal(payload, &req)

	resp := &protocol.LeaderboardResp{}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	entries, err := h.rankService.GetLeaderboard(limit, req.Offset)
	if err != nil {
		resp.Code = 500
		resp.Message = err.Error()
		h.sendMessage(conn, protocol.TypeLeaderboardResp, resp)
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

	h.sendMessage(conn, protocol.TypeLeaderboardResp, resp)
}

func (h *WSHandler) handleUserStats(conn *websocket.Conn, client *WSClient, payload json.RawMessage) {
	var req protocol.UserStatsReq
	json.Unmarshal(payload, &req)

	resp := &protocol.UserStatsResp{}

	userID := req.UserID
	if userID == 0 {
		userID = client.UserID
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		resp.Code = 404
		resp.Message = "user not found"
		h.sendMessage(conn, protocol.TypeUserStatsResp, resp)
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

	h.sendMessage(conn, protocol.TypeUserStatsResp, resp)
}

func (h *WSHandler) handleDisconnect(client *WSClient) {
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

	log.Printf("WebSocket User %d disconnected from room %d", client.UserID, roomID)
}

func (h *WSHandler) startGame(room *model.Room) {
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

func (h *WSHandler) broadcastToRoom(roomID int64, msg protocol.Message, excludeUserID int64) {
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
			h.sendMessage(client.Conn, msg.MessageType(), msg)
		}
	}
}

func (h *WSHandler) sendMessage(conn *websocket.Conn, msgType uint16, msg protocol.Message) {
	resp := WSResponse{
		Type:    msgType,
		Payload: msg,
	}
	conn.WriteJSON(resp)
}

func (h *WSHandler) sendError(conn *websocket.Conn, code int, message string) {
	h.sendMessage(conn, protocol.TypeError, &protocol.ErrorResp{
		Code:    code,
		Message: message,
	})
}

func (h *WSHandler) RemoveClient(userID int64) {
	h.mu.Lock()
	delete(h.clients, userID)
	h.mu.Unlock()
}

func (h *WSHandler) GetClient(userID int64) *WSClient {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[userID]
}

func (h *WSHandler) updateGameResult(roomID int64, players []int64, winner int64) {
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
