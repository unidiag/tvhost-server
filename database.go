package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func db_init(flag int) {
	createTableSQL := make(map[string]string)

	createTableSQL["channels"] = `chname TEXT NOT NULL,
								  uri TEXT NOT NULL,
								  enable INTEGER DEFAULT 0,
								  visible INTEGER DEFAULT 1,
								  folder TEXT NOT NULL,
								  xmlid TEXT NOT NULL DEFAULT '',
								  u_time INTEGER DEFAULT 0,
								  c_time INTEGER DEFAULT 0,
								  e_time INTEGER DEFAULT 0`

	createTableSQL["config"] = `key TEXT NOT NULL,
								value TEXT,
								descr TEXT,
								last TEXT,
								c_time INTEGER,
								e_time INTEGER`

	createTableSQL["slog"] = `c_msg TEXT NOT NULL,
							  c_group TEXT,
							  c_time INTEGER`

	createTableSQL["streams"] = `uri TEXT NOT NULL,
							  play TEXT NOT NULL,
							  link TEXT NOT NULL,
							  ip TEXT NOT NULL,
							  enable INTEGER DEFAULT 0,
							  b_time INTEGER DEFAULT 0,
							  s_time INTEGER,
							  c_time INTEGER,
							  e_time INTEGER`

	// createTableSQL["epg"] = `chid INTEGER,
	// 						s_time INTEGER,
	// 						f_time INTEGER,
	// 						title TEXT NOT NULL,
	// 						subtitle TEXT NOT NULL,
	// 						descr TEXT NOT NULL,
	// 						category TEXT NOT NULL,
	// 						c_time INTEGER`
	createTableSQL["epg"] = `chid INTEGER,
							s_time INTEGER,
							f_time INTEGER,
							title TEXT NOT NULL`

	createTable(createTableSQL)

	for i := 1; i <= 20; i++ {
		en := 1
		if i > 5 {
			en = 0
		}
		db_query("INSERT INTO streams (uri, enable, play, link, ip, s_time, c_time, e_time) VALUES ('udp://lo@239.1.1." + toStr(i) + "', " + toStr(en) + ", '???', '', '', 0, unixepoch('now'), 0)")
	}

	// при новом создании базы данных
	if flag == 1 {
		conf("wwwport", "8088", "Dashboard www-port")
		conf("site", "http://192.168.1.25", "Website address")
		conf("siteport", "80", "Website port")
		conf("pass_admin", "admin", "Password for system administrator")
		conf("pass_oper", "oper", "Password for system operator")
		conf("rectime", "72", "Recording duration (hours)")
		//conf("epgimport", "http://192.168.1.25:8088/epg", "EPG import URL")
		conf("xmltv", "./xmltv.xml", "Path to the XMLTV-file")
		conf("udpmain", "udp://lo@239.0.10.1:1234", "Main archive stream")
		conf("lcnmain", "200", "Main archive LCN")
		conf("trademark", "                                      TVHOST.CC", "Trademark")
		conf("paths", "channels,streams", "Access for OPER")
	}

}

func createTable(schema map[string]string) {
	for k, v := range schema {
		sql := "CREATE TABLE IF NOT EXISTS " + k + "(id INTEGER PRIMARY KEY AUTOINCREMENT,\n" + v + ");"
		_, err := db.Exec(sql)
		if err != nil {
			slog("Error create table '"+k+"'", "err")
		}
	}
}

func db_query(sql string) []any {
	out := []any{}
	// SELECT, INSERT,  UPDATE, DELETE
	words := strings.Fields(sql)
	if len(words) == 0 {
		return out
	}

	if strings.ToUpper(words[0]) == "SELECT" {
		rows, err := db.Query(sql)
		if err != nil {
			return out
		}
		defer rows.Close()

		for rows.Next() {
			list := []any{}
			items := map[string]string{}
			cols, _ := rows.Columns()
			for i := 0; i < len(cols); i++ {
				list = append(list, new(string))
			}

			if err := rows.Scan(list...); err != nil {
				//что-то отдаёт ошибки хм...
				if _debug_ {
					//fmt.Println("[ERROR db.scan] " + sql + " (may be is NULL in rows)")
				}
			}

			for k, v := range list {
				s, ok := v.(*string)
				if ok {
					items[cols[k]] = *s
				}
			}
			out = append(out, items)
		}

		if err := rows.Err(); err != nil {
			// не надо использовать slog (вызовет цепочку бесконечную)
			fmt.Println("[ERROR db.query] " + sql)
		}
	} else {
		var err error
		_, err = db.Exec(sql)
		if err != nil {
			// не надо использовать slog (вызовет цепочку бесконечную)
			fmt.Println("[ERROR db.exec] " + sql)
		}
	}
	return out
}

func backup_db() bool {
	ret := true
	sourceFile, err := os.Open(fileDb)
	if err != nil {
		ret = false
		slog("Open file error", "BACKUP")
	}
	defer sourceFile.Close()
	destinationFile, err := os.Create("_" + fileDb)
	if err != nil {
		ret = false
		slog("Create file error", "BACKUP")
	}
	defer destinationFile.Close()
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		ret = false
		slog("Copy file error", "BACKUP")
	}
	return ret
}

func db_fetchrow(res []any) map[string]interface{} {
	out := map[string]interface{}{}

	for k, v := range res {
		tmp, ok := v.(map[string]string)
		if ok {
			kk := strconv.Itoa(k + 1)
			out[kk] = tmp
			//spew.Dump("[>>>]", tmp)
		}
	}
	return out
}

func db_fetchassoc(res []any) map[string]string {
	out := map[string]string{}
	if len(res) > 0 {
		if result2, ok := res[0].(map[string]string); ok {
			out = result2
		}
	}
	return out
}
