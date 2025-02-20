package packet

import (
	"bytes"
	"errors"
	"io"
)

type basePacket struct {
	Length    int32
	Packet_id int32
}

func (p *basePacket) readVarint(r io.Reader) (result int32, err error) {
	read := make([]byte, 1)
	numRead := 0

	for {
		_, err = io.ReadFull(r, read)
		if err != nil {
			return
		}

		readByte := read[0]

		value := int32(readByte & 0b01111111)
		result |= (value << (7 * numRead))

		numRead++
		if numRead > 5 {
			err = errors.New("error var int")
			return
		}

		if (readByte & 0b10000000) == 0 {
			break
		}
	}

	return result, nil
}

func (p *basePacket) readHaeder(r io.Reader) error {
	n, err := p.readVarint(r)
	if err != nil {
		return nil
	}
	p.Length = n

	n, err = p.readVarint(r)
	if err != nil {
		return nil
	}
	p.Packet_id = n

	return nil
}

func (p *basePacket) read(b []byte) (io.Reader, error) {
	r := bytes.NewReader(b)

	if err := p.readHaeder(r); err != nil {
		return nil, err
	}

	return r, nil
}
