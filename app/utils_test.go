package app

import (
	"testing"

	"gotest.tools/assert"
)

func TestSetBit(t *testing.T) {
	tests := []struct {
		n   byte
		pos byte
		out byte
	}{
		{0, 0, 1}, {0, 1, 2}, {0, 2, 4}, {0, 3, 8}, {0, 4, 16},
		{1, 0, 1}, {1, 1, 3}, {1, 2, 5}, {1, 3, 9}, {1, 4, 17},
		{3, 0, 3}, {3, 1, 3}, {3, 2, 7}, {3, 3, 11}, {3, 4, 19},
	}

	for _, test := range tests {
		out := SetBit(test.n, test.pos)
		assert.Equal(t, out, test.out)
	}
}

func TestClearBit(t *testing.T) {
	tests := []struct {
		n   byte
		pos byte
		out byte
	}{
		{0, 0, 0}, {0, 1, 0}, {0, 2, 0}, {0, 3, 0}, {0, 4, 0},
		{15, 0, 14}, {15, 1, 13}, {15, 2, 11}, {15, 3, 7}, {15, 4, 15},
		{16, 0, 16}, {16, 1, 16}, {16, 2, 16}, {16, 3, 16}, {16, 4, 0},
	}

	for _, test := range tests {
		out := ClearBit(test.n, test.pos)
		assert.Equal(t, out, test.out)
	}
}

func TestToggleBit(t *testing.T) {
	tests := []struct {
		n   byte
		pos byte
		out byte
	}{
		{0, 0, 1}, {0, 1, 2}, {0, 2, 4}, {0, 3, 8}, {0, 4, 16},
		{15, 0, 14}, {15, 1, 13}, {15, 2, 11}, {15, 3, 7}, {15, 4, 31},
	}

	for _, test := range tests {
		out := ToggleBit(test.n, test.pos)
		assert.Equal(t, out, test.out)
	}
}

func TestHasBit(t *testing.T) {
	tests := []struct {
		n   byte
		pos byte
		out bool
	}{
		{0, 0, false}, {0, 1, false}, {0, 2, false}, {0, 3, false}, {0, 4, false},
		{15, 0, true}, {15, 1, true}, {15, 2, true}, {15, 3, true}, {15, 4, false},
	}

	for _, test := range tests {
		out := HasBit(test.n, test.pos)
		assert.Equal(t, out, test.out)
	}
}

func TestConvertTo8BitBinaryString(t *testing.T) {
	tests := []struct {
		num byte
		out string
	}{
		{0, "00000000"}, {1, "00000001"}, {2, "00000010"}, {3, "00000011"}, {4, "00000100"},
		{5, "00000101"}, {6, "00000110"}, {7, "00000111"}, {8, "00001000"}, {9, "00001001"},
		{10, "00001010"}, {11, "00001011"}, {12, "00001100"}, {13, "00001101"}, {14, "00001110"},
		{15, "00001111"}, {16, "00010000"},
		{255, "11111111"},
	}

	for _, test := range tests {
		out := ConvertTo8BitBinaryString(test.num)
		assert.Equal(t, out, test.out)
	}
}
