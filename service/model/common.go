package model

type (
	Pagination struct {
		Page int32
		Size int32
	}
)

func (p Pagination) Offset() int32 {
	return (p.Page - 1) * p.Size
}

func (p Pagination) Limit() int32 {
	return p.Size
}
