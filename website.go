package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

func website() {
	web1 := http.NewServeMux()
	web1.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		header := "Archive #"
		content := "<div class=\"error\">404 Not found stream!</div>"
		footer := "TVHOST.CC"

		if !fileExists("./index.html") {
			fileData, _ := staticFiles.ReadFile("build/website.html")
			os.WriteFile("./index.html", fileData, 0644)
		}
		html, _ := os.ReadFile("./index.html")

		ch := strings.Replace(r.URL.Query().Get("ch"), "'", "", -1)             // переход на страницу выбора каналов
		chid := strings.Replace(r.URL.Query().Get("chid"), "'", "", -1)         // переход на коркнетный канал
		epgid := strings.Replace(r.URL.Query().Get("id"), "'", "", -1)          // переход на коркнетный канал
		pos := strToInt(strings.Replace(r.URL.Query().Get("pos"), "'", "", -1)) // позиция с которой надо воспроизводить файл

		//
		//
		//
		//
		//
		//
		//
		//
		//
		//
		//
		//
		//
		//
		// ███████╗███████╗██╗     ███████╗ ██████╗████████╗     ██████╗██╗  ██╗ █████╗ ███╗   ██╗███╗   ██╗███████╗██╗
		// ██╔════╝██╔════╝██║     ██╔════╝██╔════╝╚══██╔══╝    ██╔════╝██║  ██║██╔══██╗████╗  ██║████╗  ██║██╔════╝██║
		// ███████╗█████╗  ██║     █████╗  ██║        ██║       ██║     ███████║███████║██╔██╗ ██║██╔██╗ ██║█████╗  ██║
		// ╚════██║██╔══╝  ██║     ██╔══╝  ██║        ██║       ██║     ██╔══██║██╔══██║██║╚██╗██║██║╚██╗██║██╔══╝  ██║
		// ███████║███████╗███████╗███████╗╚██████╗   ██║       ╚██████╗██║  ██║██║  ██║██║ ╚████║██║ ╚████║███████╗███████╗
		// ╚══════╝╚══════╝╚══════╝╚══════╝ ╚═════╝   ╚═╝        ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝╚══════╝

		// переход по хешу (ch=2345923874502834)
		if ch != "" {
			row := db_fetchassoc(db_query("SELECT * FROM streams WHERE link='" + ch + "'"))
			header += row["id"]

			kill("tsplay_ch" + row["id"])

			if len(row) > 0 {

				// занимаем стрим
				db_query("UPDATE streams SET ip='" + getClientIP(r) + "', b_time=unixepoch('now') WHERE id=" + toStr(row["id"]))

				content = "\t<div class=\"logotypes\">"

				row2 := db_fetchrow(db_query("SELECT * FROM channels WHERE visible=1 ORDER BY id;"))
				keys := make([]string, 0, len(row2))
				for k := range row2 {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				for _, v := range keys {
					r := row2[v].(map[string]string)
					content += "\n\t\t<div class=\"logo-wrap\"><a href=\"/?chid=" + r["id"] + "\">"
					content += "<img rel=\"" + r["id"] + "\" src=\"/logo.png?id=" + r["id"] + "\" alt=\"" + r["chname"] + "\" width=\"210\" height=\"210\" />"
					content += "</a>"
					content += "<div class=\"logo-title\">" + r["chname"] + "</div>"
					content += "</div>"
				}
				content += "\n\t</div>"
			}
			//
			//
			//
			//
			//
			//
			//
			//
			//
			//
			//
			//
			//
			// ███████╗███████╗██╗     ███████╗ ██████╗████████╗    ██████╗ ██████╗  ██████╗  ██████╗ ██████╗  █████╗ ███╗   ███╗
			// ██╔════╝██╔════╝██║     ██╔════╝██╔════╝╚══██╔══╝    ██╔══██╗██╔══██╗██╔═══██╗██╔════╝ ██╔══██╗██╔══██╗████╗ ████║
			// ███████╗█████╗  ██║     █████╗  ██║        ██║       ██████╔╝██████╔╝██║   ██║██║  ███╗██████╔╝███████║██╔████╔██║
			// ╚════██║██╔══╝  ██║     ██╔══╝  ██║        ██║       ██╔═══╝ ██╔══██╗██║   ██║██║   ██║██╔══██╗██╔══██║██║╚██╔╝██║
			// ███████║███████╗███████╗███████╗╚██████╗   ██║       ██║     ██║  ██║╚██████╔╝╚██████╔╝██║  ██║██║  ██║██║ ╚═╝ ██║
			// ╚══════╝╚══════╝╚══════╝╚══════╝ ╚═════╝   ╚═╝       ╚═╝     ╚═╝  ╚═╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝

		} else if chid != "" {
			ch, _ := r.Cookie("ch")
			row := db_fetchassoc(db_query("SELECT * FROM streams WHERE link='" + ch.Value + "'"))
			if len(row) > 0 {

				kill("tsplay_ch" + row["id"])

				db_query("UPDATE streams SET play='', ip='" + getClientIP(r) + "', b_time=unixepoch('now') WHERE id=" + toStr(row["id"]))
				r := db_fetchassoc(db_query("SELECT * FROM channels WHERE id=" + chid))
				header = "<img src=\"/logo.png?id=" + r["id"] + "\" alt=\"" + r["chname"] + "\" class=\"logo-img\" /> " + r["chname"]
				content = ""
				rr := db_fetchrow(db_query("SELECT * FROM epg WHERE chid=" + chid + " AND s_time < unixepoch('now') ORDER BY s_time DESC"))
				if len(rr) > 0 {

					keys := make([]int, 0, len(rr))
					for k := range rr {
						keys = append(keys, strToInt(k))
					}
					sort.Ints(keys)

					//spew.Dump(keys)

					content = ""
					nday := ""
					first := "first"

					for _, v := range keys {
						rrr := rr[toStr(v)].(map[string]string)
						format := time.Unix(int64(strToInt(rrr["s_time"])), 0).Format("Mon, 02 Jan")

						if nday != format {
							if nday != "" {
								content += "\n\t\t\t\t</div>"
							}
							content += "\n\t\t\t\t<div class=\"epg\">"
							fformat := format
							if format == time.Now().Format("Mon, 02 Jan") {
								fformat = "Today"
							}
							content += "\n\t\t\t\t\t<div class=\"epg-day\">" + fformat + "</div>"
							nday = format
						}

						format = time.Unix(int64(strToInt(rrr["s_time"])), 0).Format("15:04")
						content += "\n\t\t\t\t\t<div class=\"epg-item " + first + "\"><span class=\"epg-item-time\">" + format + "</span><a href=\"/?id=" + rrr["id"] + "\" class=\"epg-item-title\">" + rrr["title"] + "</a></div>"
						first = ""

					}
					content += "\n\t\t\t\t</div>"
				} else {
					content = "<div class=\"error\">No recorded programs!</div>"
				}
			} else {
				content = "<div class=\"error\">404 Not found hash!</div>"
			}
			//
			//
			//
			//
			//
			//
			//
			//
			//
			//
			//
			//
			//
			//
			//
			//
			// ██████╗ ██╗      █████╗ ██╗   ██╗███████╗██████╗
			// ██╔══██╗██║     ██╔══██╗╚██╗ ██╔╝██╔════╝██╔══██╗
			// ██████╔╝██║     ███████║ ╚████╔╝ █████╗  ██████╔╝
			// ██╔═══╝ ██║     ██╔══██║  ╚██╔╝  ██╔══╝  ██╔══██╗
			// ██║     ███████╗██║  ██║   ██║   ███████╗██║  ██║
			// ╚═╝     ╚══════╝╚═╝  ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝
		} else if epgid != "" {
			ch, _ := r.Cookie("ch")
			row := db_fetchassoc(db_query("SELECT * FROM streams WHERE link='" + ch.Value + "' AND ip='" + getClientIP(r) + "'"))
			if len(row) > 0 {

				r1 := db_fetchassoc(db_query("SELECT * FROM epg WHERE id=" + epgid))
				rr := db_fetchassoc(db_query("SELECT * FROM channels WHERE id=" + toStr(r1["chid"])))

				if pos == 0 {
					slog("User " + getClientIP(r) + " start #" + epgid + ": `" + r1["title"] + "` (" + rr["chname"] + ")")
				}

				header = "<img src=\"/logo.png?id=" + rr["id"] + "\" alt=\"" + rr["chname"] + "\" class=\"logo-img\" /> " + rr["chname"]

				date_f := time.Unix(int64(strToInt(r1["s_time"])), 0).Format("Mon, 02 Jan 15:04")

				content = ""

				recInfo := startTsplay(row["id"], epgid, pos)
				if recInfo[1] > recInfo[0]+2 {
					t := "Recording duration #" + epgid + " is corrupted: " + toStr(recInfo[0]) + "/" + toStr(recInfo[1]) + " min"
					content += "<div class=\"player-error\">" + t + "</div>"
					slog(t, "err")
				}

				content += "<div class=\"player-date\">" + date_f + "</div>"
				content += "<div class=\"player-title\">" + r1["title"] + "</div>"

				dif := (strToInt(r1["f_time"]) + recInc) - (strToInt(r1["s_time"]) - recInc)

				content += "<div class=\"player-slider\">"
				content += "<input type=\"range\" min=\"0\" max=\"" + toStr(dif/60) + "\" value=\"0\" id=\"player-slider-input\">"
				content += "</div>"

				content += "<div class=\"player-position\">"
				content += "<span id=\"player-position-start\">00:00</span>"
				content += "<span id=\"player-position-now\">00:00:00</span>"
				content += "<span id=\"player-position-finish\">" + fmt.Sprintf("%02d:%02d", dif/3600, (dif%3600)/60) + "</span>"
				content += "</div>"

				content += "<a href=\"/?ch=" + ch.Value + "\" class=\"player-leave\">LEAVE PLAYER</a>"

			}
		}

		//
		//
		//
		//
		//
		//
		//
		//
		//
		//
		//
		//
		//
		//
		//
		//
		//

		w.Header().Set("Content-Type", "text/html")
		oo := strings.ReplaceAll(string(html), "{content}", content)
		oo = strings.ReplaceAll(oo, "{header}", header)
		oo = strings.ReplaceAll(oo, "{footer}", footer)
		if pos != 0 {
			oo = "OK"
		}

		if ch == "" && chid == "" && epgid == "" {
			oo = "Access denided. Use link from QR-code!"
		}

		_, err = w.Write([]byte(oo))
		if err != nil {
			log.Println(err)
		}
	})

	web1.HandleFunc("/logo.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Type", "image/png")
		_, err = w.Write(getLogoBin(strings.Replace(r.URL.Query().Get("id"), "'", "", -1)))
		if err != nil {
			log.Println(err)
		}
	})

	// Запуск сервера на порту
	siteport := conf("siteport")
	slog("Website was run on the :" + siteport)
	log.Fatal(http.ListenAndServe(":"+siteport, web1))
}
