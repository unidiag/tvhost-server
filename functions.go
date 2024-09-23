package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
)

func echo(s any) {
	switch reflect.TypeOf(s).String() {
	case "string":
		fmt.Printf("%s\n", s)
	case "int", "uint", "uint32", "int32", "uint64", "int64":
		fmt.Printf("%d\n", s)
	case "[]uint8":
		fmt.Printf("%02X\n", s)
	}
}

func chkDir(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Директория не существует, создаем её
		err := os.Mkdir(dir, 0777)
		if err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}
		return false
	} else {
		return true
	}
}

func strToInt(s string) int {
	o, _ := strconv.Atoi(s)
	return o
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func countFilesInDirectory(dir string) (int, error) {
	var count int
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func getDirSize(path string) (int64, error) {
	var totalSize int64
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return totalSize, nil
}

// FileInfoWithPath включает информацию о файле и его путь
type FileInfoWithPath struct {
	os.FileInfo
	Path       string
	Size       int64
	CreateTime time.Time
}

// getFileCreateTime получает время создания файла
func getFileCreateTime(info os.FileInfo) (time.Time, error) {
	stat := info.Sys().(*syscall.Stat_t)
	return time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec), nil
}

// getFiles получает все файлы в директории
func getFiles(dir string) ([]FileInfoWithPath, error) {
	var files []FileInfoWithPath

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			createTime, err := getFileCreateTime(info)
			if err != nil {
				return err
			}
			files = append(files, FileInfoWithPath{info, path, info.Size(), createTime})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func getMbFromBytes(s int64) float64 {
	return math.Round((float64(s)/(1024*1024))*100) / 100
}

func CreateDirIfNotExist(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		parentDir := filepath.Dir(dirPath)
		if err := CreateDirIfNotExist(parentDir); err != nil {
			return err
		}
		if err := os.Mkdir(dirPath, 0777); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func slog(s ...string) {
	ss := ""
	grp := "INFO"

	if len(s) == 2 {
		grp = strings.ToUpper(s[1])
		ss = time.Now().Format("2006/01/02 15:04:05") + " [" + grp + "] " + s[0]
		if s[1] == "err" {
			color.Red(ss)
		} else {
			fmt.Println(ss)
		}
		db_query("INSERT INTO slog (c_msg, c_group, c_time) VALUES ('" + strings.ReplaceAll(s[0], "'", "''") + "', '" + grp + "', unixepoch('now'))")
		return
	} else if len(s) == 1 {
		ss = time.Now().Format("2006/01/02 15:04:05") + " [" + grp + "] " + s[0]
		fmt.Println(ss)
	}
	db_query("INSERT INTO slog (c_msg, c_group, c_time) VALUES ('" + strings.ReplaceAll(s[0], "'", "''") + "', '" + grp + "', unixepoch('now'))")
}

func toStr(value interface{}) string {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case string:
		return v
	case []byte:
		//return hex.EncodeToString(v)
		return string(v)
	default:
		// Обработка других типов или возврат ошибки
		return fmt.Sprintf("%v", value)
	}
}

func md5hash(str string) string {
	hash := md5.New()
	hash.Write([]byte(str))
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes)
}

// это для файлов ./build вебсервера
func readDirRecursively(dirPath string) ([]FileInfo, error) {
	var result []FileInfo
	files, err := staticFiles.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		fullPath := dirPath + "/" + file.Name()

		info := FileInfo{
			Path:  fullPath,
			IsDir: file.IsDir(),
		}
		result = append(result, info)
		if file.IsDir() {
			subdirContents, err := readDirRecursively(fullPath)
			if err != nil {
				return nil, err
			}
			result = append(result, subdirContents...)
		}
	}
	return result, nil
}

func getFileExtension(filePath string) string {
	parts := strings.Split(filePath, "/")
	fileName := parts[len(parts)-1]
	fileParts := strings.Split(fileName, ".")
	if len(fileParts) > 1 {
		extension := fileParts[len(fileParts)-1]
		return extension
	}
	return ""
}

