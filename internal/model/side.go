package model

type Side string

const (
	Buy  Side = "1" // Buy side
	Sell Side = "2" // Sell side
)

func (s Side) IsValid() bool {
	return s == Buy || s == Sell
}
