package tools

func StrHash(data []byte, length uint64) uint64 {
	h := uint64(0x811C9DC5)
	for i := uint64(0); i < length; i++ {
		h = (h + uint64(data[i])) * 0x01000193
	}
	return h
}
