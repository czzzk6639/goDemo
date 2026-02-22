package protocol

import (
	"encoding/binary"
	"io"
)

const (
	HeaderLen  = 8
	MaxBodyLen = 65535
)

type Packet struct {
	Len     uint32
	Type    uint16
	Seq     uint16
	Payload []byte
}

func NewPacket(msgType uint16, seq uint16, payload []byte) *Packet {
	return &Packet{
		Len:     uint32(HeaderLen + len(payload)),
		Type:    msgType,
		Seq:     seq,
		Payload: payload,
	}
}

func (p *Packet) Encode() ([]byte, error) {
	buf := make([]byte, HeaderLen+len(p.Payload))
	binary.BigEndian.PutUint32(buf[0:4], p.Len)
	binary.BigEndian.PutUint16(buf[4:6], p.Type)
	binary.BigEndian.PutUint16(buf[6:8], p.Seq)
	copy(buf[HeaderLen:], p.Payload)
	return buf, nil
}

func DecodeHeader(r io.Reader) (*Packet, error) {
	header := make([]byte, HeaderLen)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	p := &Packet{
		Len:  binary.BigEndian.Uint32(header[0:4]),
		Type: binary.BigEndian.Uint16(header[4:6]),
		Seq:  binary.BigEndian.Uint16(header[6:8]),
	}

	if p.Len < HeaderLen {
		return nil, ErrInvalidPacket
	}

	return p, nil
}

func (p *Packet) DecodeBody(r io.Reader) error {
	bodyLen := p.Len - HeaderLen
	if bodyLen > 0 {
		p.Payload = make([]byte, bodyLen)
		if _, err := io.ReadFull(r, p.Payload); err != nil {
			return err
		}
	}
	return nil
}

func ReadPacket(r io.Reader) (*Packet, error) {
	p, err := DecodeHeader(r)
	if err != nil {
		return nil, err
	}
	if err := p.DecodeBody(r); err != nil {
		return nil, err
	}
	return p, nil
}
