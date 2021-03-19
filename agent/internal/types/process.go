package types

type Pid uint32

func (p Pid) Uint32() uint32 {
	return uint32(p)
}
