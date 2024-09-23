package main

import "strings"

func database_new() {

	db_query("DROP TABLE slog")
	s := make(map[string]string)
	s["slog"] = "c_msg TEXT NOT NULL,c_group TEXT,c_time INTEGER"
	createTable(s)

	// новая колонка в таблице ecmg
	// r := db_fetchassoc(db_query("SELECT sql FROM sqlite_master WHERE name = 'ecmg';"))
	// if !strings.Contains(r["sql"], "key TEXT") {
	// 	slog("Create new column in table `ecmg`: key (custom ecmkey)")
	// 	db_query("ALTER TABLE ecmg ADD COLUMN key TEXT DEFAULT '';")
	// }

	if conf("epgimport") != "NULL" {
		db_query("DELETE FROM config WHERE key='epgimport'")
	}

	if conf("epgexport") != "NULL" {
		db_query("DELETE FROM config WHERE key='epgexport'")
	}

	udpmain := conf("udpmain")
	if udpmain == "NULL" {
		conf("udpmain", "udp://lo@239.0.10.1:1234", "Main archive stream")
	}

	lcnmain := conf("lcnmain")
	if lcnmain == "NULL" {
		conf("lcnmain", "200", "Main archive LCN")
	}

	xmltv := conf("xmltv")
	if xmltv == "NULL" {
		conf("xmltv", "./xmltv.xml", "Path to the XMLTV-file")
	}

	r := db_fetchassoc(db_query("SELECT sql FROM sqlite_master WHERE name = 'channels';"))
	if !strings.Contains(r["sql"], "xmlid TEXT NOT NULL") {
		slog("Create new column in table `channels`: xmlid (for EPG)")
		db_query("ALTER TABLE channels ADD COLUMN xmlid TEXT NOT NULL DEFAULT '';")
	}

}
