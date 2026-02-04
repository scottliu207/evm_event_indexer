package model

type (
	Pagination struct {
		Page uint64
		Size uint64
	}
)

func (p Pagination) Offset() uint64 {
	if p.Page == 0 {
		return 0
	}

	return (p.Page - 1) * p.Size
}

func (p Pagination) Limit() uint64 {
	return p.Size
}
