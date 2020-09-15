package main

import ( "net" ; "fmt" ; "bufio" ; "strings" ; "os" ; "sync" )

//var peerMap = make(map[string] net.Conn)
var peers []net.Conn
var mutex sync.Mutex

var messagesSent =  make(map[string]bool)

 // TODO: populate an empty set with the messages it has sent 
 // TODO: checks strings from other connections or inputed,  if it has already been sent. if so, dont do anything.
 // - otherwise add it to messages MessagesSend(map) and send it to all its connections (Remember concurency control)
 // TODO: if message added to MessagesSend(), then print it to the console ....


func handleConnection(conn net.Conn) {
	defer conn.Close()

	// adding conn to the networklist
	peers = append(peers, conn)

	// reads on incomming
	reader := bufio.NewReader(conn)

	// other end of conn
	otherEnd := conn.RemoteAddr().String()

  for {
		msg, err := reader.ReadString('\n')
		if (err != nil) {
			fmt.Println("Ending session with " + otherEnd)
			return
		}
		if !messagesSent[msg] {
			fmt.Print(string(msg))
			go userOutput(msg)
			messagesSent[msg] = true
		}
	}
}

// reads input from commandline
func userInput(){
	reader := bufio.NewReader(os.Stdin)
	for {
		msg, _ := reader.ReadString('\n')
		mutex.Lock() 
		for _, peer := range peers {
			if !messagesSent[msg] {
				peer.Write([]byte(msg))
			}
		}
		messagesSent[msg] = true
		mutex.Unlock()
	}
}

// forwards message if not send
func userOutput(msg string) {
	for _, peer := range peers {
		peer.Write([]byte(msg))
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Write IP and port")
	fmt.Print(">")

	// address which to dial
	peerAddr, _ := reader.ReadString('\n')
	peerAddr = strings.TrimSpace(peerAddr) // returns peerAdrr without trailing suffix

	// Calling the peerAddr
	conn, err := net.Dial("tcp", peerAddr)

	go userInput()

	// Checks if there is an error when dialing the conncection
	if err != nil {
		fmt.Println("The peer has not been found")
		fmt.Println("Creating network...")
	} else {
		go handleConnection(conn)
	}

	// Opening a port with a random port
	ln, _ := net.Listen("tcp", ":" )
	_, port, _ := net.SplitHostPort(ln.Addr().String()) //splits the port "host:port"


	//Listing and accepting incoming clients
	fmt.Println("Listening on port: " + port)
	for {
		conn, _ := ln.Accept()
		fmt.Println("Got a connection...")
		go handleConnection(conn)
	}
}