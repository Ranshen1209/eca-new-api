package common

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateGroupedPasswordFormat(t *testing.T) {
	password, err := GenerateGroupedPassword()
	require.NoError(t, err)
	parts := strings.Split(password, "-")
	require.Len(t, parts, groupedPasswordGroups)
	assert.Len(t, password, groupedPasswordGroups*groupedPasswordGroupLen+groupedPasswordGroups-1)
	for _, part := range parts {
		assert.Len(t, part, groupedPasswordGroupLen)
		for _, r := range part {
			assert.True(t, unicode.IsLower(r) || unicode.IsDigit(r), "char %q", r)
			assert.False(t, unicode.IsUpper(r))
		}
	}
}

func TestGenerateGroupedPasswordUnique(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 20; i++ {
		password, err := GenerateGroupedPassword()
		require.NoError(t, err)
		_, exists := seen[password]
		assert.False(t, exists)
		seen[password] = struct{}{}
	}
}
