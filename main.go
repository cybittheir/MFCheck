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
	version string
	build   string
)

func getIP() string {

	ifaces, err := net.Interfaces()

	if err != nil {
		fmt.Println(err)
	}

	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Println(err)
		}
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

				strIP := strconv.Itoa(t0) + "." + strconv.Itoa(t1) + "." + strconv.Itoa(t2) + "." + strconv.Itoa(t3)

				if t0 == 192 && t1 == 168 {
					//					fmt.Println(ip)
					//					fmt.Printf(strIP)
					return strIP
				}

			}

		}
	}
	return ""

}

var getTickCount = syscall.NewLazyDLL("kernel32.dll").NewProc("GetTickCount64")

func isProcRunning(batchPath string, batchName string, name string) (bool, error) {

	fmt.Print(".")

	cmd := exec.Command(batchPath+batchName, name)
	cmd.Dir = batchPath

	out, err := cmd.Output()

	if err != nil {
		fmt.Print(".")
		return false, err
	}

	if bytes.Contains(out, []byte(name)) {
		fmt.Print(".")
		return true, nil
	}
	return false, nil
}

func getUptime() (time.Duration, error) {
	ret, _, err := getTickCount.Call()
	if errno, ok := err.(syscall.Errno); !ok || errno != 0 {
		return time.Duration(0), err
	}
	return time.Duration(ret) * time.Millisecond, nil
}

func main() {

	// Open our jsonFile

	jsonFile, err := os.Open("conf.json")

	// if we os.Open returns an error then handle it

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Successfully Opened conf.json")

	// defer the closing of our jsonFile so that we can parse it later on

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var confResult map[string]map[string]string
	var checkProc map[string]map[string]map[string]string

	json.Unmarshal([]byte(byteValue), &confResult)
	json.Unmarshal([]byte(byteValue), &checkProc)

	var emptyConfig bool

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
		os.Exit(0)
	}

	url := fmt.Sprintf("%s", confResult["connect"]["url"])
	token := fmt.Sprintf("%s", confResult["connect"]["token"])
	pin := fmt.Sprintf("%s", confResult["connect"]["pin"])
	batchName := fmt.Sprintf("%s", confResult["connect"]["batch"])
	batchPath := fmt.Sprintf("%s", confResult["connect"]["path"])

	var procList string

	if len(checkProc["check"]["process"]) > 0 {

		fmt.Printf("%s", "Check Processes ")

		for i, v := range checkProc["check"]["process"] {

			if v != "" {
				forCheck := fmt.Sprintf("%s", v)

				isRunning, _ := isProcRunning(batchPath, batchName, forCheck)

				if isRunning {
					forCheckParam := fmt.Sprintf("%s", i)
					procList = procList + "&" + forCheckParam + "=1"
				}
			}
		}

		fmt.Println(" Finished")

	} else {

		fmt.Println("Warning! Section [check][process] in conf.json file is needed for cheching running processes")
	}

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	uptime, err := getUptime()

	CompUptime, err := time.ParseDuration(fmt.Sprint(uptime))

	PCUp := fmt.Sprint(int(CompUptime.Minutes()))

	queryPIN := "&mfpin=" + pin
	queryIP := "&mfip=" + getIP()
	queryUPTime := "&mfuptime=" + PCUp
	queryPCName := "&mfname=" + hostname

	dt := time.Now()

	queryTIME := "&mfdate=" + dt.Format("2006-02-01") + "&mftime=" + dt.Format("15:04:05")

	urlQuery := queryPIN + queryTIME + queryIP + queryUPTime + queryPCName + procList

	fmt.Println("Sending info to server:" + urlQuery)

	// Make HTTP GET request
	response, err := http.Get(url + "?UID=" + token + urlQuery)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// Copy data from the response to standard output
	n, err := io.Copy(os.Stdout, response.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n-----------------")
	log.Println("Number of bytes copied to STDOUT:", n)

}
