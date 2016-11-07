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
	for _, IO := range Ref.IOs {
		IO.RestoreState()
	}
	Print("... successfully restored.\n")
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
	out, err := exec.Command("i2cdetect", args...).Output()
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
func (t *I2Cx) WriteDeprecated(register, value byte) error {
	args := []string{
		"-y",
		fmt.Sprintf("%d", t.Bus),
		t.AddrHex,
		fmt.Sprintf("0x%x", register),
		fmt.Sprintf("0x%x", value),
	}
	out, err := exec.Command("i2cset", args...).Output()
	if err != nil {
		return err
	}
	if string(out) != "" {
		return errors.New("Write() returned empty response?")
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
func (t *I2Cx) ReadDeprecated(register byte) (byte, error) {
	args := []string{
		"-y",
		fmt.Sprintf("%d", t.Bus),
		t.AddrHex,
		fmt.Sprintf("0x%x", register),
	}
	out, err := exec.Command("i2cget", args...).Output()
	if err != nil {
		return 0, err
	}
	hex := string(out)[0:4]
	out2, err := strconv.ParseUint(hex, 0, 8)
	if err != nil {
		return 0, err
	}
	return byte(out2), nil
}

func (t *I2Cx) SetupIC() {
	t.Write(0x00, 0x00) // IODIRA
	t.Write(0x01, 0xff) // IODIRB
	t.Write(0x03, 0xff) // IPOLB
	t.Write(0x0d, 0xff) // GPPUB
}
func (t *I2Cx) Check() {
	t.SetupIC()

	input, err := t.Read(0x13)
	if err != nil {
		return
	}

	output, err := t.Read(0x14)
	if err != nil {
		return
	}
	output_new := output

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
		if HasBit(output, bit) {
			output_new = ClearBit(output_new, bit)
		} else {
			output_new = SetBit(output_new, bit)
		}
	}
	if output != output_new {
		t.ChangeState(output_new)
		Print("Phisical | Device #%d-0x%x has new state = %s\n", t.Bus, t.Addr, ConvertTo8BitBinaryString(output_new))
	}
	t.OutputState = output_new
}
func (t *I2Cx) ChangeState(output byte) {
	t.Write(0x14, output)
	t.OutputState = output

	bucket := "DEVICE-OUTPUT"
	k := Ref.DB.PrepareValue(t.Bus, t.Addr)
	Ref.DB.Set(bucket, k, Ref.DB.PrepareValue(output))
}
func (t *I2Cx) RestoreState() {
	bucket := "DEVICE-OUTPUT"
	k := Ref.DB.PrepareValue(t.Bus, t.Addr)
	if output_string, ok := Ref.DB.Get(bucket, k); ok {
		i, err := strconv.Atoi(output_string)
		if err == nil {
			t.ChangeState(byte(i))
		}
	}
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
