package util

import "github.com/shopspring/decimal"

// Ascending Comparator (for Asks)
func DecimalAscComparator(a, b interface{}) int {
	d1 := a.(decimal.Decimal)
	d2 := b.(decimal.Decimal)
	return d1.Cmp(d2) // d1 < d2 → -1, d1 == d2 → 0, d1 > d2 → 1
}

// Descending Comparator (for Bids)
func DecimalDescComparator(a, b interface{}) int {
	d1 := a.(decimal.Decimal)
	d2 := b.(decimal.Decimal)
	return d2.Cmp(d1) // Reversed: d2 < d1 → -1 → means d1 > d2
}
