package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cesbo/go-mpegts"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type Records struct {
	StartTime int64
	LastTime  int64
}

var recordsMap = make(map[string]Records)

//////////////////
// проверяет если есть процесс, то обнровляет, если отклбчен, то отключает
//////////////////

func check_record(row map[string]string) {
	if row["enable"] == "1" {
		tttime := time.Now().Unix()
		mu.Lock()
		if _, ok := recordsMap[row["id"]]; !ok {
			recordsMap[row["id"]] = Records{
				StartTime: tttime,
				LastTime:  tttime,
			}
			mu.Unlock()
			go udp_record(row)
		} else {
			mu.Unlock()
		}
	} else {
		mu.Lock()
		if _, ok := recordsMap[row["id"]]; ok {
			delete(recordsMap, row["id"])
		}
		mu.Unlock()
	}
}

/////////////////

func udp_record(row map[string]string) {

	row["folder"] = strings.TrimSuffix(row["folder"], "/")

	re := regexp.MustCompile(`^udp://([^@]*)@([0-9.]+)(?::(\d+))?$`)
	matches := re.FindStringSubmatch(row["uri"])
	if len(matches) != 4 {
		slog("Invalid address format: " + row["uri"])
	}

	ifi, err := net.InterfaceByName(matches[1])
	if matches[1] == "" || err != nil {
		ifi = nil
	}

	port, _ := strconv.Atoi(matches[3])
	if port < 100 || port > 65535 {
		port = 1234
	}

	conn, _ := openSocket4(ifi, net.ParseIP(matches[2]), port)
	//spew.Dump(ifi)
	defer conn.Close()

	ttt := time.Now().Unix()
	file := row["folder"] + "/" + toStr(ttt) + ".ts"
	buf := make([]byte, 4*1024)
	buf2 := []byte{}
	prevUnix := time.Now().Unix()
	label := true

	for {

		if label {
			ttt = time.Now().Unix()
			file = row["folder"] + "/" + toStr(ttt) + ".ts"
			buf = make([]byte, 4*1024)
			buf2 = []byte{}
			prevUnix = time.Now().Unix()
			label = false
			runtime.GC()
		}

		nowUnix := time.Now().Unix()
		n, _, _ := conn.ReadFrom(buf)

		buf2 = append(buf2, buf[:n]...)

		if nowUnix != prevUnix {
			// сначала проверим, существует ли в map нащ ID
			flag1 := true
			mu.Lock()
			if record, ok := recordsMap[row["id"]]; ok {
				record.LastTime = time.Now().Unix()
				recordsMap[row["id"]] = record
				flag1 = false
			}
			mu.Unlock()
			if flag1 {
				return
			}

			if nowUnix%10 == 0 {
				file_record(file, buf2)
				buf2 = []byte{}

				if nowUnix >= ttt+60*1 { // 1 minute
					label = true
				}
			}
			prevUnix = nowUnix
		}
	}
}

/////////

func file_record(file string, buf2 []byte) {
	// Открываем файл с флагами для создания, если его нет, и добавления в конец
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening file %s: %v", file, err)
		return
	}
	defer f.Close() // Закрыть файл при выходе из функции

	// Используем буферизованный писатель
	writer := bufio.NewWriter(f)
	_, err = writer.Write(buf2)
	if err != nil {
		log.Printf("Error writing to file %s: %v", file, err)
		return
	}

	// Обязательно сбрасываем буфер, чтобы записать все данные
	err = writer.Flush()
	if err != nil {
		log.Printf("Error flushing buffer to file %s: %v", file, err)
	}

	//проверим, если самый последний (первый) файл по времени более чем надо то удалим его
	ff, _ := getFiles(filepath.Dir(file))
	if len(ff) > strToInt(conf("rectime"))*60 {
		os.Remove(ff[0].Path)
	}
}

/////////

func udp_scan(udp string) []string {

	p := strings.Split(udp, ":")
	out := []string{}

	udpAddr := p[0]
	udpPort := p[1]

	port, _ := strconv.Atoi(udpPort)
	sdt := map[string]string{}
	conn, _ := openSocket4(nil, net.ParseIP(udpAddr), port)
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))

	var slicer mpegts.Slicer
	buffer := make([]byte, 32*1024)
	cnt := 0
	for {

		n, _, _ := conn.ReadFrom(buffer)
		for packet := slicer.Begin(buffer[:n]); packet != nil; packet = slicer.Next() {
			pid := packet.PID()
			if pid == 0x11 {
				conn.Close()
				length_prov := int(packet[24] - 1)
				sdt["udp"] = udp
				sdt["provider"] = string(packet[26 : 26+length_prov])

				sdt["servicename"] = string(packet[26+length_prov+1 : int(packet[7]+4)])
				sdt["pnr"] = strconv.Itoa(int(packet[16])<<8 | int(packet[17]))

				if packet[27+length_prov] == 0x01 {
					decoder := charmap.ISO8859_5.NewDecoder()
					as, _, _ := transform.Bytes(decoder, []byte(sdt["servicename"]))
					sdt["servicename"] = string(as)
				}
				out = append(out, udp)
				// if udp == "239.1.100.1:1234" {
				// 	fmt.Printf("%x\n", byte(sdt["servicename"][0]))
				// }
				if byte(sdt["servicename"][0]) < 0x16 {
					sdt["servicename"] = sdt["servicename"][1:]
				}
				out = append(out, sdt["servicename"])
				return out
			}
		}
		cnt++
		if cnt > 1000 {
			return out
		}
		// if err := slicer.Err(); err != nil {
		// 	fmt.Println(err)
		// }
	}

}
