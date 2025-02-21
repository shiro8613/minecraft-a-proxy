package packet

import "encoding/binary"

type HelloPacket struct {
	ProtocolVersion int32
	Hostname         string
	Port             int16
	State            int32
	basePacket
}

func (p *HelloPacket) Read(b []byte) (bool, error) {
	r, err := p.read(b)
	if err != nil {
		return false, err
	}

	if p.Packet_id != 0x00 {
		return false, nil
	}
	n, err := p.readVarint(r)
	if err != nil {
		return false, err
	}
	p.ProtocolVersion = n

	n, err = p.readVarint(r)
	if err != nil {
		return false, err
	}

	buf := make([]byte, n)
	n1, err := r.Read(buf)
	if err != nil {
		return false, err
	}
	p.Hostname = string(buf[:n1])

	buf = make([]byte, 2)
	_, err = r.Read(buf)
	if err != nil {
		return false, err
	}
	p.Port = int16(binary.BigEndian.Uint16(buf))

	n, err = p.readVarint(r)
	if err != nil {
		return false, err
	}
	p.State = n

	return true, nil
}
