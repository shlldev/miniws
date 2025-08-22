package sockets

type buffer []byte

func (b *buffer) Zero() {
	*b = make(buffer, len(*b))
}
