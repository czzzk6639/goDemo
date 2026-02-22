package protocol

import (
	"encoding/json"
	"errors"
)

var (
	ErrInvalidPacket  = errors.New("invalid packet")
	ErrPacketTooLarge = errors.New("packet too large")
	ErrUnknownMsgType = errors.New("unknown message type")
	ErrInvalidPayload = errors.New("invalid payload")
)

type Codec struct{}

func NewCodec() *Codec {
	return &Codec{}
}

func (c *Codec) Encode(msg Message, seq uint16) ([]byte, error) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	if len(payload) > MaxBodyLen {
		return nil, ErrPacketTooLarge
	}

	p := NewPacket(msg.MessageType(), seq, payload)
	return p.Encode()
}

func (c *Codec) Decode(p *Packet) (Message, error) {
	if len(p.Payload) == 0 {
		return nil, ErrInvalidPayload
	}

	var msg Message
	switch p.Type {
	case TypePing:
		msg = &PingReq{}
	case TypePong:
		msg = &PongResp{}
	case TypeLogin:
		msg = &LoginReq{}
	case TypeLoginResp:
		msg = &LoginResp{}
	case TypeRegister:
		msg = &RegisterReq{}
	case TypeRegisterResp:
		msg = &RegisterResp{}
	case TypeCreateRoom:
		msg = &CreateRoomReq{}
	case TypeCreateRoomResp:
		msg = &CreateRoomResp{}
	case TypeJoinRoom:
		msg = &JoinRoomReq{}
	case TypeJoinRoomResp:
		msg = &JoinRoomResp{}
	case TypeLeaveRoom:
		msg = &LeaveRoomReq{}
	case TypeLeaveRoomResp:
		msg = &LeaveRoomResp{}
	case TypeRoomList:
		msg = &RoomListReq{}
	case TypeRoomListResp:
		msg = &RoomListResp{}
	case TypeRoomInfo:
		msg = &RoomInfo{}
	case TypePlayerJoin:
		msg = &PlayerJoin{}
	case TypePlayerLeave:
		msg = &PlayerLeave{}
	case TypeMove:
		msg = &MoveReq{}
	case TypeMoveResp:
		msg = &MoveResp{}
	case TypeGameOver:
		msg = &GameOver{}
	case TypeGameStart:
		msg = &GameStart{}
	case TypeBoardUpdate:
		msg = &BoardUpdate{}
	case TypeForfeitReq:
		msg = &ForfeitReq{}
	case TypeForfeitResp:
		msg = &ForfeitResp{}
	case TypeLeaderboardReq:
		msg = &LeaderboardReq{}
	case TypeLeaderboardResp:
		msg = &LeaderboardResp{}
	case TypeUserStatsReq:
		msg = &UserStatsReq{}
	case TypeUserStatsResp:
		msg = &UserStatsResp{}
	case TypeError:
		msg = &ErrorResp{}
	default:
		return nil, ErrUnknownMsgType
	}

	if err := json.Unmarshal(p.Payload, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func EncodeMessage(msg Message, seq uint16) ([]byte, error) {
	return NewCodec().Encode(msg, seq)
}

func DecodePacket(p *Packet) (Message, error) {
	return NewCodec().Decode(p)
}
