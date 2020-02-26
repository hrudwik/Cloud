package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/biorad/unitynext/S3Upload_InvokeProcessUtility/model"
	"github.com/biorad/unitynext/S3Upload_InvokeProcessUtility/processor"
	"github.com/biorad/unitynext/S3Upload_InvokeProcessUtility/services"
	"github.com/biorad/unitynext/S3Upload_InvokeProcessUtility/utility"

	"github.com/takama/daemon"
	"gopkg.in/yaml.v2"
)

var stdlog, errlog *log.Logger

// Service has embedded daemon
type Service struct {
	daemon.Daemon
}

// Manage by daemon commands or run the daemon
func (service *Service) Manage() (string, error) {

	usage := "Usage: S3Upload_InvokeProcessUtility install | remove | start | stop | status"

	// if received any kind of command, do it
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	queuename := utility.Cfg.Queue.Name
	destLocation := utility.Cfg.DestinationLocation.Path
	configFileName := utility.Cfg.ConfigFile.Location + string(os.PathSeparator) + utility.Cfg.ConfigFile.Name

	//Validate provided queuename and destination Location to save files form S3
	services.ValidateQueueNameAndDestinationPath(queuename, destLocation)

	//Create the session
	sess, SQSConn, downloader, QueueURL := services.CreateSession(queuename)

	log.Printf("\r\n \r\n*** New Session Started ***\r\n")
	log.Printf("Queue: %q, Destination Directory: %q.\r\n", queuename, destLocation)

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	port := ":" + utility.Cfg.Service.Port
	// Set up listener for defined host and port
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return "Possibly was a problem with the port binding", err
	}

	// set up channel on which to send accepted connections
	listen := make(chan net.Conn, 100)
	go acceptConnection(listener, listen)

	ticker := time.NewTicker(utility.Cfg.Service.Timer * time.Millisecond * 1000)
	tickerDone := make(chan bool)

	// loop work cycle with accept connections or interrupt
	// by system signal
	for {

		select {
		case <-ticker.C:
			stdlog.Println("fakeProcess called")
			go mainProcess(queuename, destLocation, configFileName, sess, SQSConn, downloader, QueueURL)
		case <-tickerDone:
			return "Timer stopped", nil
		case killSignal := <-interrupt:
			stdlog.Println("Got signal:", killSignal)
			stdlog.Println("Stoping listening on ", listener.Addr())
			listener.Close()
			if killSignal == os.Interrupt {
				return "Daemon was interrupted by system signal", nil
			}
			return "Daemon was killed", nil
		}
	}

	// never happen, but need to complete code
	return usage, nil
}

// Accept a client connection and collect it in a channel
func acceptConnection(listener net.Listener, listen chan<- net.Conn) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		listen <- conn
	}
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func mainProcess(queuename, destLocation, configFileName string, sess *session.Session, SQSConn *sqs.SQS,
	downloader *s3manager.Downloader, QueueURL *string) {
	startTime := time.Now()
	processor.Process(queuename, destLocation, configFileName, sess, SQSConn, downloader, QueueURL)

	defer timeTrack(startTime, "mainProcess")

}

// Force Go program to use all cores of the system to parallel processing.
func init() {
	stdlog = log.New(os.Stdout, "", 0)
	errlog = log.New(os.Stderr, "", 0)
}

func processFatalError(err error) {
	fmt.Println(err)
	fmt.Println(errlog) // Log to system Log  /var/log/syslog
	os.Exit(2)
}

func readFile(cfg *model.Config) {
	f, err := os.Open(utility.ConfigFile)
	if err != nil {
		processFatalError(err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		processFatalError(err)
	}
}

func main() {
	fmt.Println("In Main")
	readFile(&utility.Cfg)
	fmt.Println(utility.Cfg)
	logFileName := utility.Cfg.LogFile.Location + string(os.PathSeparator) + utility.Cfg.LogFile.Name

	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("Logging initalized for service " + utility.Name)

	srv, err := daemon.New(utility.Name, utility.Description, utility.Dependencies...)
	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}
	service := &Service{srv}
	status, err := service.Manage()
	if err != nil {
		errlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(status)

}
