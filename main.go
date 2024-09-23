package main

import (
	"database/sql"
	"embed"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3"
)

type FileInfo struct {
	Path  string
	IsDir bool
}

var mu sync.Mutex

//go:embed build/*
var staticFiles embed.FS

var _debug_ = false
var err error
var demo bool = false // включение демо режима
var _version_ = [3]string{"2024-06-27", "18:31:10", "1.004"}
var wwwport string
var startTime time.Time
var fileDb = "db.sqlite3"
var db *sql.DB // Глобальная переменная для соединения с базой данных
var tmpDir = "/tmp/.tvhost"
var epgkey = ""
var recInc = 150 // зазор записей для воспроизведения (в секундах) по умолчанию 2.5 минуты (150 секунд)

// врменную директорию желательно примонтировать в tmpfs
// /etc/fstab: `tmpfs /tmp tmpfs defaults,noatime,mode=1777 0 0`
// потом выполнить: `mount -o remount /tmp`
var tmpLink = ""

func main() {

	startTime = time.Now()
	if len(os.Args) > 1 {
		if os.Args[1] == "asd" {
			_debug_ = true
		} else {
			hello()
		}
	}

	// if _debug_ {
	// 	go func() {
	// 		// http://192.168.1.25:6060/debug/pprof
	// 		log.Println(http.ListenAndServe(":6060", nil))
	// 	}()
	// }

	epgkey = generateRandomString(6)

	// при запуске удалим и вновь создадим временную директорию
	os.RemoveAll(tmpDir)
	chkDir(tmpDir)
	copy2tmp([]string{"astra", "astra_epg", "ffmpeg", "tsplay"})

	// ПОДКЛЮЧАЕМСЯ К БАЗЕ ДАННЫХ
	db, err = sql.Open("sqlite3", fileDb) // db.sqlite3
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// проверяем есть ли файл demo.sqlite3 в директории проекта
	fileInfo, err := os.Stat("demo.sqlite3")
	if !os.IsNotExist(err) {
		//ttime := startTime.Unix()
		demo = true
		color.Red("Yahoo!!! This is demo mode!!!")
	}

	// запуск впервые (создание нужных таблиц)
	fileInfo, err = os.Stat(fileDb)
	if os.IsNotExist(err) || fileInfo.Size() == 0 {
		db_init(1) // флаг=1 чтобы ещё сделать заполнить конфиг в БД
		systemd()
		//
		//
		//
		//
		if !fileExists("/etc/astra/license.txt") {
			os.WriteFile(tmpDir+"/start.sh", []byte(`#!/bin/sh
	
/sed -i '/^.*cesbo.com/d' /etc/hosts
echo "127.0.0.1    ls1.cesbo.com" >> /etc/hosts
echo "127.0.0.1    ls2.cesbo.com" >> /etc/hosts
echo "127.0.0.1    ls3.cesbo.com" >> /etc/hosts
`), 0777)
			exec.Command(tmpDir + "/start.sh").Run()
			os.Remove(tmpDir + "/start.sh")
		}
		//
		//
		//
		//
	} else {
		// создание новых полей
		database_new()
	}

	test()

	wwwport = conf("wwwport")

	if isValidPort("wwwport") {
		go webserver()
	}
	if isValidPort("siteport") {
		go website()
	}

	db_query("UPDATE streams SET play='', ip='', b_time=0, s_time=0")

	if _debug_ {
		go printMemUsage()
	}

	go main_stream()

	for {

		time.Sleep(1000 * time.Millisecond)
		ttime := time.Now().Unix()

		// check records goroutines..
		rows := db_fetchrow(db_query("SELECT * FROM channels"))

		// check epg every 70 sec
		if ttime%70 == 0 {
			rrow := map[string]string{}
			for k, _ := range rows {
				if row, ok := rows[k].(map[string]string); ok {
					u_time := strToInt(row["u_time"])

					if row["enable"] == "1" && u_time < int(ttime-1800) {
						//spew.Dump(row)
						rrow = row
						if u_time == 1 {
							break
						}
					}

				}
			}
			go update_epg(rrow)
		}

		for k, _ := range rows {
			if row, ok := rows[k].(map[string]string); ok {
				_ = row
				check_record(row)
			}
		}

		rows = db_fetchrow(db_query("SELECT * FROM streams WHERE enable=1"))
		// check streams goroutines..
		for k, _ := range rows {
			if row, ok := rows[k].(map[string]string); ok {
				_ = row
				check_stream(row)
			}
		}

	}

}
