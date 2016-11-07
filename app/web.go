package app

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{} // use default options
)

func WebStart() {
	http.HandleFunc("/", webHome)
	http.HandleFunc("/ws", webWS)
	http.HandleFunc("/state", webState)
	http.HandleFunc("/change", webChange)
	http.HandleFunc("/config", webConfig)
	http.HandleFunc("/log", webLog)
	http.HandleFunc("/static/", webStatic)
	for {
		err := http.ListenAndServe(":"+strconv.Itoa(Ref.WebPort), nil)
		if err != nil {
			Print("Web | error | StartWebServer() | %s\n", err)
		}
		time.Sleep(5 * time.Second)
	}
}

func webHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	t, err := template.ParseFiles(Ref.RootPath + "template/home.html")
	if err != nil {
		Print("Web | error | webHome() | %s\n", err)
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
		Print("Web | error | webHome() | %s\n", err)
		return
	}
}

func webWS(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		Print("WebWS | error | %s\n", err)
		return
	}
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			Print("WebWS read error | %s\n", err)
			break
		}
		Print("WebWS recv | %s\n", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			Print("WebWS | write error | %s\n", err)
			break
		}
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
		loop_count := 0
		for loop {
			for _, IO := range Ref.IOs {
				if state[IO.Addr] != IO.OutputState {
					loop = false
					break
				}
			}
			loop_count++
			time.Sleep(10 * time.Millisecond)
			// 100 loops is 1sec
			if loop_count > 500 {
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

	addr_string := r.URL.Query().Get("addr")
	addr64, err := strconv.ParseUint(addr_string, 10, 8)
	if err != nil {
		Print("Web | notice | webChange() | called with invalid 'addr' parameter = '%s'\n", addr_string)
		return
	}
	addr := byte(addr64)
	if addr < 0 || 255 < addr {
		Print("Web | notice | webChange() | called with out of bound 'addr' parameter = '%s'\n", addr_string)
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
		Print("Web | notice | webChange() | called with unknown 'addr' parameter = '%s'\n", addr_string)
		return
	}

	bit_string := r.URL.Query().Get("bit")
	bit64, err := strconv.ParseUint(bit_string, 10, 32)
	if err != nil {
		Print("Web | notice | webChange() | called with invalid 'bit' parameter = '%s'\n", bit_string)
		return
	}
	bit := byte(bit64)
	if 0 > bit || bit > 7 {
		Print("Web | notice | webChange() | called with out of bound 'bit' parameter = '%s'\n", bit_string)
		return
	}

	// set name
	if _, ok := r.Form["name"]; ok {
		name := r.URL.Query().Get("name")
		bucket := "BIT-NAME"
		k := Ref.DB.PrepareValue(IO.Bus, IO.Addr, bit)
		if name != "" {
			Ref.DB.Set(bucket, k, name)
			Print("Web | Device #%d-0x%x had bit %d named as '%s'\n", IO.Bus, IO.Addr, bit, name)
			fmt.Fprintf(w, "NAME-SET")
		} else {
			Ref.DB.Del(bucket, k)
			Print("Web | Device #%d-0x%x had bit %d name removed\n", IO.Bus, IO.Addr, bit)
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
			Print("Web | Device #%d-0x%x had bit %d order as '%s'\n", IO.Bus, IO.Addr, bit, ord)
			fmt.Fprintf(w, "ORD-SET")
		} else {
			Ref.DB.Del(bucket, k)
			Print("Web | Device #%d-0x%x had bit %d order removed\n", IO.Bus, IO.Addr, bit)
			fmt.Fprintf(w, "ORD-DEL")
		}
		return
	}

	// change state
	output, err := IO.Read(0x14)
	if err != nil {
		Print("Web | error | webChange() | %s\n", err)
		return
	}
	if HasBit(output, bit) {
		output = ClearBit(output, bit)
		fmt.Fprintf(w, "OFF")
	} else {
		output = SetBit(output, bit)
		fmt.Fprintf(w, "ON")
	}
	IO.ChangeState(output)
	Print("Web | Device #%d-0x%x has new state = %s\n", IO.Bus, IO.Addr, ConvertTo8BitBinaryString(output))
}

func webConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	t, err := template.ParseFiles(Ref.RootPath + "template/config.html")
	if err != nil {
		Print("Web | error | webConfig() | %s\n", err)
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
		Print("Web | error | webConfig() | %s\n", err)
		return
	}
}

func webLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	fname := Ref.RootPath + "run.log"
	file, err := os.Open(fname)
	if err != nil {
		Print("Web | error | webLog() | %s\n", err)
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
		Print("Web | error | webLog() | %s\n", err)
		return
	}
	fmt.Fprintf(w, "...")
	fmt.Fprintf(w, string(buf))
}

func webStatic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, Ref.RootPath+r.URL.Path[1:])
}
