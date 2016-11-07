package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	Ref Reference
)

type Reference struct {
	RootPath string
	DBFile   string
	DB       *DBx
	IOs      []*I2Cx
	WebPort  int
}

func Print(format string, a ...interface{}) {
	fmt.Fprint(os.Stdout, "[", time.Now().String(), "] ")
	fmt.Fprintf(os.Stdout, format, a...)
}

func SetBit(n byte, pos byte) byte {
	n |= (1 << pos)
	return n
}
func ClearBit(n byte, pos byte) byte {
	var mask byte = ^(1 << pos)
	n &= mask
	return n
}
func HasBit(n byte, pos byte) bool {
	val := n & (1 << pos)
	return (val > 0)
}
func ConvertTo8BitBinaryString(num byte) string {
	bin := strconv.FormatInt(int64(num), 2)
	return strings.Repeat("0", 8-len(bin)) + bin
}
