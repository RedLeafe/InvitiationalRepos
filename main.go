// Run controlpanel if flag --controlpanel is set, run rover if flag --rover is set, show help if no flags are set

package main

import (
	"flag"
	"log"
	"os"

	cp "tartarus.moon.mine/cmd/controlpanelapi"
	rapi "tartarus.moon.mine/cmd/roverapi"
)

func main() {
	controlpanel := flag.Bool("controlpanel", false, "Run the control panel")
	rover := flag.Bool("rover", false, "Run the rover")
	flag.Parse()

	// if KUBECONFIG environment is not set, error and warn the user
	if os.Getenv("KUBECONFIG") == "" {
		log.Println("KUBECONFIG environment variable not set")
		// set KUBECONFIG to os.Getenv("HOME") + "/.kube/config"
		os.Setenv("KUBECONFIG", os.Getenv("HOME")+"/.kube/config")
	}
	if *controlpanel {
		cp.Controlpanel()
	} else if *rover {
		rapi.Roverapi()
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}
}
