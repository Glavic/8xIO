package main

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/Glavic/8xIO/app"
)

func main() {
	// init Reference
	Ref = Reference{
		DBFile:  ".data.db",
		WebPort: 8080,
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		Print("FilePath error: %s\n", err)
		os.Exit(0)
	}
	dir += string(filepath.Separator)
	Ref.RootPath = dir

	// init DB
	db, err := DB(Ref.RootPath + Ref.DBFile)
	if err != nil {
		Print("DB error: %s\n", err)
		os.Exit(0)
	}
	defer db.Close()
	Ref.DB = db
	Print("DB set up and running.\n")

	// setup I2C devices
	I2C()

	// setup web server
	Print("Starting web server on port %d...\n", Ref.WebPort)
	go WebStart()

	// inf. loop checking phisical switches
	Ref.ButtonPressDelay = 10 * time.Millisecond
	Print("Infinitive loop for checking phisical switches...\n")
	for {
		for _, IO := range Ref.IOs {
			IO.Check()
		}
		time.Sleep(Ref.ButtonPressDelay)
	}
}
