package binary

type BinaryReader struct {
	Offset     uint32
	Buffer     []uint8
	BufferSize uint32
	Pos        uint32
}

func (p *BinaryReader) read_uint8() uint8 {
	if p.Pos+1 > p.BufferSize {
		p.Pos = p.Pos + 1
		return 0
	}
	ret := p.Buffer[p.Pos]
	p.Pos = p.Pos + 1
	return ret

}

func (p *BinaryReader) read_uint16() uint16 {
	if p.Pos+2 > p.BufferSize {
		p.Pos = p.Pos + 2
		return 0
	}
	ret := uint16(p.Buffer[p.Pos])<<8 | uint16(p.Buffer[p.Pos+1])
	p.Pos = p.Pos + 2
	return ret

}

func (p *BinaryReader) read_int32() int32 {
	if p.Pos+4 > p.BufferSize {
		p.Pos = p.Pos + 4
		return 0
	}
	ret := int32(p.Buffer[p.Pos])<<24 | int32(p.Buffer[p.Pos+1])<<16 | int32(p.Buffer[p.Pos+2])<<8 | int32(p.Buffer[p.Pos+3])
	p.Pos = p.Pos + 4
	return ret
}

func (p *BinaryReader) read_uint32() uint32 {
	if p.Pos+4 > p.BufferSize {
		p.Pos = p.Pos + 4
		return 0
	}
	ret := uint32(p.Buffer[p.Pos])<<24 | uint32(p.Buffer[p.Pos+1])<<16 | uint32(p.Buffer[p.Pos+2])<<8 | uint32(p.Buffer[p.Pos+3])
	p.Pos = p.Pos + 4
	return ret

}

func (p *BinaryReader) read_uint64() uint64 {
	if p.Pos+8 > p.BufferSize {
		p.Pos = p.Pos + 8
		return 0
	}
	ret := uint64(p.Buffer[p.Pos])<<56 | uint64(p.Buffer[p.Pos+1])<<48 | uint64(p.Buffer[p.Pos+2])<<40 | uint64(p.Buffer[p.Pos+3])<<32 | uint64(p.Buffer[p.Pos+4])<<24 | uint64(p.Buffer[p.Pos+5])<<16 | uint64(p.Buffer[p.Pos+6])<<8 | uint64(p.Buffer[p.Pos+7])
	p.Pos = p.Pos + 8
	return ret

}

func (p *BinaryReader) read_int64() int64 {
	if p.Pos+8 > p.BufferSize {
		p.Pos = p.Pos + 8
		return 0
	}
	ret := int64(p.Buffer[p.Pos])<<56 | int64(p.Buffer[p.Pos+1])<<48 | int64(p.Buffer[p.Pos+2])<<40 | int64(p.Buffer[p.Pos+3])<<32 | int64(p.Buffer[p.Pos+4])<<24 | int64(p.Buffer[p.Pos+5])<<16 | int64(p.Buffer[p.Pos+6])<<8 | int64(p.Buffer[p.Pos+7])
	p.Pos = p.Pos + 8
	return ret

}

func de_hexify(nibble uint8) uint8 {
	val := nibble - '0'
	if val > 9 {
		val = 10 + nibble - 'a'
	}
	if val > 9 {
		val = 10 + nibble - 'A'
	}
	return val

}

func (p *BinaryReader) read_hex16_uint8() uint8 {
	return de_hexify(p.read_uint8())<<4 | de_hexify(p.read_uint8())

}

func (p *BinaryReader) read_hex16_uint64() uint64 {
	if p.Pos+16 > p.BufferSize {
		p.Pos = p.Pos + 16
		return 0
	}
	ret := uint64(p.read_hex16_uint8())<<56 | uint64(p.read_hex16_uint8())<<48 | uint64(p.read_hex16_uint8())<<40 | uint64(p.read_hex16_uint8())<<32 | uint64(p.read_hex16_uint8())<<24 | uint64(p.read_hex16_uint8())<<16 | uint64(p.read_hex16_uint8())<<8 | uint64(p.read_hex16_uint8())
	return ret

}

func (p *BinaryReader) read_remainder_len() uint32 {
	return p.BufferSize - p.Pos

}

func (p *BinaryReader) read_data(len uint32) []uint8 {
	if p.Pos+len > p.BufferSize {
		p.Pos = p.Pos + len
		return nil
	}
	ret := p.Buffer[p.Pos : p.Pos+len]
	p.Pos = p.Pos + len
	return ret

}

func (p *BinaryReader) test_only_current_pos() uint32 {
	return p.Pos

}
