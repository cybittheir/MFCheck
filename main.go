package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	appname      string
	version      string
	build        string
	getTickCount = syscall.NewLazyDLL("kernel32.dll").NewProc("GetTickCount64")
)

type netMACIP struct {
	ip  string
	mac string
}

func getIP() netMACIP {
	// getting PCs IP-address

	ifaces, err := net.Interfaces()

	if err != nil {
		fmt.Println(err)
	}

	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		status := i.Flags.String()
		mac := i.HardwareAddr.String()

		statuses := strings.Split(status, "|")

		if err != nil {
			fmt.Println(err)
		} else if statuses[0] == "up" {
			// handle err
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				// process IP address
				tpart := strings.Split(ip.String(), ".")

				if len(tpart) == 4 {
					t0, _ := strconv.Atoi(tpart[0])
					t1, _ := strconv.Atoi(tpart[1])
					t2, _ := strconv.Atoi(tpart[2])
					t3, _ := strconv.Atoi(tpart[3])

					if (t0 == 10) || (t0 == 192 && t1 == 168) || (t0 == 172 && t1 > 15 && t1 < 32) {
						strIP := strconv.Itoa(t0) + "." + strconv.Itoa(t1) + "." + strconv.Itoa(t2) + "." + strconv.Itoa(t3)
						return netMACIP{strIP, mac}
					}

				}
			}
		}
	}

	return netMACIP{}

}

func isProcRunning(batchPath string, batchName string, name string, silent bool) (bool, error) {
	//check if process is running. Windows tasklist in batchfile uses
	if !silent {
		fmt.Print(".")
	}

	cmd := exec.Command(batchPath+batchName, name)
	cmd.Dir = batchPath

	out, err := cmd.Output()

	if err != nil {
		if !silent {
			fmt.Print(".")
		}
		return false, err
	}

	if bytes.Contains(out, []byte(name)) {
		if !silent {
			fmt.Print(".")
		}
		return true, nil
	}
	return false, nil
}

func checkHost(host map[string]string) (bool, error) {

	hostIP := host["ip"]
	port := host["port"]
	address := net.JoinHostPort(hostIP, port)
	conn, err := net.DialTimeout("tcp", address, time.Second)
	results := make(map[string]bool)

	if err != nil {
		results[port] = false
		return false, err
		// todo log handler
	} else {
		if conn != nil {
			results[port] = true
			_ = conn.Close()
			return results[port], nil
		} else {
			results[port] = false
			return false, nil
		}
	}

}

func getUptime() (time.Duration, error) {
	//Getting uptime of PC
	ret, _, err := getTickCount.Call()
	if errno, ok := err.(syscall.Errno); !ok || errno != 0 {
		return time.Duration(0), err
	}
	return time.Duration(ret) * time.Millisecond, nil
}

func sendQuery(url string, token string, urlQuery string, silent bool) (bool, error) {
	// Make HTTP GET request
	response, err := http.Get(url + "?UID=" + token + urlQuery)
	if err != nil {
		log.Fatal(err)
		return false, err
	} else {
		if !silent {
			fmt.Println("Sending info to server:")
			fmt.Println(urlQuery)
		}
		// Copy data from the response to standard output
		_, err := io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
			return false, err
		}
	}
	defer response.Body.Close()
	return true, nil

}

func timer(startTime int64, timePeriod int, silent bool) {

	jobTimeNow := time.Now()
	jobTime := jobTimeNow.Unix() - startTime

	usePeriod := timePeriod - int(jobTime)
	period := fmt.Sprintf("%d", timePeriod)

	if !silent {
		fmt.Println("\n\nSleeping for " + period + " seconds... Zzz...\n")
	}

	time.Sleep(time.Duration(usePeriod) * time.Second)

}

func initArgs() (bool, error) {

	silent := false

	if len(os.Args) != 1 {

		arg := os.Args[1]

		if arg == "-h" || arg == "-help" {

			fmt.Println("")
			fmt.Println("Use conf.json file configuration:")
			fmt.Println("connect:")
			fmt.Println("   period: 60 // seconds, 60 sec minimum")
			fmt.Println("   url: https://[url]")
			fmt.Println("   token: [token] // ?UID=token")
			fmt.Println("   pin: 000000 // ?pin=[pin]")
			fmt.Println("   batch: batch.bat // context with tasklist.exe /FO CSV /NH | find '%1'")
			fmt.Println("   path: C:\\[PATH]\\ //path to batch-file")
			fmt.Println("check:")
			fmt.Println("   process: //tests applications are running. =9 if OK, =1 if failed")
			fmt.Println("      app1: app1.exe")
			fmt.Println("      app2: app2.exe")
			fmt.Println("   device: //tests hosts are reachable. =failed if NOT")
			fmt.Println("      dev1:")
			fmt.Println("         ip: 192.168.0.1")
			fmt.Println("         port: 80")
			fmt.Println("      dev2:")
			fmt.Println("         ip: 192.168.0.2")
			fmt.Println("         port: 22")
			fmt.Println("")
			fmt.Println("Use -s (OR -silent) option for hiding all messages except errors")
			fmt.Println("Use Ctrl+C for exit")
			fmt.Println("")

			os.Exit(0)
		} else if arg == "-s" || arg == "-silent" {
			silent = true
			return silent, nil
		}
	}

	return silent, nil

}

