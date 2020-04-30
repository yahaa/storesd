package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringSliceEqual(t *testing.T) {
	tests := []struct {
		Input [2][]string
		Want  bool
	}{
		{
			Input: [2][]string{{"ab", "bb", "cc"}, {"bb", "ab", "cc"}},
			Want:  true,
		},
		{
			Input: [2][]string{{"ab", "bb", "cc"}, {"bb", "ab", ""}},
			Want:  false,
		},
		{
			Input: [2][]string{{"ab", "bb", "cc"}, {"bb", "ab"}},
			Want:  false,
		},
	}

	for _, c := range tests {
		assert.Equal(t, c.Want, StringSliceEqual(c.Input[0], c.Input[1]))
	}
}
