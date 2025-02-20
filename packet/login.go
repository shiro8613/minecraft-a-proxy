package packet

import "github.com/google/uuid"

type LoginPacket struct {
	Name string
	Uuid string
	basePacket
}

func (p *LoginPacket) Read(b []byte) (bool, error) {
	r, err := p.read(b)
	if err != nil {
		return false, err
	}

	if p.Packet_id != 0x00 {
		return false, nil
	}

	name_len, err := p.readVarint(r)
	if err != nil {
		return false, err
	}
	buf := make([]byte, name_len)
	n, err := r.Read(buf)
	if err != nil {
		return false, err
	}

	p.Name = string(buf[:n])

	buf = make([]byte, 16)
	n, err = r.Read(buf)
	if err != nil  {
		return false, err
	}
	uid, err := uuid.FromBytes(buf)
	if err != nil {
		return false, nil
	}
	p.Uuid = uid.String()

	return true, nil
}
