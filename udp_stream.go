package main

import (
	"os"
	"os/exec"
	"strings"
	"time"
)

type Streams struct {
	Id        int
	StartTime int64
	LastTime  int64
}

func check_stream(row map[string]string) {

	// это удаление bgx.jpg должно происходить при отсутствии пидов tsplay_X
	// может даже в отдельной горутине
	//
	unixtime := time.Now().Unix()
	if (unixtime-int64(strToInt(row["s_time"])))%(60*15) == 0 && (row["play"] == "" || int64(strToInt(row["b_time"])+60) < unixtime) {
		t := tmpDir + "/bg" + row["id"] + ".jpg"
		os.Remove(t)
		//color.Green("Удаляем: " + t)
	}

	if !fileExists(tmpDir + "/bg" + row["id"] + ".jpg") {
		//color.Red(row["id"])
		makePictureCh(row["id"])
		chFile := tmpDir + "/" + row["id"] + "/ch.ts"
		cmd := exec.Command(tmpDir+"/bin/ffmpeg",
			"-y",
			"-f", "image2",
			"-i", tmpDir+"/"+row["id"]+"/img%d.jpg",
			"-aspect", "16:9",
			"-qscale", "1",
			"-g", "100",
			"-mpegts_service_id", "0x64",
			"-mpegts_pmt_start_pid", "0x190",
			"-mpegts_start_pid", "0x191",
			"-metadata", "service_provider=tvhost.cc",
			"-metadata", "service_name=Archive_"+row["id"],
			chFile)
		cmd.Run()
		copyFile(chFile, tmpDir+"/"+row["id"]+".ts")
		os.RemoveAll(tmpDir + "/" + row["id"])

		db_query("UPDATE streams SET link='" + tmpLink + "' WHERE id=" + row["id"])
		// if row["id"] == "4" {
		// 	color.Red("Обновлён hash: " + row["link"] + " => " + tmpLink)
		// }

		astraStart(row)
	}

}

func astraStart(row map[string]string) {
	id := row["id"]
	stopProcess("astra_ch" + id)
	lua := "make_channel({ name = 'astra_ch" + id + "',\n"
	lua += "input = {\n"
	remap := "set_pnr=" + id + "&map.pmt=" + toStr(strToInt(id)+100) + "&map.video=" + toStr(strToInt(id)+200) + "&map.audio=" + toStr(strToInt(id)+300)
	lua += "'udp://lo@" + getTmpUdp(id) + "#no_eit&" + remap + "',\n"
	lua += "'file://" + tmpDir + "/" + id + ".ts#bitrate_limit=32&loop&" + remap + "',\n"
	lua += "},\n"
	lua += "  output = {'" + row["uri"] + "'} })\n"
	//создаём файл конфига астры
	file, _ := os.Create(tmpDir + "/" + id + ".lua")
	file.WriteString(lua)
	//символическая ссылка
	os.Symlink(tmpDir+"/bin/astra", tmpDir+"/bin/astra_ch"+id)

	// запускаем..
	if _debug_ {
		slog("Restart `astra_ch"+id+"`: "+row["uri"], "debug")
	}
	db_query("UPDATE streams SET s_time=unixepoch('now') WHERE id=" + id)
	startProcess(tmpDir+"/bin/astra_ch"+id, tmpDir+"/"+id+".lua")

	delay(50)
	if !_debug_ {
		os.Remove(tmpDir + "/" + id + ".lua")
	}

}

func getTmpUdp(id string) string {
	if strToInt(id) > 255 {
		return "239.60.28." + toStr(strToInt(id)-255)
	} else {
		return "239.60.27." + id
	}
}

