package packet

type Packet struct {
	Payload []byte
}

func (p *Packet) Encode() ([]byte, error) {
	return p.Payload, nil
}

func (p *Packet) Decode(payload []byte) error {
	p.Payload = payload
	return nil
}
