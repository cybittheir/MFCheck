/*

Getting some information from PC:
  - uptime;
  - PC time;
  - name;
  - IP & MAC;
  - selected processes running

  and sending results to server via GET request

  Aleksandr Lovin aka cybittheir
*/

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
	"net/url"
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
	getTickCount      = syscall.NewLazyDLL("kernel32.dll").NewProc("GetTickCount64")
	silent       bool = false // display all messages
)

const (
	def_period  = 60 // default circle period
	err_timeout = 10 // timeout if connections error
)

type netMACIP struct {
	ip  string
	mac string
}

/*
func Greeting() {

	fmt.Println(appname, version, "build", build)
	fmt.Println("Simple agent for checking links to devices and running applications on remote PC")
	fmt.Println("https://github.com/cybittheir/MFCheck")

}
*/

func getIP() netMACIP {
	// getting PCs IP-address and MAC

	mac := ""
	strIP := ""

	for mac == "" && strIP == "" {
		ifaces, err := net.Interfaces()

		if err == nil {

			// handle err
			for _, i := range ifaces {
				addrs, err := i.Addrs()
				status := i.Flags.String()
				mac = i.HardwareAddr.String()

				statuses := strings.Split(status, "|")

				if err != nil {
					log.Println(err)
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

							if ((t0 == 192 && t1 == 168) || (t0 == 10) || (t0 == 172 && t1 > 15 && t1 < 32)) && t3 != 1 && mac != "" {
								strIP = fmt.Sprintf("%d.%d.%d.%d", t0, t1, t2, t3)
								return netMACIP{strIP, mac}
							}

						}
					}
				}
			}

		} else {
			log.Println(err)
		}

		log.Println("Error: all interfaces are DOWN")
		time.Sleep(err_timeout * time.Second)
	}

	return netMACIP{}
}

func isProcRunning(batchPath, batchName, name string) (bool, error) {
	//check if process is running. Windows tasklist in batchfile uses
	printMSG(".")

	cmd := exec.Command(batchPath+batchName, name)
	cmd.Dir = batchPath

	out, err := cmd.Output()

	if err != nil {
		printMSG(".")
		return false, err
	}

	if bytes.Contains(out, []byte(name)) {
		printMSG(".")
		return true, nil
	}
	return false, nil
}

func printMSGln(message string) {
	if !silent {
		fmt.Println(message)
	}
}

func printMSG(message string) {
	if !silent {
		fmt.Print(message)
	}
}

func getTargetHost(target_url string) (map[string]string, error) {

	var target map[string]string
	target = make(map[string]string)

	u, err := url.Parse(target_url)

	target["address"] = u.Host
	target["port"] = u.Scheme

	return target, err

}

func checkTarget(host map[string]string) (bool, error) {
	// checking access for any target host (local and remote)
	target := ""
	port := host["port"]
	if host["address"] != "" {

		target = host["address"]

	} else if host["ip"] != "" {

		target = host["ip"]

	} else {
		return false, nil
	}

	targetAddress := net.JoinHostPort(target, port)
	conn, err := net.DialTimeout("tcp", targetAddress, time.Second)

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

func sendQuery(url, token, urlQuery string) (bool, error) {
	// Make HTTP GET request
	response, err := http.Get(url + "?UID=" + token + urlQuery)
	if err == nil {
		if !silent {
			io.WriteString(os.Stdout, fmt.Sprintf("Sending info to server: \n%s", urlQuery))
		}
		// Copy data from the response to standard output
		_, warning := io.Copy(os.Stdout, response.Body)
		if warning != nil {
			return false, warning
		}
	} else {
		return false, err
	}
	defer response.Body.Close()
	return true, nil

}

func timer(startTime int64, timePeriod int) {
	// set timer for pause
	jobTimeNow := time.Now()
	jobTime := jobTimeNow.Unix() - startTime

	usePeriod := timePeriod - int(jobTime)
	period := fmt.Sprintf("%d", timePeriod)

	printMSGln("\n\nSleeping for " + period + " seconds... Zzz...\n")

	time.Sleep(time.Duration(usePeriod) * time.Second)

}

func initArgs() {
	// reading starting parameters

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
			silent = true // display only main messages & errors
		}
	}
}

func readConfJSON() ([]byte, error) {

	// Open our jsonFile

	var byteValue []byte
	jsonFile, err := os.Open("conf.json")
	// if we os.Open returns an error then handle it

	if err == nil {

		// defer the closing of our jsonFile so that we can parse it later on
		defer jsonFile.Close()

		byteValue, warning := ioutil.ReadAll(jsonFile)

		return byteValue, warning
	} else {
		log.Println("Cannot open conf.json")
		log.Println(err)
		os.Exit(1)
	}
	return byteValue, err

}

