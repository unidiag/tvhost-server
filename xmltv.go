package main

import (
	"encoding/xml"
	"io"
	"os"
	"time"
)

type TV struct {
	XMLName    xml.Name    `xml:"tv"`
	Programmes []Programme `xml:"programme"`
}

type Programme struct {
	XMLName xml.Name `xml:"programme"`
	Title   string   `xml:"title"`
	Start   string   `xml:"start,attr"`
	Stop    string   `xml:"stop,attr"`
	XmlId   string   `xml:"channel,attr"`
}

func update_xmltv(row map[string]string) {

	xmltvPath := conf("xmltv")

	if row["xmlid"] == "" || xmltvPath == "" {
		return
	}

	if fileExists(xmltvPath) {
		xmlFile, err := os.Open(xmltvPath)
		if err != nil {
			slog("Error opening XMLTV-file: "+xmltvPath, "err")
			return
		}
		defer xmlFile.Close()
		byteValue, err := io.ReadAll(xmlFile)
		if err != nil {
			slog("Error reading XMLTV-file: "+xmltvPath, "err")
			return
		}
		var tv TV
		err = xml.Unmarshal(byteValue, &tv)
		if err != nil {
			slog("Error unmarshalling XMLTV-file: "+xmltvPath, "err")
			return
		}

		cnt := 0
		ttime := time.Now().Unix()
		clear_flag := false

		for _, programme := range tv.Programmes {
			if programme.XmlId == row["xmlid"] {
				s_time := convertToUTC(programme.Start)
				f_time := convertToUTC(programme.Stop)
				title := truncateString(programme.Title)
				if s_time > ttime || row["u_time"] == "1" {
					r := db_fetchassoc(db_query("SELECT * FROM epg WHERE chid=" + row["id"] + " AND s_time=" + toStr(s_time) + " AND title='" + esc(title) + "';"))
					if len(r) == 0 || row["u_time"] == "1" {
						if clear_flag == false {
							db_query("DELETE FROM epg WHERE s_time > unixepoch('now') AND chid=" + row["id"] + ";")
							clear_flag = true
						}
						db_query("INSERT INTO epg (chid, s_time, f_time, title) VALUES (" + row["id"] + ", " + toStr(s_time) + ", " + toStr(f_time) + ", '" + esc(title) + "')")
						cnt++
					}
				}
			}
		}

		if cnt > 0 {
			db_query("UPDATE channels SET u_time=unixepoch('now') WHERE id=" + row["id"])
			slog("Update EPG (from XMLTV) for channel #" + row["id"] + ": " + row["chname"] + " [" + toStr(cnt) + "]")
		}

	} else if xmltvPath != "" {
		slog("Not found XMLTV-file: "+xmltvPath, "err")
	}

	return
}
