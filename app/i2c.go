package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/exp/io/i2c"
)

func I2C() {
	I2C_find()
	Print("Searching for I2C devices...\n")
	if len(Ref.IOs) < 1 {
		Print("... no I2C devices found :(\n")
		os.Exit(0)
	}
	Print("... found %d devices.\n", len(Ref.IOs))

	Print("Restoring I2C devices state...\n")
	restore_count := 0
	for _, IO := range Ref.IOs {
		if IO.RestoreState() {
			restore_count += 1
		}
	}
	Print("... restored %d devices.\n", restore_count)
}

type I2Cx struct {
	Bus         byte
	Addr        byte
	AddrHex     string
	Device      *i2c.Device
	InputState  [8]bool
	OutputState byte
}
type I2Cx_bits []I2Cx_bit
type I2Cx_bit struct {
	IO    *I2Cx
	Bit   byte
	State bool
	Name  string
	Ord   int
}

func I2C_find() {
	for _, bus := range []byte{0, 1, 2} {
		addresses, err := I2C_scan(bus)
		if err != nil {
			continue
		}
		for _, addr := range addresses {
			IO := I2C_create(bus, addr)
			IO.SetupIC()
			Ref.IOs = append(Ref.IOs, IO)
			Print("Device #%d-0x%x found\n", bus, addr)
		}
	}
}
func I2C_scan(bus byte) ([]byte, error) {
	args := []string{
		"-y",
		fmt.Sprintf("%d", bus),
	}
	out, err := exec.Command("/usr/sbin/i2cdetect", args...).Output()
	out_string := string(out)
	if err != nil || out_string[5:9] != "0  1" {
		return nil, errors.New("Cannot retrive any info from bus #" + strconv.Itoa(int(bus)))
	}
	addresses := []byte{}
	for _, row := range strings.Split(out_string, "\n") {
		if len(row) < 3 {
			continue
		}
		if row[2:3] != ":" {
			continue
		}
		for _, v := range strings.Split(strings.Trim(row[3:], " "), " ") {
			if v == "--" {
				continue
			}
			hex := "0x" + v
			addr, err := strconv.ParseUint(hex, 0, 8)
			if err != nil {
				continue
			}
			addresses = append(addresses, byte(addr))
		}
	}
	return addresses, nil
}
func I2C_create(bus, addr byte) *I2Cx {
	dev, err := i2c.Open(&i2c.Devfs{Dev: "/dev/i2c-" + strconv.Itoa(int(bus))}, int(addr))
	if err != nil {
		Print("Device #%d-0x%x cannot be opened?\n", bus, addr)
	}
	return &I2Cx{
		Bus:     bus,
		Addr:    addr,
		AddrHex: fmt.Sprintf("0x%x", addr),
		Device:  dev,
	}
}
func I2C_explode_to_bits() I2Cx_bits {
	var IObits I2Cx_bits
	for _, IO := range Ref.IOs {
		var bit byte
		for bit = 0; bit < 8; bit++ {
			IObit := I2Cx_bit{
				IO:    IO,
				Bit:   bit,
				State: HasBit(IO.OutputState, bit),
			}
			k := Ref.DB.PrepareValue(IO.Bus, IO.Addr, bit)
			if name, ok := Ref.DB.Get("BIT-NAME", k); ok {
				IObit.Name = name
			}
			if ord, ok := Ref.DB.Get("BIT-ORD", k); ok {
				IObit.Ord, _ = strconv.Atoi(ord)
			}
			IObits = append(IObits, IObit)
		}
	}
	return IObits
}

func (t *I2Cx) Write(register, value byte) error {
	if err := t.Device.WriteReg(register, []byte{value}); err != nil {
		return err
	}
	return nil
}
func (t *I2Cx) Read(register byte) (byte, error) {
	b := []byte{0}
	if err := t.Device.ReadReg(register, b); err != nil {
		return 0, err
	}
	return b[0], nil
}

func (t *I2Cx) SetupIC() {
	t.Write(0x00, 0x00)          // IODIRA
	t.Write(0x01, 0xff)          // IODIRB
	t.Write(0x03, 0xff)          // IPOLB
	t.Write(0x0d, 0xff)          // GPPUB
	t.Write(0x14, t.OutputState) // OLATA
}
func (t *I2Cx) Check() {
	t.SetupIC()

	input, err := t.Read(0x13)
	if err != nil {
		return
	}

	var bit byte = 0
	for bit = 0; bit < 8; bit++ {
		if !HasBit(input, bit) {
			t.InputState[bit] = false
			continue
		}
		if t.InputState[bit] {
			continue
		}
		t.InputState[bit] = true
		t.OutputState = ToggleBit(t.OutputState, bit)
		t.ChangeState()
		Print("Phisical | Device #%d-0x%x has new state = %s\n", t.Bus, t.Addr, ConvertTo8BitBinaryString(t.OutputState))
	}
}
func (t *I2Cx) ChangeState() {
	bucket := "DEVICE-OUTPUT"
	k := Ref.DB.PrepareValue(t.Bus, t.Addr)
	Ref.DB.Set(bucket, k, Ref.DB.PrepareValue(t.OutputState))
}
func (t *I2Cx) RestoreState() bool {
	bucket := "DEVICE-OUTPUT"
	k := Ref.DB.PrepareValue(t.Bus, t.Addr)
	if output_string, ok := Ref.DB.Get(bucket, k); ok {
		i, err := strconv.Atoi(output_string)
		if err == nil {
			t.OutputState = byte(i)
			t.ChangeState()
			return true
		}
	}
	return false
}

func (t I2Cx_bits) Len() int {
	return len(t)
}
func (t I2Cx_bits) Less(i, j int) bool {
	return t[i].Ord < t[j].Ord
}
func (t I2Cx_bits) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
