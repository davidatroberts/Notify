package main

import (
	"Notify/Observer"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/gorilla/mux"
	"github.com/olebedev/config"
)

func parseCmdAction(device, action string, cfg *config.Config) (string, []string) {
	// get global cmd
	cmd, err := cfg.String(fmt.Sprintf("%s.cmd", device))
	if err != nil {
		panic("Failed to parse cmd")
	}

	// looking for path to cmd
	path, err := exec.LookPath(cmd)
	if err != nil {
		panic(fmt.Sprintf("Failed to find command: %v\n", err))
	}

	// get the global flags
	gFlags, err := cfg.List(fmt.Sprintf("%s.common_flags", device))
	if err != nil {
		fmt.Printf("No common flags for: %s\n", device)
	}

	// get the action flags
	flags, err := cfg.List(fmt.Sprintf("%s.actions.%s.flags", device, action))
	if err != nil {
		panic("Failed to parse the action flags")
	}

	var allFlags []string
	for _, flag := range gFlags {
		flagStr, ok := flag.(string)
		if ok {
			allFlags = append(allFlags, flagStr)
		}
	}

	for _, flag := range flags {
		flagStr, ok := flag.(string)
		if ok {
			allFlags = append(allFlags, flagStr)
		}
	}

	return path, allFlags
}

func newCmdLineObserver(device, action string, subject *Observer.Subject, cfg *config.Config) *Observer.Observer {
	cmd, flags := parseCmdAction(device, action, cfg)
	return &Observer.Observer{
		Chnl: subject.AddObserver(action),
		Handler: func(event Observer.Event, cmdPath string, flags []string) {
			_, ok := event.(Observer.MessageEvent)
			if ok {
				cmd := exec.Command(cmdPath, flags...)
				if output, err := cmd.CombinedOutput(); err != nil {
					fmt.Printf("%s failed: %s\n", device, output)
				}
			}
		},
		Cmd:   cmd,
		Flags: flags,
	}
}

func handleRequests(subject *Observer.Subject) {
	router := mux.NewRouter()
	router.HandleFunc(
		"/notify/phone/msg",
		func(w http.ResponseWriter, req *http.Request) {
			msg := Observer.MessageEvent{Message: "received"}
			subject.NotifyObservers("text", msg)
		},
	).Methods("POST")

	router.HandleFunc(
		"/notify/phone/call/missed",
		func(w http.ResponseWriter, req *http.Request) {
			// TODO: change this to use CallEvent
			msg := Observer.MessageEvent{Message: "missed call"}
			subject.NotifyObservers("missed", msg)
		},
	).Methods("POST")

	router.HandleFunc(
		"/notify/default",
		func(w http.ResponseWriter, req *http.Request) {
			msg := Observer.MessageEvent{Message: "received"}
			subject.NotifyObservers("default", msg)
		},
	).Methods("POST")

	http.ListenAndServe(":8080", router)
}

func main() {
	// get the command line args
	if len(os.Args) < 2 {
		fmt.Printf("Please enter config file\n")
		os.Exit(1)
	}
	cfg, err := config.ParseYamlFile(os.Args[1])
	if err != nil {
		fmt.Printf("Unable to load config file: %v\n", err)
		os.Exit(1)
	}

	// get the toplevel list of devices
	var devices map[string]interface{}
	devices, err = cfg.Map("")
	if err != nil {
		fmt.Printf("Unable to parse the top level list of devices")
		os.Exit(1)
	}

	// create the subject with single channel
	subject := Observer.NewSubject()

	// iterate through each device and create associated observer(s)
	var observers []*Observer.Observer
	for device := range devices {
		actionMap, err := cfg.Map(fmt.Sprintf("%s.actions", device))
		if err != nil {
			fmt.Printf("Unable to parse actions for %s\n", device)
		}

		for action := range actionMap {
			newobs := newCmdLineObserver(device, action, subject, cfg)
			observers = append(observers, newobs)
		}
	}

	// start the observers
	for _, obs := range observers {
		obs.Process()
	}

	handleRequests(subject)
}
