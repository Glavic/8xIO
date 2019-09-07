package app

import (
	"strconv"
	"strings"
	"time"
)

var (
	Ref Reference
)

type Reference struct {
	RootPath         string
	DBFile           string
	DB               *DBx
	IOs              []*I2Cx
	WebPort          int
	ButtonPressDelay time.Duration
}

func SetBit(n, pos byte) byte {
	n |= (1 << pos)
	return n
}

func ClearBit(n, pos byte) byte {
	var mask byte = ^(1 << pos)
	n &= mask
	return n
}

func ToggleBit(n, pos byte) byte {
	if HasBit(n, pos) {
		return ClearBit(n, pos)
	}
	return SetBit(n, pos)
}

func HasBit(n, pos byte) bool {
	return (n & (1 << pos)) > 0
}

func ConvertTo8BitBinaryString(num byte) string {
	bin := strconv.FormatInt(int64(num), 2)
	return strings.Repeat("0", 8-len(bin)) + bin
}
