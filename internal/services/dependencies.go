package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

type IDGenerator interface {
	Generate(prefix string) string
}

type Clock interface {
	Now() time.Time
}

type RandomIDGenerator struct{}

func (RandomIDGenerator) Generate(prefix string) string {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return fmt.Sprintf("%s-%06d", prefix, time.Now().UnixNano()%1_000_000)
	}

	return fmt.Sprintf("%s-%06d", prefix, n.Int64())
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}
