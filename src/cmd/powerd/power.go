package main

import (
	"fmt"
	"http"
	"json"
	"flag"
	"exec"
	"os"
	"io/ioutil"
	"strings"
	daemon "flunky/daemon"
)


type outletNode struct {
	Address string
	Outlet  string
}

var resources map[string]outletNode
var powerDaemon *daemon.Daemon
var fileDir string

func DumpCall(w http.ResponseWriter, req *http.Request) {
        powerDaemon.DaemonLog.DebugHttp(req)
	req.ProtoMinor = 0
	err := powerDaemon.AuthN.HTTPAuthenticate(req, true)
        if err != nil{
           powerDaemon.DaemonLog.LogError("Unauthorized request for dump.", err)
	   w.WriteHeader(http.StatusUnauthorized)      
	   return
        }
	tmp, err := json.Marshal(resources)
	powerDaemon.DaemonLog.LogError("Cannot Marshal power resources", err)
	_, err = w.Write(tmp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	powerDaemon.DaemonLog.Log("Serviced request for data dump")
}

func printCmd(nodes []string, cmd string){
    switch(cmd){
       case "on" : powerDaemon.DaemonLog.Log(fmt.Sprintf("%s have been turned %s", nodes, cmd))
                   break
       case "off": powerDaemon.DaemonLog.Log(fmt.Sprintf("%s have been turned %s", nodes, cmd))
                   break
       case "reboot":powerDaemon.DaemonLog.Log(fmt.Sprintf("%s have been %sed", nodes, cmd))
                    break
      }
      return
}
func command(w http.ResponseWriter, req *http.Request){
    var nodes []string
    req.ProtoMinor = 0
    powerDaemon.DaemonLog.DebugHttp(req)
    err := powerDaemon.AuthN.HTTPAuthenticate(req, true)
    if err != nil{
       powerDaemon.DaemonLog.LogError("Access not permitted.", err)      
       w.WriteHeader(http.StatusUnauthorized)
       return
    }
    dex := strings.Split(req.RawURL, "/")
    cmd := dex[2]
    switch(cmd){
       case "on", "off", "reboot" : break
       default: powerDaemon.DaemonLog.LogError(fmt.Sprintf("%s command not supported", cmd), os.NewError("unsupported"))
                w.WriteHeader(http.StatusNotFound) 
                return
		break
     }
    body, err := powerDaemon.ReadRequest(req)
    powerDaemon.DaemonLog.LogError("Unable to read request", err)

    err = json.Unmarshal(body, &nodes)
    powerDaemon.DaemonLog.LogError(fmt.Sprintf("Unable to unmarshal nodes for %s command.",cmd), err)

    for _, value := range nodes {
        if _, ok := resources[value]; ok {
            go func(value string) {
            err = exec.Command("./powerCont.sh", resources[value].Address, "admn", "admn", cmd, resources[value].Outlet).Run()
            powerDaemon.DaemonLog.LogError("Failed to run powerCont.sh in rebootList.", err)
	    }(value)
	}
    }
    printCmd(nodes, cmd) 
}


func statusList(w http.ResponseWriter, req *http.Request) {
        powerDaemon.DaemonLog.DebugHttp(req)
	powerDaemon.DaemonLog.LogDebug("Retreiving status for list given by client.")
	var nodes []string
	outletStatus := make(map[string]string)
	req.ProtoMinor = 0
        err := powerDaemon.AuthN.HTTPAuthenticate(req, true)
        if err != nil{
           powerDaemon.DaemonLog.LogError("Access not permitted.", err)
	   w.WriteHeader(http.StatusUnauthorized)      
           return
        }
	body, err := powerDaemon.ReadRequest(req)
	powerDaemon.DaemonLog.LogError("Could not read request", err)

	err = json.Unmarshal(body, &nodes)
	powerDaemon.DaemonLog.LogError("Unable to unmarshal nodes to be turned off.", err)

	for _, value := range nodes {
		_, ok := outletStatus[value]
		_, ok2 := resources[value]

		if !ok && ok2 {
			someBytes, err := exec.Command("./powerCont.sh", resources[value].Address, "admn", "admn", "status").Output()
			powerDaemon.DaemonLog.LogError("Failed to execute powerCont.sh and get out put in power status request.", err)

			tmpStatusLines := strings.Split(string(someBytes), "\n")

			for i := 18; i < 42; i++ {
				tmpStatusFields := strings.Split(tmpStatusLines[i], " ")

				for _, value2 := range nodes {
					if _, ok3 := resources[value2]; ok3 && ok2 {
						if resources[value2].Address == resources[value].Address && resources[value2].Outlet == tmpStatusFields[3] {
							outletStatus[value2] = tmpStatusFields[13]
						}
					}
				}
			}
		}
	}

	jsonStat, err := json.Marshal(outletStatus)
	powerDaemon.DaemonLog.LogError("Unable to marshal outlet status response.", err)

	_, err = w.Write(jsonStat)
	powerDaemon.DaemonLog.LogError("Unable to write outlet status response.", err)
	
}

func main() {
	flag.Parse()
	var err os.Error

	powerDaemon, err = daemon.New("power")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	user, pass, _ := powerDaemon.AuthN.GetUserAuth()
	err = powerDaemon.AuthN.Authenticate(user, pass, true)
	if err != nil{
 	   fmt.Println(fmt.Sprintf("You dont have permissions to start %s daemon.", powerDaemon.Name))
	   os.Exit(1)
	}

	powerDB, err := ioutil.ReadFile(daemon.FileDir + "power.db")
	powerDaemon.DaemonLog.LogError("Unable to open power.db for reading.", err)

	err = json.Unmarshal(powerDB, &resources)
	powerDaemon.DaemonLog.LogError("Failed to unmarshal data read from power.db file.", err)

	http.HandleFunc("/dump", DumpCall)
	http.HandleFunc("/command/", command)
	http.HandleFunc("/status", statusList)
	powerDaemon.DaemonLog.Log(fmt.Sprintf("%s started on %s", powerDaemon.Name, powerDaemon.URL))
	err = powerDaemon.ListenAndServe()
	if err != nil {
		os.Exit(1)
	}
}
