package sockets

type buffer []byte

func (b *buffer) Zero() {
	// shoutout to The C Programming Language by Kernighan & Ritchie
	*b = make(buffer, len(*b))
}
