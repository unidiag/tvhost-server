package main

import (
	"crypto/rand"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

func restAPI(data map[string]interface{}, user string) map[string]interface{} {
	out := map[string]interface{}{}

	//currentTime := time.Now()
	//ttime := strconv.FormatInt(currentTime.Unix()-1, 10)

	switch data["op"] {

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
	case "getEpg":
		chid := toStr(data["id"])
		row := db_fetchassoc(db_query("SELECT * FROM channels WHERE id=" + chid))
		out = db_fetchrow(db_query("SELECT s_time, title FROM epg WHERE chid=" + chid + " ORDER BY s_time"))
		out["chname"] = row["chname"]
		out["u_time"] = row["u_time"]

	case "clearEpg":
		chid := toStr(data["id"])
		row := db_fetchassoc(db_query("SELECT * FROM channels WHERE id=" + chid))
		slog("Clear EPG for channel #" + chid + ": " + row["chname"])
		db_query("UPDATE channels SET u_time=1 WHERE id=" + chid)
		db_query("DELETE FROM epg WHERE chid=" + chid)
		out["status"] = "OK"

	//
	//
	//
	//
	//
	//
	//
	//
	//  ██████╗██╗  ██╗ █████╗ ███╗   ██╗███╗   ██╗███████╗██╗     ███████╗
	// ██╔════╝██║  ██║██╔══██╗████╗  ██║████╗  ██║██╔════╝██║     ██╔════╝
	// ██║     ███████║███████║██╔██╗ ██║██╔██╗ ██║█████╗  ██║     ███████╗
	// ██║     ██╔══██║██╔══██║██║╚██╗██║██║╚██╗██║██╔══╝  ██║     ╚════██║
	// ╚██████╗██║  ██║██║  ██║██║ ╚████║██║ ╚████║███████╗███████╗███████║
	//  ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═══╝╚══════╝╚══════╝╚══════╝

	case "getChannels":
		out = db_fetchrow(db_query("SELECT * FROM channels ORDER BY id DESC"))

	case "addChannel":
		sql := "INSERT INTO channels (chname, uri, folder, enable, visible, u_time, c_time, e_time, xmlid) VALUES ("
		folderPath := toStr(data["folder"])
		if folderPath[0] != '/' && folderPath[0] != '.' {
			folderPath = "/" + folderPath
		}
		chname := toStr(data["chname"])
		sql2 := "'" + esc(chname) + "', '" + esc(toStr(data["uri"])) + "', '" + esc(folderPath) + "'"
		slog("Add channel '" + chname + "'")
		xmlid := esc(toStr(data["xmlid"]))
		if xmlid == "<nil>" {
			xmlid = ""
		}
		db_query(sql + sql2 + ", " + toStr(data["enable"]) + ", " + esc(toStr(data["visible"])) + ", 1, unixepoch('now'), 0, '" + xmlid + "')")
		rr := db_fetchassoc(db_query("SELECT * FROM channels ORDER BY id DESC LIMIT 1")) // добавляем `-ID``
		ff := rr["folder"]
		if len(rr["folder"]) > 1 {
			ff += "-"
		}
		ff += rr["id"]
		db_query("UPDATE channels SET folder='" + ff + "' WHERE id=" + rr["id"])
		CreateDirIfNotExist(ff)
		out["status"] = "OK"

	case "editChannel":
		chname := toStr(data["chname"])
		sql := "chname='" + esc(chname) + "', uri='" + esc(toStr(data["uri"])) + "', enable=" + toStr(data["enable"]) + ", visible=" + toStr(data["visible"]) + ", xmlid='" + toStr(data["xmlid"]) + "'"
		slog("Edit channel #" + toStr(data["id"]) + ": " + sql)
		db_query("UPDATE channels SET " + sql + ", e_time=unixepoch('now') WHERE id=" + toStr(data["id"]))
		out["status"] = "OK"

	case "deleteChannel":
		id := toStr(data["id"])
		row := db_fetchassoc(db_query("SELECT * FROM channels WHERE id=" + id))
		os.RemoveAll(row["folder"])
		slog("Delete channel #" + id + ": " + row["chname"])
		db_query("DELETE FROM channels WHERE id=" + toStr(data["id"]))
		db_query("DELETE FROM epg WHERE chid=" + id)
		os.Remove("./logos/" + id + ".png")
		mu.Lock()
		if _, ok := recordsMap[id]; ok {
			delete(recordsMap, id)
		}
		mu.Unlock()
		out["status"] = "OK"

	case "scanUdp":
		cmd := exec.Command("netstat", "-u", "-n", "-a")
		output, err := cmd.Output()
		if err != nil {
			slog("Error execute netstat -una")
		}
		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")
		required := []string{}
		for _, line := range lines {
			if strings.Contains(line, "udp") {
				re := regexp.MustCompile(`(22[4-9]|23[0-9]\.\d+\.\d+\.\d+):(\d+)`)
				matches := re.FindStringSubmatch(line)
				if len(matches) == 3 {
					required = append(required, matches[0])
				}
			}
		}

		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(required))))
		//fmt.Println(len(required))
		out["scan"] = udp_scan(required[n.Int64()])
		out["status"] = "OK"

	case "getInfoFolder":
		dir := toStr(data["folder"])
		s, _ := getDirSize(dir)
		out["size"] = getMbFromBytes(s)
		out["count"], _ = countFilesInDirectory(dir)

		files, _ := getFiles(dir)
		if len(files) > 0 {
			sort.Slice(files, func(i, j int) bool {
				return files[i].Name() < files[j].Name()
			})
			out["starttime"] = files[0].CreateTime.Format(time.RFC3339)
			out["startfile"] = filepath.Base(files[0].Path)
			out["startsize"] = getMbFromBytes(files[0].Size)

			out["finishtime"] = files[len(files)-1].CreateTime.Format(time.RFC3339)
			out["finishfile"] = filepath.Base(files[len(files)-1].Path)
			out["finishsize"] = getMbFromBytes(files[len(files)-1].Size)
		}

		out["status"] = "OK"

	case "deleteLogo":
		chid := toStr(data["chid"])
		os.Remove("./logos/" + chid + ".png")
		slog("Delete channel logotype #" + chid)
		out["status"] = "OK"

	// ███████╗████████╗██████╗ ███████╗ █████╗ ███╗   ███╗███████╗
	// ██╔════╝╚══██╔══╝██╔══██╗██╔════╝██╔══██╗████╗ ████║██╔════╝
	// ███████╗   ██║   ██████╔╝█████╗  ███████║██╔████╔██║███████╗
	// ╚════██║   ██║   ██╔══██╗██╔══╝  ██╔══██║██║╚██╔╝██║╚════██║
	// ███████║   ██║   ██║  ██║███████╗██║  ██║██║ ╚═╝ ██║███████║
	// ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝

	case "getInfo":
		play := strings.Split(toStr(data["play"]), ":")
		r := db_fetchassoc(db_query("SELECT * FROM epg WHERE id=" + play[1]))
		rr := db_fetchassoc(db_query("SELECT * FROM channels WHERE id=" + play[0]))
		out["chid"] = rr["id"]
		out["chname"] = rr["chname"]
		out["title"] = r["title"]
		out["s_time"] = r["s_time"]
		out["f_time"] = r["f_time"]

	case "editStream":
		newuri := esc(toStr(data["uri"]))
		r := db_fetchassoc(db_query("SELECT * FROM streams WHERE id=" + toStr(data["id"])))
		db_query("UPDATE streams SET uri='" + newuri + "', e_time=unixepoch('now') WHERE id=" + r["id"])
		slog("Changed stream URI #" + r["id"] + ": " + r["uri"] + " => " + newuri)
		astraStart(r)
		delay(150)
		out["status"] = "OK"

	case "getStreams":
		out = db_fetchrow(db_query("SELECT * FROM streams ORDER BY id"))
		for k, v := range out {
			if vv, ok := v.(map[string]string); ok {
				vv["link"] = conf("site") + "/?ch=" + vv["link"]
				out[k] = vv
			}
		}

	case "switchOnOff":
		id := toStr(data["id"])
		enable := 0
		ext := ""
		if data["state"] == "0" {
			enable = 1
			astraStart(db_fetchassoc(db_query("SELECT * FROM streams WHERE id=" + id)))
			slog("Stream #" + id + ": switch ON")
		} else {
			kill("astra_ch" + id)
			kill("tsplay_ch" + id)
			ext = ", s_time=0"
			slog("Stream #" + id + ": switch OFF")
		}
		db_query("UPDATE streams SET enable=" + toStr(enable) + ext + ", play='', ip='', b_time=0, e_time=unixepoch('now') WHERE id=" + id)

	//  ██████╗ ██████╗ ███╗   ██╗███████╗
	// ██╔════╝██╔═══██╗████╗  ██║██╔════╝
	// ██║     ██║   ██║██╔██╗ ██║█████╗
	// ██║     ██║   ██║██║╚██╗██║██╔══╝
	// ╚██████╗╚██████╔╝██║ ╚████║██║
	//  ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝
	case "getConf":
		w_oper := ""
		if user == "oper" {
			w_oper = " WHERE key='pass_oper'"
		}
		out = db_fetchrow(db_query("SELECT * FROM config" + w_oper + " ORDER BY id"))

	case "setConf":
		formData := []interface{}{}
		_ = formData
		if !demo {
			if formData, ok := data["formData"].([]interface{}); ok {
				backup_db()
				for _, v := range formData {
					if entry, ok := v.(map[string]interface{}); ok {
						key, keyOk := entry["key"].(string)
						value, valueOk := entry["value"].(string)
						if keyOk && valueOk {
							conf(key, value)
						}
					}
				}
				_, err = exec.Command("systemctl", "restart", "tvhost").Output()
				if err != nil {
					slog("Fail restart service TVHOST:", "err")
				}
			}

			out["status"] = "OK"
		} else {
			out["status"] = "Error update config"
		}
		//go restart(4)

		// ██████╗  █████╗ ████████╗██╗  ██╗
		// ██╔══██╗██╔══██╗╚══██╔══╝██║  ██║
		// ██████╔╝███████║   ██║   ███████║
		// ██╔═══╝ ██╔══██║   ██║   ██╔══██║
		// ██║     ██║  ██║   ██║   ██║  ██║
		// ╚═╝     ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝

	case "getPaths":
		out["mode"] = "0"
		if user == "admin" {
			out["paths"] = "streamer,logs"
		} else {
			out["paths"] = conf("paths")
		}

	// ██╗      ██████╗  ██████╗ ███████╗
	// ██║     ██╔═══██╗██╔════╝ ██╔════╝
	// ██║     ██║   ██║██║  ███╗███████╗
	// ██║     ██║   ██║██║   ██║╚════██║
	// ███████╗╚██████╔╝╚██████╔╝███████║
	// ╚══════╝ ╚═════╝  ╚═════╝ ╚══════╝

	case "getLogs":
		out = db_fetchrow(db_query("SELECT * FROM slog WHERE c_msg LIKE '%" + toStr(data["search"]) + "%' COLLATE NOCASE OR c_group LIKE '" + strings.ToUpper(toStr(data["search"])) + "%' ORDER BY id DESC LIMIT 50"))

	case "clearLogs":
		if backup_db() {
			db_query("DROP TABLE slog;")
			db_init(0)
			slog("Cleared LOGs database")
			out["status"] = "OK"
		} else {
			out["error"] = "Error create backup DB"
		}
	default:
		out["error"] = "Not exists OPERATION '" + toStr(data["op"]) + "' command"
	}

	return out

}
