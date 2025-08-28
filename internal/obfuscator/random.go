package obfuscator

import (
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	mathrand "math/rand"
	"time"
)

func InitRNG(seedOpt *int64, seeded bool) *mathrand.Rand {
	if seeded {
		return mathrand.New(mathrand.NewSource(*seedOpt))
	}
	var b [8]byte
	if _, err := crand.Read(b[:]); err != nil {
		return mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	}
	seed := int64(0)
	for i := 0; i < 8; i++ {
		seed = (seed << 8) | int64(b[i])
	}
	return mathrand.New(mathrand.NewSource(seed))
}

func RandIdent(r *mathrand.Rand, n int) string {
	if n < 2 {
		n = 2
	}
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var b []rune
	for i := 0; i < n; i++ {
		b = append(b, letters[r.Intn(len(letters))])
	}
	return string(b)
}

func SumSha256(b []byte) []byte {
	h := sha256.Sum256(b)
	return h[:]
}

func RandPerm(r *mathrand.Rand, a []string) {
	for i := range a {
		j := r.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}

func HexString(b []byte) string {
	return hex.EncodeToString(b)
}
