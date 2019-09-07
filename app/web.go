package app

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func WebStart() {
	http.HandleFunc("/", webHome)
	http.HandleFunc("/state", webState)
	http.HandleFunc("/change", webChange)
	http.HandleFunc("/config", webConfig)
	http.HandleFunc("/log", webLog)
	http.HandleFunc("/static/", webStatic)
	for {
		err := http.ListenAndServe(":"+strconv.Itoa(Ref.WebPort), nil)
		if err != nil {
			log.WithFields(log.Fields{
				"app":  "web",
				"func": "WebStart",
			}).Error(err)
		}
		time.Sleep(5 * time.Second)
	}
}

func webHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	t, err := template.ParseFiles(Ref.RootPath + "template/home.html")
	if err != nil {
		log.WithFields(log.Fields{
			"app":  "web",
			"func": "webHome",
		}).Error(errors.Wrap(err, "failed to parse template"))
		return
	}

	var IObits I2Cx_bits
	for _, IObit := range I2C_explode_to_bits() {
		if IObit.Name != "" {
			IObits = append(IObits, IObit)
		}
	}
	sort.Sort(IObits)

	type Data struct {
		IObits []I2Cx_bit
	}
	data := Data{
		IObits: IObits,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.WithFields(log.Fields{
			"app":  "web",
			"func": "webHome",
		}).Error(errors.Wrap(err, "failed to execute"))
		return
	}
}

func webState(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// wait for change
	if _, ok := r.Form["w8"]; ok {
		state := map[byte]byte{}
		for _, IO := range Ref.IOs {
			state[IO.Addr] = IO.OutputState
		}
		loop := true
		loopCount := 0
		for loop {
			for _, IO := range Ref.IOs {
				if state[IO.Addr] != IO.OutputState {
					loop = false
					break
				}
			}
			loopCount++
			time.Sleep(10 * time.Millisecond)
			// 100 loops is 1sec
			if loopCount > 500 {
				loop = false
			}
		}
	}

	// print state
	fmt.Fprintf(w, "[")
	for i, IO := range Ref.IOs {
		fmt.Fprintf(w, "{\"addr\":%d,\"addr_hex\":\"0x%x\",\"out\":%d,\"out_bin\":\"%s\"}", IO.Addr, IO.Addr, IO.OutputState, ConvertTo8BitBinaryString(IO.OutputState))
		if i+1 != len(Ref.IOs) {
			fmt.Fprintf(w, ",")
		}
	}
	fmt.Fprintf(w, "]")
}

func webChange(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	addrString := r.URL.Query().Get("addr")
	bitString := r.URL.Query().Get("bit")
	logger := log.WithFields(log.Fields{
		"app":  "web",
		"func": "webChange",
		"addr": addrString,
		"bit":  bitString,
	})

	addr64, err := strconv.ParseUint(addrString, 10, 8)
	if err != nil {
		logger.Warn("called with invalid 'addr' parameter")
		return
	}
	addr := byte(addr64)
	if addr < 0 || 255 < addr {
		logger.Warn("called without of bound 'addr' parameter")
		return
	}

	var IO *I2Cx
	for _, _IO := range Ref.IOs {
		if _IO.Addr == addr {
			IO = _IO
			break
		}
	}
	if IO == nil {
		logger.Warn("called with unknown 'addr' parameter")
		return
	}

	bit64, err := strconv.ParseUint(bitString, 10, 32)
	if err != nil {
		logger.Warn("called with invalid 'bit' parameter")
		return
	}
	bit := byte(bit64)
	if 0 > bit || bit > 7 {
		logger.Warn("called without of bound 'bit' parameter")
		return
	}

	// set name
	if _, ok := r.Form["name"]; ok {
		name := r.URL.Query().Get("name")
		bucket := "BIT-NAME"
		k := Ref.DB.PrepareValue(IO.Bus, IO.Addr, bit)
		if name != "" {
			Ref.DB.Set(bucket, k, name)
			logger.Infof("device #%d-0x%x had bit %d named as '%s'", IO.Bus, IO.Addr, bit, name)
			fmt.Fprintf(w, "NAME-SET")
		} else {
			Ref.DB.Del(bucket, k)
			logger.Infof("device #%d-0x%x had bit %d name removed", IO.Bus, IO.Addr, bit)
			fmt.Fprintf(w, "NAME-DEL")
		}
		return
	}

	// set ord
	if _, ok := r.Form["ord"]; ok {
		ord := r.URL.Query().Get("ord")
		bucket := "BIT-ORD"
		k := Ref.DB.PrepareValue(IO.Bus, IO.Addr, bit)
		if ord != "" {
			Ref.DB.Set(bucket, k, ord)
			logger.Infof("device #%d-0x%x had bit %d order as '%s'", IO.Bus, IO.Addr, bit, ord)
			fmt.Fprintf(w, "ORD-SET")
		} else {
			Ref.DB.Del(bucket, k)
			logger.Infof("device #%d-0x%x had bit %d order removed", IO.Bus, IO.Addr, bit)
			fmt.Fprintf(w, "ORD-DEL")
		}
		return
	}

	// change state
	status := "OFF"
	if HasBit(IO.OutputState, bit) {
		IO.OutputState = ClearBit(IO.OutputState, bit)
		fmt.Fprintf(w, "OFF")
	} else {
		IO.OutputState = SetBit(IO.OutputState, bit)
		fmt.Fprintf(w, "ON")
		status = "ON"
	}
	IO.ChangeState()
	logger.WithField("ip", r.RemoteAddr).Infof("device #%d-0x%x-%d = %s (%s)", IO.Bus, IO.Addr, bit, status, ConvertTo8BitBinaryString(IO.OutputState))
}

func webConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	t, err := template.ParseFiles(Ref.RootPath + "template/config.html")
	if err != nil {
		log.WithFields(log.Fields{
			"app":  "web",
			"func": "webConfig",
		}).Error(errors.Wrap(err, "failed to parse template"))
		return
	}

	IObits := I2C_explode_to_bits()
	sort.Sort(IObits)

	type Data struct {
		IObits []I2Cx_bit
	}
	data := Data{
		IObits: IObits,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.WithFields(log.Fields{
			"app":  "web",
			"func": "webConfig",
		}).Error(errors.Wrap(err, "failed to execute"))
		return
	}
}

func webLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	fname := Ref.RootPath + "run.log"
	file, err := os.Open(fname)
	if err != nil {
		log.WithFields(log.Fields{
			"app":  "web",
			"func": "webLog",
		}).Error(errors.Wrap(err, "failed to find log file"))
		return
	}
	defer file.Close()

	var size int64 = 112 * 1000
	stat, err := os.Stat(fname)
	if stat.Size() < size {
		size = stat.Size()
	}
	buf := make([]byte, size)
	start := stat.Size() - size
	_, err = file.ReadAt(buf, start)
	if err != nil {
		log.WithFields(log.Fields{
			"app":  "web",
			"func": "webLog",
		}).Error(errors.Wrap(err, "failed to read log file"))
		return
	}
	fmt.Fprintf(w, "...")
	fmt.Fprintf(w, string(buf))
}

func webStatic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, Ref.RootPath+r.URL.Path[1:])
}