func startTsplay(sid string, epgid string, pos int) []int { // pos - это позиция в минутах ++

	out := make([]int, 2)

	epg := db_fetchassoc(db_query("SELECT * FROM epg WHERE id=" + epgid))
	channel := db_fetchassoc(db_query("SELECT * FROM channels WHERE id=" + epg["chid"]))

	if len(channel) > 0 {
		db_query("UPDATE streams SET play='" + epg["chid"] + ":" + epgid + "', b_time=unixepoch('now') WHERE id=" + sid)
	} else {
		db_query("UPDATE streams SET play='', b_time=0 WHERE id=" + sid)
	}
	stream := db_fetchassoc(db_query("SELECT * FROM streams WHERE id=" + sid))

	kill("tsplay_ch" + sid)
	//символическая ссылка
	os.Symlink(tmpDir+"/bin/tsplay", tmpDir+"/bin/tsplay_ch"+sid)
	if pos == 0 {
		astraStart(stream)
	}

	start := 0
	finish := 0
	needfiles := []string{}
	files, _ := getFiles(channel["folder"])
	for i := 0; i < len(files); i++ {
		fn := files[i].Name()
		f := strToInt(fn[:len(fn)-3])
		start = (strToInt(epg["s_time"]) - recInc) + pos
		finish = (strToInt(epg["f_time"]) + recInc)
		if f >= start && f < finish && files[i].Size > 1000 {
			needfiles = append(needfiles, "\""+files[i].Path+"\"")
		}
	}

	// out info
	out[0] = len(needfiles)
	out[1] = (finish - start) / 60

	play := `#!/bin/bash

FILES=(` + strings.Join(needfiles, " ") + `)
play_file() {
local file=$1
` + tmpDir + `/bin/tsplay_ch` + sid + ` "$file" -i 127.0.0.1 -udp ` + getTmpUdp(sid) + `:1234
}
for file in "${FILES[@]}"; do
play_file "$file"
done` + "\n"
	go func(sid string) {
		os.WriteFile(tmpDir+"/play.sh", []byte(play), 0777)
		err := exec.Command(tmpDir + "/play.sh").Run()
		//логично здесь сделать удаление информации об использовании стрима
		if err == nil {
			db_query("UPDATE streams SET play='' WHERE id=" + sid)
		}
	}(sid)
	delay(50)
	if !_debug_ {
		os.Remove(tmpDir + "/play.sh")
	}
	return out
}

// обновляем главный поток
func main_stream() {
	for {

		makePictureCh0()
		chFile := tmpDir + "/0/ch.ts"
		cmd := exec.Command(tmpDir+"/bin/ffmpeg",
			"-y",
			"-f", "image2",
			"-i", tmpDir+"/0/img%d.jpg",
			"-aspect", "16:9",
			"-qscale", "1",
			"-g", "100",
			"-mpegts_service_id", "0x64",
			"-mpegts_pmt_start_pid", "0x190",
			"-mpegts_start_pid", "0x191",
			"-metadata", "service_provider=tvhost.cc",
			"-metadata", "service_name=Archive_channels",
			chFile)
		cmd.Run()
		copyFile(chFile, tmpDir+"/0.ts")
		os.RemoveAll(tmpDir + "/0")

		stopProcess("astra_ch0")
		delay(100)

		udpmain := conf("udpmain")

		lua := "make_channel({ name = 'astra_ch0',\n"
		lua += "input = {\n"
		lua += "'file://" + tmpDir + "/0.ts#bitrate_limit=32&loop&set_pnr=1000&map.pmt=1100&map.video=1200',\n"
		lua += "},\n"
		lua += "  output = {'" + udpmain + "'} })\n"
		file, _ := os.Create(tmpDir + "/0.lua")
		file.WriteString(lua)
		os.Symlink(tmpDir+"/bin/astra", tmpDir+"/bin/astra_ch0")
		if _debug_ {
			slog("Restart `astra_ch0`: "+udpmain, "debug")
		}

		go func(tmpDir string) {
			startProcess(tmpDir+"/bin/astra_ch0", tmpDir+"/0.lua")
		}(tmpDir)

		delay(250)
		if _debug_ {
			os.Remove(tmpDir + "/0.lua")
		}

		// ждём минутку...
		time.Sleep(54 * time.Second) // -6 сек на работу ffmpeg-а

	}
}