func is_user(r *http.Request) bool {
	user, err := r.Cookie("user")
	hash, err2 := r.Cookie("hash")
	if err == nil && err2 == nil && user != nil && hash != nil && len(user.Value) > 0 {
		pass := conf("pass_" + user.Value)
		if hash.Value == md5hash(md5hash(pass)) {
			return true
		}
		return false
	} else {
		return false
	}
}

func conf(param ...string) string {
	out := "NULL"
	currentTime := time.Now()
	unixTime := strconv.FormatInt(currentTime.Unix(), 10)
	if len(param) == 1 { // простое берём переменную..
		result := db_query("SELECT value FROM config WHERE key='" + param[0] + "'")
		if len(result) > 0 {
			tmp, ok := result[0].(map[string]string)
			if ok {
				out = tmp["value"]
			}
		}
	} else if len(param) == 2 { // устанавливаем переменную
		v := conf(param[0])
		if v == "NULL" {
			db_query("INSERT INTO config (key, value, c_time) VALUES ('" + param[0] + "', '" + param[1] + "', '" + unixTime + "')")
		} else if v != param[1] {
			db_query("UPDATE config SET value='" + param[1] + "', last='" + v + "', e_time='" + unixTime + "' WHERE key='" + param[0] + "'")
			slog("Change config '" + param[0] + "': " + v + " => " + param[1])
		}
		out = "1"
	} else if len(param) == 3 { // устаналиваем и переменную и описание
		v := conf(param[0])
		if v == "NULL" {
			db_query("INSERT INTO config (key, value, descr, c_time) VALUES ('" + param[0] + "', '" + param[1] + "', '" + param[2] + "', '" + unixTime + "')")
		} else {
			db_query("UPDATE config SET value='" + param[1] + "', descr='" + param[2] + "', last='" + v + "', e_time='" + unixTime + "' WHERE key='" + param[0] + "'")
		}
		out = "1"
	}
	return out
}

func esc(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func systemd() {

	if !_debug_ {
		text := `[Unit]
Description=tvhost
After=network.target

[Service]
Type=simple
ExecStart=`
		executablePath, _ := os.Executable()
		text += executablePath + "\n"
		text += "WorkingDirectory=" + filepath.Dir(executablePath) + "\n\n"
		text += `[Install]
WantedBy=multi-user.target
Alias=tvhost.service
`

		os.WriteFile("/etc/systemd/system/tvhost.service", []byte(text), 0644)
		fmt.Println("Create unit [tvhost] in systemd. Run:\n\tsystemctl enable tvhost\n\tsystemctl start tvhost")
	}

}

func isValidPort(str string) bool {
	port := conf(str)
	num, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	ret := (num >= 80 && num <= 65535) || num == 0
	if !ret {
		color.Red("Not valid " + str + ": " + port)
	}
	return ret
}

type AddrUdp struct {
	Eth  string
	Addr string
	Port int
}

func parseAddrUdp(uri string) AddrUdp {
	out := AddrUdp{}
	re := regexp.MustCompile(`^udp://([^@]*)@([0-9.]+)(?::(\d+))?$`)
	matches := re.FindStringSubmatch(uri)
	if len(matches) != 4 {
		slog("Invalid address format: " + uri)
	}
	if IsIPv4(out.Eth) {
		out.Eth = matches[1]
	} else {
		out.Eth = getAddrByName(matches[1])
	}

	if out.Eth == "<nil>" || out.Eth == "" {
		out.Eth = "127.0.0.1"
	}

	out.Addr = matches[2]
	if matches[3] == "" {
		matches[3] = "1234"
	}
	out.Port, _ = strconv.Atoi(matches[3])
	return out
}

func IsIPv4(address string) bool {
	ip := net.ParseIP(address)
	if ip == nil {
		return false
	}
	return ip.To4() != nil
}

func getAddrByName(eth string) string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	var iface net.Interface
	for _, i := range interfaces {
		if i.Name == eth {
			iface = i
			break
		}
	}
	if iface.Index == 0 {
		return ""
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return ""
	}
	var ipv4Addr net.IP
	for _, addr := range addrs {
		switch v := addr.(type) {
		case *net.IPNet:
			if v.IP.To4() != nil {
				ipv4Addr = v.IP
			}
		}
	}
	return ipv4Addr.String()
}

