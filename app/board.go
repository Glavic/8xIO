package app

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/exp/io/i2c"
)

func BoardsInit() *Boards {
	b := Boards{}
	b.load()
	return &b
}

type Boards struct {
	List []*Board
}

func (t *Boards) load() {
	for _, bus := range []byte{0, 1, 2} {
		out, err := exec.Command("i2cdetect", "-y", fmt.Sprintf("%d", bus)).Output()
		out_string := string(out)
		if err != nil || out_string[5:9] != "0  1" {
			continue
		}
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

				dev, err := i2c.Open(&i2c.Devfs{Dev: "/dev/i2c-" + strconv.Itoa(int(bus))}, int(addr))
				if err != nil {
					Print("Device #%d-0x%x cannot be opened!\n", bus, addr)
					continue
				}

				t.List = append(t.List, &Board{
					Bus:    bus,
					Addr:   byte(addr),
					Device: dev,
				})
			}
		}
	}
}

func (t *Boards) Find(bus byte, addr byte) (bool, *Board) {
	for _, board := range t.List {
		if board.Bus == bus && board.Addr == addr {
			return true, board
		}
	}
	return false, nil
}

type Board struct {
	Bus    byte
	Addr   byte
	Device *i2c.Device
}

func (t *Board) Read(register byte) (byte, error) {
	b := []byte{0}
	if err := t.Device.ReadReg(register, b); err != nil {
		return 0, err
	}
	return b[0], nil
}

func (t *Board) Write(register, value byte) error {
	if err := t.Device.WriteReg(register, []byte{value}); err != nil {
		return err
	}
	return nil
}
