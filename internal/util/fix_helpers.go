package util

import (
	"fmt"
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func GeneratePrefixedID(prefix string) string {
	return fmt.Sprintf("%s-%06d", prefix, rng.Intn(1000000))
}

func FormatDate(ts int64) string {
	return time.Unix(0, ts).Format("20060102") // YYYYMMDD
}
