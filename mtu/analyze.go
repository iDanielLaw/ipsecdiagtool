package mtu

import (
	"code.google.com/p/gopacket/examples/util"
	"log"
	"time"
	"strconv"
	"math/rand"
	"net"
)

//Package setup
var appID int
var srcIP net.IP
var destIP net.IP
var destPort int
var incStep int

//Package internal temp. variables
var currentMTU int
var mtuOKchan (chan int)

//Setup configures the MTU-daemon with the necessary information to
//determine the MTU between two nodes. At some point it will likely get
//it's information automatically from a central config reader within the
//application. But it's still useful if you want to use the MTU-detection
//directly in your application.
//If you set the application ID 0, a random one will be automatically generated.
func Setup(applicationID int, sourceIP string, destinationIP string, destinationPort int, incrementationStep int) {
	if(applicationID == 0){
		rand.Seed(time.Now().UnixNano()) //Seed is required otherwise we always get the same number
		appID = rand.Intn(100000)
	} else {
		appID = applicationID
	}
	srcIP = net.ParseIP(sourceIP)
	destIP = net.ParseIP(destinationIP)
	destPort = destinationPort
	incStep = incrementationStep
	currentMTU = 500 //Starting MTU
}

//Analyze determines the ideal MTU (Maximum Transmission Unit) between two nodes
//by sending increasingly big packets between them. Analyze determine the MTU
//exactly once and return the value of the ideal MTU. To continuously determine
//the MTU you should run [not implemented yet].
func Analyze() {
	defer util.Run()()
	setDefaultValues()
	log.Println("Analyzing MTU..")

	//Setup a channel for communication with capture
	mtuOKchan = make(chan int)  // Allocate a channel.

	//Capture all traffic via goroutine in separate thread
	go startCapture("tcp port " + strconv.Itoa(destPort))

	//Fire first packet to determine MTU. Later this should be done at
	//certain times or via outside input in form of a cronjob.
	time.Sleep(1000 * time.Millisecond)
	go findMTU()

}

//setDefaultValues is run when the user doesn't configure the MTU package via Setup().
func setDefaultValues() {
	if destPort == 0 {
		Setup(0, "127.0.0.1","127.0.0.1",22, 100)
		log.Println("Setting default values, because Analyze() was called before or without Setup()")
	}
}

func findMTU(){
	//1. Initiate MTU discovery by sending first packet.
	sendPacket(srcIP, destIP, destPort, currentMTU, "MTU?")

	//2. Either we get a message from our mtu channel or the timeout channel will message us after 10s.
	for {
		//2.1 Setting up the timeout channel
		//http://blog.golang.org/go-concurrency-patterns-timing-out-and
		timeout := make(chan bool, 1)
		go func() {
			time.Sleep(10 * time.Second)
			timeout <- true
		}()

		select {
			case <-mtuOKchan:
				log.Println("Main Routine notified about state in subroutine.")
			case <-timeout:
				log.Println("Timeout has occured. We've steped over the MTU!")
				//TODO: break out of loop.
		}
	}
	//3. Report MTU
}
