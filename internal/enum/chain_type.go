package enum

type ChainType int8

const (
	_ ChainType = iota
	CHEthereum
	CHOther
)

func (s ChainType) String() string {
	switch s {
	case CHEthereum:
		return "ethereum"
	case CHOther:
		return "other"
	default:
		return "unknown"
	}
}
