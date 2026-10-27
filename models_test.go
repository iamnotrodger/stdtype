package stdtype

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingsValidate(t *testing.T) {
	t.Parallel()

	lists := []string{"english", "english_1k"}

	t.Run("should accept default settings", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, DefaultSettings().Validate(lists))
	})

	tests := []struct {
		name    string
		modify  func(*Settings)
		wantErr error
		wantMsg string
	}{
		{
			name:    "should reject unknown mode",
			modify:  func(s *Settings) { s.Mode = "zen" },
			wantErr: ErrInvalidMode,
			wantMsg: `invalid mode "zen": must be "time" or "words"`,
		},
		{
			name:    "should reject non-positive duration",
			modify:  func(s *Settings) { s.Duration = 0 },
			wantErr: ErrInvalidDuration,
			wantMsg: "invalid duration 0: must be positive",
		},
		{
			name:    "should reject non-positive word count",
			modify:  func(s *Settings) { s.WordCount = -5 },
			wantErr: ErrInvalidWordCount,
			wantMsg: "invalid words -5: must be positive",
		},
		{
			name:    "should reject non-positive lines",
			modify:  func(s *Settings) { s.Lines = 0 },
			wantErr: ErrInvalidLines,
			wantMsg: "invalid lines 0: must be positive",
		},
		{
			name:    "should reject unknown word list",
			modify:  func(s *Settings) { s.WordList = "klingon" },
			wantErr: ErrInvalidList,
			wantMsg: `invalid list "klingon": must be one of [english english_1k]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := DefaultSettings()
			tt.modify(&s)

			err := s.Validate(lists)

			require.ErrorIs(t, err, tt.wantErr)
			assert.EqualError(t, err, tt.wantMsg)
		})
	}

	t.Run("should report every invalid field at once", func(t *testing.T) {
		t.Parallel()

		s := DefaultSettings()
		s.Duration = 0
		s.WordList = "klingon"

		err := s.Validate(lists)

		require.ErrorIs(t, err, ErrInvalidDuration)
		assert.ErrorIs(t, err, ErrInvalidList)
	})
}