func main() {

	//	fmt.Println(appname, version, "build", build)
	//	fmt.Println("Simple agent for checking links to devices and running applications on remote PC")
	//	fmt.Println("https://github.com/cybittheir/MFCheck")

	Greeting()

	initArgs() // read commandline parameters

	byteValue, err := readConfJSON()

	if err == nil {
		printMSGln("Successfully opened config file")
	} else {
		log.Println(err)
		os.Exit(1)
	}

	var confResult map[string]map[string]string
	var checkProc map[string]map[string]map[string]string
	var checkConn map[string]map[string]map[string]map[string]string

	json.Unmarshal([]byte(byteValue), &confResult)
	json.Unmarshal([]byte(byteValue), &checkProc)
	json.Unmarshal([]byte(byteValue), &checkConn)

	var emptyConfig bool
	var timePeriod int
	// checking config parameters

	for key, val := range confResult["connect"] {
		if val == "" {
			log.Println("Config fatal error: Parameter [connect][" + key + "] is empty OR JSON structure error")
			emptyConfig = true
		}
	}

	target_url := confResult["connect"]["url"]
	if target_url == "" {
		log.Println("Config fatal error: Parameter [connect][url] is required.")
		emptyConfig = true
	}

	token := confResult["connect"]["token"]
	if token == "" {
		log.Println("Config fatal error: Parameter [connect][token] is required.")
		emptyConfig = true
	}

	pin := confResult["connect"]["pin"]
	if pin == "" {
		log.Println("Config fatal error: Parameter [connect][pin] is required.")
		emptyConfig = true
	}

	batchName := confResult["connect"]["batch"]
	if batchName == "" {
		log.Println("Config fatal error: Parameter [connect][batch] is required.")
		emptyConfig = true
	}

	batchPath := confResult["connect"]["path"]
	if batchPath == "" {
		log.Println("Config fatal error: Parameter [connect][path] is required.")
		emptyConfig = true
	}
	// check batchfile exists
	if _, err := os.Stat(batchPath + batchName); err != nil {
		log.Println("Config fatal error: Batch file", batchPath+batchName, "not exists. Check it")
		emptyConfig = true
	}

	if emptyConfig {
		log.Println("use -help argument")
		os.Exit(0)
	}

	period := confResult["connect"]["period"]
	if period == "" {
		period = strconv.Itoa(def_period)
	}

	timePeriod, _ = strconv.Atoi(period)

	if timePeriod < def_period {
		timePeriod = def_period
	}

	// get hostname
	hostname, err := os.Hostname()
	if err != nil {
		log.Println(err)
	}

	// get IP & MAC
	netResult := getIP()

	for netResult.mac == "" && netResult.ip == "" {
		time.Sleep(err_timeout * time.Second)
		netResult = getIP()
	}

	queryPIN := "&mfpin=" + pin
	queryPCName := "&mfname=" + hostname
	queryIP := "&mfip=" + netResult.ip
	queryMAC := "&mfmac=" + netResult.mac

	// start checking running processes
	for {
		procList := ""
		deviceList := ""

		startTimeNow := time.Now()
		startTime := startTimeNow.Unix()

		if len(checkProc["check"]["process"]) > 0 {

			if !silent {
				fmt.Printf("%s", "Check Processes ")
			}

			for i, v := range checkProc["check"]["process"] {

				if v != "" {
					forCheck := v

					isRunning, _ := isProcRunning(batchPath, batchName, forCheck)

					forCheckParam := i

					if isRunning {
						procList = procList + "&" + forCheckParam + "=9" // just because I decide this
					} else {
						procList = procList + "&" + forCheckParam + "=1"
					}
				}
			}
			// print results if -s not used
			if !silent {
				fmt.Println(" Finished")
			}

		} else {

			fmt.Println("Warning! Section [check][process] in conf.json file is needed for cheching running processes")
			fmt.Println("use -help argument")
		}

		// start checking devices connection

		if len(checkConn["check"]["device"]) > 0 {
			printMSGln("devices checking:")
			for i, v := range checkConn["check"]["device"] {
				if len(v) > 0 {

					//deviceOk, _ := checkHost(v) //left while not release
					deviceOk, _ := checkTarget(v)

					if deviceOk {
						printMSGln("... " + i + " is OK")
					} else {
						printMSGln("... " + i + " Failed")
						deviceList = deviceList + "&x_" + i + "=failed"
					}
				} else {

					log.Println(i, "!error conf record!")

				}

			}
		}

		// getting another parameters of PC
		uptime, _ := getUptime()

		CompUptime, _ := time.ParseDuration(fmt.Sprint(uptime))

		PCUp := fmt.Sprint(int(CompUptime.Minutes()))

		// collecting the query string for target
		queryUPTime := "&mfuptime=" + PCUp

		dt := time.Now()

		queryTIME := "&mfdate=" + dt.Format("2006-01-02") + "&mftime=" + dt.Format("15:04") + "&mftimefull=" + dt.Format("15:04:05")

		urlQuery := queryPIN + queryTIME + queryIP + queryMAC + queryUPTime + queryPCName + procList + deviceList

		// checking target is accessable
		target_host, _ := getTargetHost(target_url)

		_, err = checkTarget(target_host)

		// if OK sending query to target, then pause for 'timeout' seconds in config
		if err == nil {
			_, err = sendQuery(target_url, token, urlQuery)
		}

		procList = ""
		queryTIME = ""
		urlQuery = ""

		// if error - print message and waiting 10 seconds before next checking
		if err == nil {
			fmt.Println("")
			log.Println("\n--------------------")

			timer(startTime, timePeriod)

		} else {

			lessInfoErr := strings.Replace(err.Error(), token, "[token]", -1)
			lessInfoErr = strings.Replace(lessInfoErr, pin, "[pin]", -1)

			log.Println(lessInfoErr)
			time.Sleep(err_timeout * time.Second)

		}

	}

}