func main() {

	fmt.Println(appname, version, "build", build)
	fmt.Println("Simple agent for checking links to devices and running applications on remote PC")
	fmt.Println("https://github.com/cybittheir/MFCheck")

	silent, _ := initArgs()

	// Open our jsonFile

	jsonFile, err := os.Open("conf.json")

	// if we os.Open returns an error then handle it

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !silent {
		fmt.Println("Successfully Opened conf.json")
	}

	// defer the closing of our jsonFile so that we can parse it later on

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var confResult map[string]map[string]string
	var checkProc map[string]map[string]map[string]string
	var checkConn map[string]map[string]map[string]map[string]string

	json.Unmarshal([]byte(byteValue), &confResult)
	json.Unmarshal([]byte(byteValue), &checkProc)
	json.Unmarshal([]byte(byteValue), &checkConn)

	var emptyConfig bool
	var timePeriod int

	if confResult["connect"]["period"] == "" {
		fmt.Println("Config fatal error: Parameter [connect][period] is empty OR JSON structure error")
		emptyConfig = true
	}
	if confResult["connect"]["url"] == "" {
		fmt.Println("Config fatal error: Parameter [connect][url] is empty OR JSON structure error")
		emptyConfig = true
	}
	if confResult["connect"]["token"] == "" {
		fmt.Println("Config fatal error: Parameter [connect][token] is empty OR JSON structure error")
		emptyConfig = true
	}
	if confResult["connect"]["pin"] == "" {
		fmt.Println("Config fatal error: Parameter [connect][pin] is empty OR JSON structure error")
		emptyConfig = true
	}
	if confResult["connect"]["batch"] == "" {
		fmt.Println("Config fatal error: Parameter [connect][batch] is empty OR JSON structure error")
		emptyConfig = true
	}
	if confResult["connect"]["path"] == "" {
		fmt.Println("Config fatal error: Parameter [connect][path] is empty OR JSON structure error")
		emptyConfig = true
	}

	if emptyConfig {
		fmt.Println("use -help argument")
		os.Exit(0)
	}

	period := confResult["connect"]["period"]
	url := confResult["connect"]["url"]
	token := confResult["connect"]["token"]
	pin := confResult["connect"]["pin"]
	batchName := confResult["connect"]["batch"]
	batchPath := confResult["connect"]["path"]
	timePeriod, _ = strconv.Atoi(period)

	if timePeriod < 60 {
		timePeriod = 60
	}

	if _, err := os.Stat(batchPath + batchName); err != nil {
		fmt.Println("Fatal error. Batch file not exists. Check its path")
		fmt.Println("use -help argument")
		os.Exit(0)
	}
	// начало цикла проверки
	for {
		startTimeNow := time.Now()
		startTime := startTimeNow.Unix()

		procList := ""
		deviceList := ""

		if len(checkProc["check"]["process"]) > 0 {

			if !silent {
				fmt.Printf("%s", "Check Processes ")
			}

			for i, v := range checkProc["check"]["process"] {

				if v != "" {
					forCheck := v

					isRunning, _ := isProcRunning(batchPath, batchName, forCheck, silent)

					forCheckParam := i

					if isRunning {
						procList = procList + "&" + forCheckParam + "=9"
					} else {
						procList = procList + "&" + forCheckParam + "=1"
					}
				}
			}

			if !silent {
				fmt.Println(" Finished")
			}

		} else {

			fmt.Println("Warning! Section [check][process] in conf.json file is needed for cheching running processes")
			fmt.Println("use -help argument")
		}

		if len(checkConn["check"]["device"]) > 0 {
			if !silent {
				fmt.Println("devices checking:")
			}
			for i, v := range checkConn["check"]["device"] {
				if len(v) > 0 {

					deviceOk, _ := checkHost(v)

					if deviceOk {
						if !silent {
							fmt.Println("...", i, "is OK")
						}
					} else {
						if !silent {
							fmt.Println("...", i, "Failed")
						}
						deviceList = deviceList + "&x_" + i + "=failed"

					}
				} else {

					fmt.Println(i, "!error conf record!")

				}

			}
		}

		hostname, err := os.Hostname()
		if err != nil {
			fmt.Println(err)
		}

		uptime, _ := getUptime()

		CompUptime, _ := time.ParseDuration(fmt.Sprint(uptime))

		PCUp := fmt.Sprint(int(CompUptime.Minutes()))

		netResult := getIP()

		queryPIN := "&mfpin=" + pin
		queryIP := "&mfip=" + netResult.ip
		queryMAC := "&mfmac=" + netResult.mac
		queryUPTime := "&mfuptime=" + PCUp
		queryPCName := "&mfname=" + hostname

		dt := time.Now()

		queryTIME := "&mfdate=" + dt.Format("2006-01-02") + "&mftime=" + dt.Format("15:04") + "&mftimefull=" + dt.Format("15:04:05")

		urlQuery := queryPIN + queryTIME + queryIP + queryMAC + queryUPTime + queryPCName + procList + deviceList

		_, err = sendQuery(url, token, urlQuery, silent)

		procList = ""
		queryTIME = ""
		urlQuery = ""

		if err != nil {
			log.Fatal(err)
			time.Sleep(5 * time.Second)
		} else {
			if !silent {
				fmt.Println("\n-----------------")
				log.Println("Success.")
			}
			timer(startTime, timePeriod, silent)
		}

	}

}
