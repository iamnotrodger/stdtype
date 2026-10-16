package local

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	stdtype "github.com/iamnotrodger/stdytype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsRepo(t *testing.T) {
	file := "results.json"

	t.Run("should load empty slice from empty directory", func(t *testing.T) {
		statsRepo, err := NewStatsRepo(t.TempDir(), file)
		require.NoError(t, err)

		got, err := statsRepo.ListResults()
		require.NoError(t, err)

		assert.Empty(t, got)
	})

	t.Run("should create result and save it to file", func(t *testing.T) {
		statsRepo, err := NewStatsRepo(t.TempDir(), file)
		require.NoError(t, err)

		want := &stdtype.Result{
			Timestamp: time.Unix(100, 0),
			Mode:      stdtype.WordsMode.String(),
			Amount:    30,
			Stats: stdtype.Stats{
				WPM:      50.0,
				Accuracy: 50.0,
			},
		}

		err = statsRepo.CreateResult(want)
		require.NoError(t, err)

		data, err := os.ReadFile(statsRepo.file)
		require.NoError(t, err)

		var got []stdtype.Result
		err = json.Unmarshal(data, &got)
		require.NoError(t, err)

		assert.Len(t, got, 1)
		assert.Equal(t, want, got[0])
	})

	t.Run("should create and list results in order", func(t *testing.T) {
		statsRepo, err := NewStatsRepo(t.TempDir(), file)
		require.NoError(t, err)

		want := []*stdtype.Result{
			{
				Timestamp: time.Unix(100, 0),
				Mode:      stdtype.WordsMode.String(),
				Amount:    30,
				Stats: stdtype.Stats{
					WPM:      50.0,
					Accuracy: 50.0,
				},
			},
			{
				Timestamp: time.Unix(200, 0),
				Mode:      stdtype.TimeMode.String(),
				Amount:    30,
				Stats: stdtype.Stats{
					WPM:      60.0,
					Accuracy: 100.0,
				},
			},
		}

		for _, r := range want {
			err = statsRepo.CreateResult(r)
			require.NoError(t, err)
		}

		got, err := statsRepo.ListResults()
		require.NoError(t, err)

		require.Len(t, got, len(want))
		for i := range want {
			assert.Equal(t, want[i], got[i])
		}
	})
}
