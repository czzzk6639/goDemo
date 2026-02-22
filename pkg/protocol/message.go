package protocol

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
	TypeRoomInfo        uint16 = 3005
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
	TypeError           uint16 = 9999
)

type Message interface {
	MessageType() uint16
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token,omitempty"`
}

func (m *LoginReq) MessageType() uint16 { return TypeLogin }

type LoginResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	UserID  int64  `json:"user_id,omitempty"`
}

func (m *LoginResp) MessageType() uint16 { return TypeLoginResp }

type RegisterReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (m *RegisterReq) MessageType() uint16 { return TypeRegister }

type RegisterResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	UserID  int64  `json:"user_id,omitempty"`
}

func (m *RegisterResp) MessageType() uint16 { return TypeRegisterResp }

type CreateRoomReq struct {
	RoomName string `json:"room_name"`
}

func (m *CreateRoomReq) MessageType() uint16 { return TypeCreateRoom }

type CreateRoomResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	RoomID  int64  `json:"room_id,omitempty"`
}

func (m *CreateRoomResp) MessageType() uint16 { return TypeCreateRoomResp }

type JoinRoomReq struct {
	RoomID int64 `json:"room_id"`
}

func (m *JoinRoomReq) MessageType() uint16 { return TypeJoinRoom }

type JoinRoomResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	RoomID  int64  `json:"room_id,omitempty"`
}

func (m *JoinRoomResp) MessageType() uint16 { return TypeJoinRoomResp }

type LeaveRoomReq struct {
	RoomID int64 `json:"room_id"`
}

func (m *LeaveRoomReq) MessageType() uint16 { return TypeLeaveRoom }

type LeaveRoomResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (m *LeaveRoomResp) MessageType() uint16 { return TypeLeaveRoomResp }

type RoomListReq struct{}

func (m *RoomListReq) MessageType() uint16 { return TypeRoomList }

type RoomListResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Rooms   []*RoomInfo `json:"rooms,omitempty"`
}

func (m *RoomListResp) MessageType() uint16 { return TypeRoomListResp }

type RoomInfo struct {
	RoomID    int64   `json:"room_id"`
	RoomName  string  `json:"room_name"`
	Players   []int64 `json:"players"`
	CreatorID int64   `json:"creator_id"`
	Status    int     `json:"status"`
}

func (m *RoomInfo) MessageType() uint16 { return TypeRoomInfo }

type PlayerJoin struct {
	RoomID   int64  `json:"room_id"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

func (m *PlayerJoin) MessageType() uint16 { return TypePlayerJoin }

type PlayerLeave struct {
	RoomID int64  `json:"room_id"`
	UserID int64  `json:"user_id"`
	Reason string `json:"reason"`
}

func (m *PlayerLeave) MessageType() uint16 { return TypePlayerLeave }

type MoveReq struct {
	RoomID int64 `json:"room_id"`
	X      int   `json:"x"`
	Y      int   `json:"y"`
}

func (m *MoveReq) MessageType() uint16 { return TypeMove }

type MoveResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Player  int64  `json:"player"`
}

func (m *MoveResp) MessageType() uint16 { return TypeMoveResp }

type GameOver struct {
	Winner  int64 `json:"winner"`
	RoomID  int64 `json:"room_id"`
	WinLine []int `json:"win_line,omitempty"`
}

func (m *GameOver) MessageType() uint16 { return TypeGameOver }

type GameStart struct {
	RoomID      int64   `json:"room_id"`
	Players     []int64 `json:"players"`
	FirstPlayer int64   `json:"first_player"`
}

func (m *GameStart) MessageType() uint16 { return TypeGameStart }

type BoardUpdate struct {
	RoomID        int64   `json:"room_id"`
	Board         [][]int `json:"board"`
	LastX         int     `json:"last_x"`
	LastY         int     `json:"last_y"`
	LastPlayer    int64   `json:"last_player"`
	CurrentPlayer int64   `json:"current_player"`
}

func (m *BoardUpdate) MessageType() uint16 { return TypeBoardUpdate }

type ForfeitReq struct {
	RoomID int64 `json:"room_id"`
}

func (m *ForfeitReq) MessageType() uint16 { return TypeForfeitReq }

type ForfeitResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Winner  int64  `json:"winner,omitempty"`
}

func (m *ForfeitResp) MessageType() uint16 { return TypeForfeitResp }

type ErrorResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (m *ErrorResp) MessageType() uint16 { return TypeError }

type PingReq struct{}

func (m *PingReq) MessageType() uint16 { return TypePing }

type PongResp struct{}

func (m *PongResp) MessageType() uint16 { return TypePong }

type LeaderboardReq struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

func (m *LeaderboardReq) MessageType() uint16 { return TypeLeaderboardReq }

type RankEntry struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Score     int    `json:"score"`
	WinCount  int    `json:"win_count"`
	LoseCount int    `json:"lose_count"`
	WinRate   string `json:"win_rate"`
	Rank      int    `json:"rank"`
}

type LeaderboardResp struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Ranks   []*RankEntry `json:"ranks,omitempty"`
}

func (m *LeaderboardResp) MessageType() uint16 { return TypeLeaderboardResp }

type UserStatsReq struct {
	UserID int64 `json:"user_id"`
}

func (m *UserStatsReq) MessageType() uint16 { return TypeUserStatsReq }

type UserStatsResp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	UserID    int64  `json:"user_id,omitempty"`
	Username  string `json:"username,omitempty"`
	Score     int    `json:"score,omitempty"`
	WinCount  int    `json:"win_count,omitempty"`
	LoseCount int    `json:"lose_count,omitempty"`
	WinRate   string `json:"win_rate,omitempty"`
	Rank      int    `json:"rank,omitempty"`
}

func (m *UserStatsResp) MessageType() uint16 { return TypeUserStatsResp }
