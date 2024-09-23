package main

import (
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

func webserver() {
	http.Handle("/build/", http.FileServer(http.FS(staticFiles)))

	ext := make(map[string]string)
	ext["html"] = "text/html"
	ext["json"] = "application/json"
	ext["css"] = "text/css"
	ext["js"] = "application/javascript"
	ext["gif"] = "image/gif"
	ext["svg"] = "image/svg+xml"
	ext["png"] = "image/png"
	ext["jpg"] = "image/jpeg"
	ext["jpeg"] = "image/jpeg"
	ext["ico"] = "image/x-icon"
	ext["woff"] = "font/woff"
	ext["woff2"] = "font/woff2"
	ext["ttf"] = "font/ttf"
	ext["eot"] = "application/vnd.ms-fontobject"

	// Получаем содержимое корневой директории
	contents, err := readDirRecursively("build")
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range contents {
		if item.IsDir == false {
			ff := item.Path
			http.HandleFunc(strings.ReplaceAll(ff, "build", ""), func(w http.ResponseWriter, r *http.Request) {
				file, err := staticFiles.ReadFile(ff)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", ext[getFileExtension(ff)])
				_, err = w.Write(file)
				if err != nil {
					log.Println(err)
				}
			})
		}
	}

	// ██╗███╗   ███╗██████╗  ██████╗ ██████╗ ████████╗     ██████╗███████╗██╗   ██╗
	// ██║████╗ ████║██╔══██╗██╔═══██╗██╔══██╗╚══██╔══╝    ██╔════╝██╔════╝██║   ██║
	// ██║██╔████╔██║██████╔╝██║   ██║██████╔╝   ██║       ██║     ███████╗██║   ██║
	// ██║██║╚██╔╝██║██╔═══╝ ██║   ██║██╔══██╗   ██║       ██║     ╚════██║╚██╗ ██╔╝
	// ██║██║ ╚═╝ ██║██║     ╚██████╔╝██║  ██║   ██║       ╚██████╗███████║ ╚████╔╝
	// ╚═╝╚═╝     ╚═╝╚═╝      ╚═════╝ ╚═╝  ╚═╝   ╚═╝        ╚═════╝╚══════╝  ╚═══╝

	http.HandleFunc("/import", func(w http.ResponseWriter, r *http.Request) {
		if _debug_ {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "content-type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,PUT,PATCH,POST,DELETE")
		}

		out := "OK"

		if !is_user(r) && !_debug_ {
			out = "Not authorized!"
		} else {
			file, header, err := r.FormFile("file")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer file.Close()
			slog("Import file: " + toStr(header.Filename) + " (" + toStr(header.Size) + " bytes)")
			csvData, err := io.ReadAll(file)
			if err != nil {
				out = "Internal server error"
				slog(out, "err")
			}

			reader := csv.NewReader(strings.NewReader(toStr(csvData)))
			reader.Comma = ';'
			records, err := reader.ReadAll()
			if err != nil {
				out = "Not correct CSV-data"
				slog(out, "err")
			}

			//spew.Dump(csvData)

			v3 := false
			keys := []string{}
			for kk, vv := range records {
				if kk == 0 {
					keys = vv
					if len(vv) < 2 || vv[0] != "serial_no" {
						out = "Not correct keys CSV-data. Need: `serial_no;...;...;...`"
						slog(out, "err")
						//color.Blue("%d", len(vv))
						break
					}

					for i := 0; i < len(keys); i++ {
						if keys[i] == "name" { // ver. tvcas3
							v3 = true
							keys = []string{"serial_no", "descr", "mode", "emm_key", "access_criteria", "protect", "pair", "start", "finish", "e_time"}
							break
						}
					}
				} else {
					tvals := []string{}
					for kkk, vvv := range keys {
						if v3 && vvv == "mode" {
							tvals = append(tvals, "'1'")
						} else if v3 && vvv == "protect" && toStr(vv[kkk]) != "1" {
							tvals = append(tvals, "'0'")
						} else if v3 && vvv == "emm_key" {
							emmk, _ := hex.DecodeString(toStr(vv[kkk]))
							kemm := hex.EncodeToString(emmk)
							tvals = append(tvals, "'"+kemm+"'")
						} else if vvv == "e_time" {
							tvals = append(tvals, "unixepoch('now')")
						} else {
							tvals = append(tvals, "'"+esc(toStr(vv[kkk]))+"'")
						}
					}
					sql := "INSERT OR REPLACE INTO smartcards (" + strings.Join(keys, ", ") + ") VALUES (" + strings.Join(tvals, ",") + ")"
					db_query(sql)
					//slog(sql)
					//db_query("INSERT OR REPLACE INTO smartcards (id, serial_no, ...) VALUES (value_unique, value2, ...);")
				}
			}

			// проверим есть ли лицензия, если нет, то удалим лишние карты
			//if is_extended() != 0x42 {
			//	db_query("DELETE FROM smartcards WHERE id NOT IN (SELECT id FROM smartcards ORDER BY id LIMIT 100)")
			//}

		}
		fmt.Fprintf(w, "%s", out)
	})

	// ███████╗██╗  ██╗██████╗  ██████╗ ██████╗ ████████╗     ██████╗███████╗██╗   ██╗
	// ██╔════╝╚██╗██╔╝██╔══██╗██╔═══██╗██╔══██╗╚══██╔══╝    ██╔════╝██╔════╝██║   ██║
	// █████╗   ╚███╔╝ ██████╔╝██║   ██║██████╔╝   ██║       ██║     ███████╗██║   ██║
	// ██╔══╝   ██╔██╗ ██╔═══╝ ██║   ██║██╔══██╗   ██║       ██║     ╚════██║╚██╗ ██╔╝
	// ███████╗██╔╝ ██╗██║     ╚██████╔╝██║  ██║   ██║       ╚██████╗███████║ ╚████╔╝
	// ╚══════╝╚═╝  ╚═╝╚═╝      ╚═════╝ ╚═╝  ╚═╝   ╚═╝        ╚═════╝╚══════╝  ╚═══╝

	http.HandleFunc("/export", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", "attachment; filename=smartcards.csv")
		w.Header().Set("Content-Type", "text/csv")

		str := ""
		if !is_user(r) && !_debug_ {
			str = "Not authorized!"
		} else {
			cols := []string{"serial_no", "descr", "mode", "emm_key", "access_criteria", "protect", "pair", "start", "finish", "e_time"}
			str = strings.Join(cols, ";") + "\n"
			for _, v := range db_query("SELECT * FROM smartcards ORDER BY id") {
				tmp, ok := v.(map[string]string)
				if ok {
					for kk, vv := range cols {
						if kk > 0 {
							str += ";"
						}
						str += tmp[vv]
					}
					//spew.Dump("[>>>]", tmp)
					str += "\n"
				}
			}
			//fmt.Println(str)
		}

		_, err = w.Write([]byte(str))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	})
	//
	//
	//
	//
	//
	//
	//
	// ███████╗███████╗██████╗ ██╗   ██╗██╗ ██████╗███████╗
	// ██╔════╝██╔════╝██╔══██╗██║   ██║██║██╔════╝██╔════╝
	// ███████╗█████╗  ██████╔╝██║   ██║██║██║     █████╗
	// ╚════██║██╔══╝  ██╔══██╗╚██╗ ██╔╝██║██║     ██╔══╝
	// ███████║███████╗██║  ██║ ╚████╔╝ ██║╚██████╗███████╗
	// ╚══════╝╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚═╝ ╚═════╝╚══════╝

	http.HandleFunc("/service/", func(w http.ResponseWriter, r *http.Request) {
		o := []byte("Not authorize!")

		if is_user(r) || _debug_ {
			op := strings.Replace(r.URL.Query().Get("op"), "'", "", -1)
			id := strings.Replace(r.URL.Query().Get("id"), "'", "", -1)
			switch op {
			case "getScreenshot":
				file := ""
				tempFile := tmpDir + "/screenshot.jpg"
				os.Remove(tempFile)
				row := db_fetchassoc(db_query("SELECT * FROM channels WHERE id=" + id))
				if len(row) > 0 {
					files, _ := getFiles(row["folder"])
					if len(files) > 0 {
						sort.Slice(files, func(i, j int) bool {
							return files[i].Name() < files[j].Name()
						})
						file = files[len(files)-1].Path
					}
					cmd := exec.Command(tmpDir+"/bin/ffmpeg", "-i", file, "-frames:v", "1", "-vsync", "vfr", "-q:v", "2", tempFile)
					//cmd := exec.Command(tmpDir+"/bin/ffmpeg", "-i", file, "-vf", "select=eq(pict_type\\,I)", "-frames:v", "1", "-vsync", "vfr", "-q:v", "2", tempFile)
					//color.Red(file)
					cmd.Run()
				}
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
				w.Header().Set("Content-Type", "image/jpeg")

				fileInfo, err := os.Stat(tempFile)
				if file != "" && err == nil && fileInfo.Size() > 100 {
					o, _ = os.ReadFile(tempFile)
				} else {
					o = makePictureNotFound()
				}

			case "getBackground":
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
				w.Header().Set("Content-Type", "image/jpeg")
				bgFile := tmpDir + "/bg" + id + ".jpg"
				fileInfo, err := os.Stat(bgFile)
				if err == nil && fileInfo.Size() > 1000 {
					o, _ = os.ReadFile(bgFile)
				} else {
					o = makePictureNotFound()
				}

			default:
				o = []byte("unknown operation")
			}
		} else {

		}
		_, err = w.Write(o)
	})
	//
	//
	//
	//
	//
	//
	// ██╗      ██████╗  ██████╗  ██████╗
	// ██║     ██╔═══██╗██╔════╝ ██╔═══██╗
	// ██║     ██║   ██║██║  ███╗██║   ██║
	// ██║     ██║   ██║██║   ██║██║   ██║
	// ███████╗╚██████╔╝╚██████╔╝╚██████╔╝
	// ╚══════╝ ╚═════╝  ╚═════╝  ╚═════╝

	http.HandleFunc("/service/logo", func(w http.ResponseWriter, r *http.Request) {

		if _debug_ {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "content-type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,PUT,PATCH,POST,DELETE")
		}

		chid := strings.Replace(r.URL.Query().Get("chid"), "'", "", -1)

		o := "Not authorize!"
		if is_user(r) || _debug_ {
			file, header, err := r.FormFile("file")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer file.Close()
			slog("Add logotype #" + chid + ": " + toStr(header.Filename) + " (" + toStr(header.Size) + " bytes)")

			logoData, err := io.ReadAll(file)
			if err != nil {
				o = "Internal server error"
				slog(o, "err")
			} else {
				chkDir("./logos")
				os.WriteFile("./logos/"+chid+".png", logoData, 0644)
				o = "OK"
			}
		}
		fmt.Fprintf(w, "%s", o)
	})

	http.HandleFunc("/service/logos", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Type", "image/png")
		w.Write(getLogoBin(strings.Replace(r.URL.Query().Get("id"), "'", "", -1)))

	})

	//
	//
	//
	//
	//
	//
	//
	// ███████╗██████╗  ██████╗
	// ██╔════╝██╔══██╗██╔════╝
	// █████╗  ██████╔╝██║  ███╗
	// ██╔══╝  ██╔═══╝ ██║   ██║
	// ███████╗██║     ╚██████╔╝
	// ╚══════╝╚═╝      ╚═════╝

	http.HandleFunc("/epg", func(w http.ResponseWriter, r *http.Request) {
		key := strings.Replace(r.URL.Query().Get("k"), "'", "", -1)
		if r.Method == http.MethodPost && epgkey == key {
			cnt := 0
			chid := "0"
			clear_flag := false
			var jsonResult map[string]interface{}

			// Чтение тела запроса
			body, _ := io.ReadAll(r.Body)
			defer r.Body.Close()

			// Декодирование JSON строки в map[string]interface{}
			json.Unmarshal(body, &jsonResult)
			for k, v := range jsonResult {
				if k == "items" {

					for _, vv := range v.([]interface{}) {
						item := vv.(map[string]interface{})
						startUTInt := 0
						stopUTInt := 0
						title := ""
						//subtitle := ""
						//desc := ""
						//category := ""

						if t, ok := item["channel"].(string); ok {
							chid = t
						}

						if startUT, ok := item["start_ut"].(float64); ok {
							startUTInt = int(startUT)
						}
						if stopUT, ok := item["stop_ut"].(float64); ok {
							stopUTInt = int(stopUT)
						}
						if t, ok := item["title"].(map[string]interface{}); ok {
							for _, vvv := range t {
								title = truncateString(vvv.(string)) // укорачиваем строку
								break
							}
						}
						// if t, ok := item["subtitle"].(map[string]interface{}); ok {
						// 	for _, vvv := range t {
						// 		subtitle = vvv.(string)
						// 		break
						// 	}
						// }
						// if t, ok := item["desc"].(map[string]interface{}); ok {
						// 	for _, vvv := range t {
						// 		desc = vvv.(string)
						// 		break
						// 	}
						// }
						// if t, ok := item["category"].([]interface{}); ok {
						// 	for _, vvv := range t {
						// 		category = vvv.(string)
						// 		break
						// 	}
						// }

						if int64(startUTInt) > time.Now().Unix() {
							r := db_fetchassoc(db_query("SELECT * FROM epg WHERE chid=" + chid + " AND s_time=" + toStr(startUTInt) + " AND title='" + esc(title) + "';"))
							if len(r) == 0 {
								if clear_flag == false {
									db_query("DELETE FROM epg WHERE s_time > unixepoch('now') AND chid=" + chid + ";")
									clear_flag = true
								}
								//db_query("INSERT INTO epg (chid, s_time, f_time, title, subtitle, descr, category, c_time) VALUES (" + chid + ", " + toStr(startUTInt) + ", " + toStr(stopUTInt) + ", '" + esc(title) + "', '" + esc(subtitle) + "', '" + esc(desc) + "', '" + esc(category) + "', unixepoch('now'))")
								db_query("INSERT INTO epg (chid, s_time, f_time, title) VALUES (" + chid + ", " + toStr(startUTInt) + ", " + toStr(stopUTInt) + ", '" + esc(title) + "')")
								cnt++
							}
						}

						//fmt.Printf("%s: %d - %d %s (%s), [[%s]]  > %s\n", chid, startUTInt, stopUTInt, title, subtitle, desc, category)
					}
				}
			}
			if cnt > 0 {
				row := db_fetchassoc(db_query("SELECT * FROM channels WHERE id=" + chid))
				slog("Update EPG (from UDP) for channel #" + chid + ": " + row["chname"] + " [" + toStr(cnt) + "]")
			}
			kill("astra_epg")
		}
		w.Write([]byte{})

	})

	// ██████╗ ███████╗███████╗████████╗
	// ██╔══██╗██╔════╝██╔════╝╚══██╔══╝
	// ██████╔╝█████╗  ███████╗   ██║
	// ██╔══██╗██╔══╝  ╚════██║   ██║
	// ██║  ██║███████╗███████║   ██║
	// ╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝

	http.HandleFunc("/rest", func(w http.ResponseWriter, r *http.Request) {
		if _debug_ {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "content-type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,PUT,PATCH,POST,DELETE")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		} else if r.Method != http.MethodPost {
			w.Header().Set("Content-type", "application/json")
			http.Error(w, "{\"error\":\"Only POST data!\"}", http.StatusMethodNotAllowed)
			return
		} else {
			w.Header().Set("Content-type", "application/json")
			out := map[string]interface{}{}
			in := map[string]interface{}{}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error read body request", http.StatusInternalServerError)
				return
			}
			err = json.Unmarshal(body, &in)
			if err != nil {
				fmt.Println("Error translate income JSON:", err)
				return
			}

			op, ok := in["op"]
			if ok && op == "login" {
				user := toStr(in["user"])
				pass := toStr(in["password"])
				if md5hash(conf("pass_"+user)) == pass {
					out["user"] = user
					out["rank"] = 1
					out["hash"] = md5hash(pass) // второй раз за-md5-чим
				} else {
					out["error"] = "Unknown user. Please check login/password.."
				}
			} else if ok { // все остальные случаи только с проверкой подлинности пользователя
				if _debug_ == false {
					if !is_user(r) {
						out["error"] = "Not authorized!"
					} else {
						user, _ := r.Cookie("user")
						//уходим в отдельный файл restAPI.go....
						out["data"] = restAPI(in, user.Value)
					}
				} else {
					//уходим в отдельный файл restAPI.go....
					out["data"] = restAPI(in, "admin") //
				}
			} else {
				out["error"] = "Unknown operation via API"
			}

			json, err := json.Marshal(out)
			if err != nil {
				slog("Fail transfer to JSON", "err")
				return
			}
			_, err = w.Write([]byte(json))
			if err != nil {
				slog("Bad try send REST-answer", "err")
			}
		}
	})

	/*	███╗   ███╗ █████╗ ██╗███╗   ██╗
		████╗ ████║██╔══██╗██║████╗  ██║
		██╔████╔██║███████║██║██╔██╗ ██║
		██║╚██╔╝██║██╔══██║██║██║╚██╗██║
		██║ ╚═╝ ██║██║  ██║██║██║ ╚████║
		╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝ */

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		file, err := staticFiles.ReadFile("build/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, err = w.Write(file)
		if err != nil {
			log.Println(err)
		}
	})

	// Запуск сервера на порту
	sstart := "TVHOST v" + _version_[2] + " was run on the :" + wwwport
	slog(sstart)
	log.Fatal(http.ListenAndServe(":"+wwwport, nil))
}
