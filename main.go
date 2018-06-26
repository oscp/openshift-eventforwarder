package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"net"

	"strings"

	"golang.org/x/build/kubernetes/api"
)

// Stream : structure for holding the stream of data coming from OpenShift
type Stream struct {
	Type  string    `json:"type,omitempty"`
	Event api.Event `json:"object"`
}

func main() {
	apiAddr := os.Getenv("OPENSHIFT_API_URL")
	apiToken := os.Getenv("OPENSHIFT_TOKEN")
	syslogServer := os.Getenv("SYSLOG_SERVER")
	syslogProto := strings.ToLower(os.Getenv("SYSLOG_PROTO"))
	syslogTag := strings.ToUpper(os.Getenv("SYSLOG_TAG"))
	ignoreSSL := strings.ToUpper(os.Getenv("IGNORE_SSL"))
	debugFlag := strings.ToUpper(os.Getenv("DEBUG"))

	// enable signal trapping
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c,
			syscall.SIGINT,  // Ctrl+C
			syscall.SIGTERM, // Termination Request
			syscall.SIGSEGV, // FullDerp
			syscall.SIGABRT, // Abnormal termination
			syscall.SIGILL,  // illegal instruction
			syscall.SIGFPE)  // floating point
		sig := <-c
		log.Fatalf("Signal (%v) Detected, Shutting Down", sig)
	}()

	// check and make sure we have the minimum config information before continuing
	if apiAddr == "" {
		// use the default internal cluster URL if not defined
		apiAddr = "https://openshift.default.svc.cluster.local"
		ignoreSSL = "TRUE"
		log.Print("Missing environment variable OPENSHIFT_API_URL. Using default API URL")
	}
	if apiToken == "" {
		// if we dont set it in the environment variable, read it out of
		// /var/run/secrets/kubernetes.io/serviceaccount/token
		log.Print("Missing environment variable OPENSHIFT_TOKEN. Leveraging serviceaccount token")
		fileData, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
		if err != nil {
			log.Fatal("Service Account token does not exist.")
		}
		apiToken = string(fileData)
	}
	if syslogTag == "" {
		// we don't need to error out here, but we do need to set a default if the variable isn't defined
		syslogTag = "OSE"
	}
	if ignoreSSL == "" {
		// we don't need to error out here, but we do need to set a default if the variable isn't defined
		ignoreSSL = "FALSE"
	}
	if debugFlag == "" {
		// we don't need to error out here, but we do need to set a default if the variable isn't defined
		debugFlag = "FALSE"
	}
	if (syslogProto == "") || (syslogProto == "tcp") || (syslogProto == "udp") {
		// we don't need to error out here, but we do need to set a default if the variable isn't defined
		if syslogProto == "" {
			syslogProto = "udp"
		} else {
			log.Printf("Will use %s for syslog protocol", syslogProto)
		}

	} else {
		log.Fatalf("SYSLOG_PROTO must be either blank, or tcp or udp not %s", syslogProto)
	}

	// Setup syslog connection only if syslogServer is defined
	if syslogServer != "" {
		sysLog, err := syslog.Dial(syslogProto, syslogServer,
			syslog.LOG_WARNING|syslog.LOG_DAEMON, syslogTag)
		if err != nil {
			log.Printf("Error connecting to %s", syslogServer)
			log.Fatal(err)
		} else {
			log.Printf("Event Forwarder configured to send all events to %s using tag %s", syslogServer, syslogTag)
			if debugFlag == "TRUE" {
				// dump the data to stdout AND syslog for testing.
				log.SetOutput(io.MultiWriter(sysLog, os.Stdout))
				ipAddr, _ := net.LookupHost(syslogServer)
				log.Printf("Connecting to IP address: %v\n", ipAddr)
			} else {
				log.SetOutput(sysLog)
			}
		}
	} else {
		log.Print("SYSLOG_SERVER environment variable not set. Sending all output to console.")
	}

	// setup ose connection
	var client http.Client
	if ignoreSSL == "TRUE" {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client = http.Client{Transport: tr}
	} else {
		client = http.Client{}
	}
	req, err := http.NewRequest("GET", apiAddr+"/api/v1/events?watch=true", nil)
	if err != nil {
		log.Fatal("## Error while opening connection to openshift api", err)
	}
	req.Header.Add("Authorization", "Bearer "+apiToken)

	for {
		resp, err := client.Do(req)

		if err != nil {
			log.Println("## Error while connecting to:", apiAddr, err)
			time.Sleep(5 * time.Second)
			continue
		}

		streamStart := time.Now()
		reader := bufio.NewReader(resp.Body)

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				log.Println("## Error reading from response stream.", err, line)
				resp.Body.Close()
				break
			}

			event := Stream{}
			decErr := json.Unmarshal(line, &event)
			if decErr != nil {
				log.Println("## Error decoding json.", err)
				resp.Body.Close()
				break
			}

			// Kubernetes sends all data from ETCD, we only want the logs since the stream started
			if event.Event.LastTimestamp.Time.After(streamStart) {
				fmt.Printf("%v | Project: %v | Name: %v | Kind: %v | Reason: %v | Message: %v\n",
					event.Event.LastTimestamp.Format(time.RFC3339),
					event.Event.Namespace, event.Event.InvolvedObject.Name,
					event.Event.Kind, event.Event.Reason, event.Event.Message)
			}
		}
	}
}
