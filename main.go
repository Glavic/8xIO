package main

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/Glavic/8xIO/app"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		DisableColors: false,
	})
}

func main() {
	// init Reference
	Ref = Reference{
		DBFile:  ".data.db",
		WebPort: 8080,
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.WithField("file", Ref.DBFile).Fatal(err)
	}
	dir += string(filepath.Separator)
	Ref.RootPath = dir

	// init DB
	db, err := DB(Ref.RootPath + Ref.DBFile)
	if err != nil {
		log.WithField("file", Ref.DBFile).Fatal(err)
	}
	defer db.Close()
	Ref.DB = db
	log.Info("DB set up and running")

	// setup I2C devices
	I2C()

	// setup web server
	log.WithField("port", Ref.WebPort).Info("starting web server")
	go WebStart()

	// inf. loop checking phisical switches
	Ref.ButtonPressDelay = 10 * time.Millisecond
	log.Info("infinitive loop for checking phisical switches")
	for {
		for _, IO := range Ref.IOs {
			IO.Check()
		}
		time.Sleep(Ref.ButtonPressDelay)
	}
}
