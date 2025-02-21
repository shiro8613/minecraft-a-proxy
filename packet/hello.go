package packet

import (
	"encoding/binary"
	"io"
)

type HelloPacket struct {
	ProtocolVersion int32
	Hostname         string
	Port             int16
	State            int32
	basePacket
}

func (p *HelloPacket) Read(b []byte) (bool, *LoginPacket, error) {
	r, err := p.read(b)
	if err != nil {
		return false, nil, err
	}

	if p.Packet_id != 0x00 && p.Length > 2 {
		return false, nil, nil
	}
	n, err := p.readVarint(r)
	if err != nil {
		return false, nil, err
	}
	p.ProtocolVersion = n

	n, err = p.readVarint(r)
	if err != nil {
		return false, nil, err
	}

	buf := make([]byte, n)
	n1, err := r.Read(buf)
	if err != nil {
		return false, nil, err
	}
	p.Hostname = string(buf[:n1])

	buf = make([]byte, 2)
	_, err = r.Read(buf)
	if err != nil {
		return false, nil, err
	}
	p.Port = int16(binary.BigEndian.Uint16(buf))

	n, err = p.readVarint(r)
	if err != nil {
		return false, nil, err
	}
	p.State = n

	bb, err := io.ReadAll(r)
	if err != nil {
		return false, nil, err
	}

	if 0 < len(bb) {
		l := &LoginPacket{}
		ok, err := l.Read(bb)
		if err != nil {
			return false, nil, err
		}

		if ok {
			return true, l, nil
		}
	}

	return true, nil, nil
}
