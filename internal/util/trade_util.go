package util

import (
	"github.com/google/uuid"
	"time"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func FormatDate(ts int64) string {
	return time.Unix(0, ts).Format("20060102") // YYYYMMDD
}