// копирует сторонний софт во временную рабочую папку
func copy2tmp(files []string) {
	tmpBin := tmpDir + "/bin"
	chkDir(tmpBin)
	for i := 0; i < len(files); i++ {
		fileData, _ := staticFiles.ReadFile("build/" + files[i])
		os.WriteFile(tmpBin+"/"+files[i], fileData, 0777)
	}
}

func delay(milliseconds int) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

func copyFile(sourceFile, destFile string) error {
	src, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer src.Close()
	dst, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy data: %v", err)
	}
	return nil
}

func kill(p string) {
	cmd := exec.Command("pidof", p)
	output, _ := cmd.Output()
	// Получить PID из команды pidof
	pid := strings.TrimSpace(string(output))
	if p != "astra_epg" {
		os.Remove(tmpDir + "/bin/" + p)
	}
	if pid == "" {
		//log.Println("Process astra_ch1 not found")
		return
	} else if strings.Contains(pid, " ") {
		parts := strings.Split(pid, " ")
		for i := 0; i < len(parts); i++ {
			exec.Command("kill", "-9", parts[i]).Run()
			delay(200)
		}
		return
	}

	exec.Command("kill", "-9", pid).Run()
	delay(100)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func getLogoBin(id string) []byte {
	o, _ := staticFiles.ReadFile("build/0.png")
	if fileExists("./logos/" + id + ".png") {
		o, _ = os.ReadFile("./logos/" + id + ".png")
	} else if fileExists("./logos/0.png") {
		o, _ = os.ReadFile("./logos/0.png")
	}
	return o
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func truncateString(s string) string {
	const maxLength = 85
	if len(s) <= maxLength {
		return s
	}
	truncated := s[:maxLength]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace != -1 {
		truncated = truncated[:lastSpace]
	}
	o := strings.TrimSuffix(truncated, ".")
	o = strings.TrimSuffix(o, ":")
	return o + "..."
}

func pidof(proc string) []string {
	o := []string{}
	cmd := exec.Command("pidof", proc)
	output, err := cmd.Output()
	if err == nil {
		o = strings.Split(string(output), " ")
	}
	return o
}

func convertToUTC(timeStr string) int64 {
	t, err := time.Parse("20060102150405 -0700", timeStr)
	if err == nil {
		return t.UTC().Unix()
	}
	return int64(0)
}

func printMemUsage() {
	go func() {
		for {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			color.Green("Alloc = %v MiB\tTotalAlloc = %v MiB\tSys = %v MiB\tNumGC = %v", bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC)
			time.Sleep(30 * time.Second)
		}
	}()
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// Структура для хранения информации о процессе
type ProcessInfo struct {
	Cmd    *exec.Cmd
	Cancel context.CancelFunc
}

// Мапа для хранения запущенных процессов
var processes = make(map[string]ProcessInfo)

// Запуск и управление процессом
func startProcess(path string, args ...string) {
	name := filepath.Base(path)
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, path, args...)
	if err := cmd.Start(); err != nil {
		slog("Failed to start process: `"+name+"`", "err")
	}
	mu.Lock()
	processes[name] = ProcessInfo{
		Cmd:    cmd,
		Cancel: cancel,
	}
	mu.Unlock()
}

// Остановка процесса по имени
func stopProcess(name string) {
	mu.Lock()
	processInfo, exists := processes[name]
	if exists {
		processInfo.Cancel()
		if err := processInfo.Cmd.Wait(); err != nil {
			if len(pidof(name)) != 0 {
				slog("Error exiting process: `"+name+"`", "err")
			}
		}
		delete(processes, name)
	}
	mu.Unlock()
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
