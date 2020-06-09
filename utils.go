package dl99

import (
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"time"
)

// Fisher-Yates
func shuffle(cards []Card) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var i, j int
	for i = 0; i < len(cards); i++ {
		j = rng.Intn(len(cards)-i) + i
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// 8 bytes timestamp + 16 bytes random bytes
func randomId(prefix string) string {
	now := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(now))
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf, uint64(now))
	rng.Read(buf[8:])
	return prefix + hex.EncodeToString(buf)
}
