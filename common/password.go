package common

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 64

	groupedPasswordAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	groupedPasswordGroups   = 4
	groupedPasswordGroupLen = 5
)

// GenerateGroupedPassword returns xxxxx-xxxxx-xxxxx-xxxxx using lowercase
// letters and digits. The hyphens are separators only; entropy is 20 chars.
func GenerateGroupedPassword() (string, error) {
	groups := make([]string, groupedPasswordGroups)
	alphabetSize := big.NewInt(int64(len(groupedPasswordAlphabet)))
	for i := range groups {
		buf := make([]byte, groupedPasswordGroupLen)
		for j := range buf {
			n, err := rand.Int(rand.Reader, alphabetSize)
			if err != nil {
				return "", fmt.Errorf("generate grouped password: %w", err)
			}
			buf[j] = groupedPasswordAlphabet[n.Int64()]
		}
		groups[i] = string(buf)
	}
	return strings.Join(groups, "-"), nil
}
