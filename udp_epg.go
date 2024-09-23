package main

import (
	"os"
)

func update_epg(row map[string]string) {
	if row["xmlid"] != "" {
		update_xmltv(row)
		return
	}
	stopProcess("astra_epg")
	db_query("DELETE FROM epg WHERE chid=" + row["id"] + " AND s_time < unixepoch('now')-3600*" + conf("rectime"))

	lua := `make_channel({
  id = "` + row["id"] + `", name = "epg",
  input = { "` + row["uri"] + `" },
  output = { "udp://lo@235.55.55.55" },
  epg_export_format = "json",
  epg_export = "http://127.0.0.1:` + conf("wwwport") + `/epg?k=` + epgkey + `"
})

timer({
    interval = 67,
    callback = function()
      os.exit()      
    end
})`

	os.WriteFile(tmpDir+"/epg.lua", []byte(lua), 0644)
	startProcess(tmpDir+"/bin/astra_epg", tmpDir+"/epg.lua")
	if !_debug_ {
		os.Remove(tmpDir + "/epg.lua")
	}
	db_query("UPDATE channels SET u_time=unixepoch('now') WHERE id=" + row["id"])
}
