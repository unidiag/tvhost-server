package main

func test() {
	if _debug_ {

		//ffmpeg -i "concat:1.ts|2.ts" -c copy -f mpegts udp://192.168.1.100:1234
		// cvlc -vvv --loop --playlist-autostart playlist.m3u --miface-addr 127.0.0.1 --http-reconnect http://sitedonor.com:1200 --sout "#std{access=udp,mux=ts,dst=239.2.100.45:1234}" --ttl 5
		//vlc --sout "#std{access=udp,mux=ts,dst=239.255.0.1:1234}" --sout-ts-pcr 100 --sout-ts-dts-delay 400 --loop --playlist-autostart playlist.m3u
		// /tmp/.tvhost/bin/ffmpeg -f image2 -i /tmp/.tvhost/1/img%d.jpg -aspect 16:9 -qscale 1 -g 100 -mpegts_service_id 0x64 -mpegts_pmt_start_pid 0x190 -mpegts_start_pid 0x191 -metadata service_provider='Archive' -metadata service_name='Archive 1' /1.ts

		//os.Exit(0)
	}
}
