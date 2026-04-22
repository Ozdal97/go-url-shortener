package hashids

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncoder_Encode_Deterministic(t *testing.T) {
	e, err := New("salty", 6)
	require.NoError(t, err)

	got1, err := e.Encode(42)
	require.NoError(t, err)
	got2, err := e.Encode(42)
	require.NoError(t, err)

	require.Equal(t, got1, got2)
	require.GreaterOrEqual(t, len(got1), 6)
}

func TestEncoder_Random_Unique(t *testing.T) {
	e, err := New("salty", 6)
	require.NoError(t, err)

	seen := make(map[string]struct{}, 200)
	for i := 0; i < 200; i++ {
		s, err := e.Random()
		require.NoError(t, err)
		_, dup := seen[s]
		require.False(t, dup, "collision after %d iters", i)
		seen[s] = struct{}{}
	}
}
