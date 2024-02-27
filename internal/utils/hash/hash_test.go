package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		hash, err := Hash([]byte("test value"), []byte("some key"))
		assert.NoError(t, err)

		assert.Equal(t, hash, "3db5dd31ce6cd89fc8df4888f68a29bf5e9ec90c39b770f6a67d55ae90109b41")
	})
}

func BenchmarkHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Hash([]byte("test value"), []byte("some key"))
	}
}
