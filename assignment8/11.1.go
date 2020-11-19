package main

import ( 
	"net" 
	"fmt"  
	"bufio"  
	"strings"  
	"os"  
	"encoding/gob"  
	"io"  
	"sync"  
	"sort"  
	"strconv"  
	"time" 
	"math/big"
	RSA "./RSA"
)

type Message struct {
	Connections []string
	Transaction *Transaction
	Broadcast string
	Ledger *Ledger
}

type Ledger struct {
	Accounts map[string]int
	Connections [] string
	Transactions[] *Transaction
	Sequencer string
	Phase int
	NewBlock *Block
}

type Block struct {
	BlockNr int
	IDlist map[int]string
	Transactions []*Transaction
	Signature string
}

type Transaction struct {
	Id string
	Amount int
	From string
	To string
	Signature string
}

var connections []net.Conn

var connectedPeers []string

var myIp string

var localMsg *Message

var ledger *Ledger

var mutex sync.Mutex

var isBroadcasted map[string]bool

var transactionIsUsed map[string]bool

var n,e *big.Int

var LocalPK *RSA.PublicKeyPair

var SequencerKeyPair *RSA.PublicKeyPair

var PKlist[] *RSA.PublicKeyPair

var PkString string

var isDesignatedSequencer bool = false //boolean detimrming whether the peer is sequencer or not

var informed bool = false //boolean which detirmines if the peer knows the sequencer

var tempOrder int = 0

var LocalBlockNumber int = 0

// THIS A NEW FEATURE
var BlocksRecieved map[int]*Block

var blockUpdated bool = false

func makeMessage() *Message {
	Message := new(Message)
	return Message
}

func makeLedger() *Ledger {
	Ledger := new(Ledger)
	Ledger.Accounts = make(map[string]int)
	return Ledger
}

func makeTransaction() *Transaction {
	Transaction := new(Transaction)
	return Transaction
}

func send(msg *Message, conn net.Conn) {
	enc := gob.NewEncoder(conn) // encodes connection
	enc.Encode(msg) // encodes and sends array
}

// THIS IS REDACTED
func updateBlock() {
	time.Sleep(10000 *time.Millisecond)
	fmt.Println("Block is now updated")
	var update = makeMessage()
	var sss = strconv.Itoa(LocalBlockNumber) + ledger.NewBlock.IDlist[0]
	var s = RSA.Sign([]byte(sss))
	var ss = convertBigIntToString(s)
	ledger.NewBlock.BlockNr = LocalBlockNumber
	ledger.NewBlock.Signature = ss
	update.Ledger = ledger
	fmt.Println("This is the ledger: ", update.Ledger.NewBlock.IDlist)
	fmt.Println(update.Ledger.NewBlock.Signature, "this is s")
	fmt.Println(ledger.NewBlock.BlockNr, "this is the block to be send")
	for _, conn := range connections{
		fmt.Println("send to peers")
		go send(update, conn)
	}
	ledger.NewBlock.IDlist = make(map[int]string)
	tempOrder = 0
	LocalBlockNumber++
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// struct for recieving array
	message := &Message{}

  for {
		dec := gob.NewDecoder(conn) //decodes on the connection
		err := dec.Decode(message) // decodes the stringarray


		// checks if Connections list is not emty - if not it will be treated as a new peer
		if ( message.Connections != nil && ledger.Phase == 1) {
			checkForConnections(message) // check if the peer is connected
			connectToNewPeers() // connect the peers <10
			broadcast(myIp) // broadcast presence
		}

		// when message.ledger not empty, this will checked if the localledger aldready have the same pk. If not, this pk will be added to the local ledger
		if(message.Ledger != nil) {
			addNewPk(message)
			// if sequencer is now empty and the peer is not sequencer it will add the sequencer -- This is phase 1
			if message.Ledger.Phase == 1 {
				if message.Ledger.Sequencer != "" && !isDesignatedSequencer { 
				ledger.Sequencer = message.Ledger.Sequencer
				SequencerKeyPair = sortKeyPair(ledger.Sequencer)
				fmt.Println("Sequencer found...")
				}
					// This part forward the sequencer, if someone in the network is uninformed
				if(!informed) { 
					sMessage := &Message{}
					/*sMessage.Ledger = ledger
					sMessage.Ledger.Phase = ledger.Phase// if the peer is informed about the squencer, it will not send the sequencer again
					sMessage.Ledger.Sequencer = ""*/
					sMessage.Ledger = ledger
					informed = true
					go send(sMessage, conn)
				}
			}
			// In phase 2 it will first check that the sequencer has informed this, and then it will read whether the 
			// there is a signed block or not
			// THIS IS A NEW FEATURE
			if ( message.Ledger.Phase == 2 ) {
				ledger.Phase = 2 // updates the local ledger with the new phase
				// // if the message contains a block number, this will be added
				if !isDesignatedSequencer {
				fmt.Println("localBlock:", LocalBlockNumber)
				fmt.Println("Block:", message.Ledger.NewBlock.BlockNr )
				fmt.Println("Signature", message.Ledger.NewBlock.Signature )
				if ( BlocksRecieved[message.Ledger.NewBlock.BlockNr] == nil && message.Ledger.NewBlock.Signature != "" ) { // if the signature is not empty, this means that there a signed transactions
					fmt.Println("this far")
					LocalBlockNumber = message.Ledger.NewBlock.BlockNr
					BlocksRecieved[LocalBlockNumber] = message.Ledger.NewBlock // if the block can be verified, it will execute alle the transactions
					for i := 0; i<LocalBlockNumber; i++ {
						fmt.Println("this far2")
						if(BlocksRecieved[i] == nil ) {
							fmt.Println("Blocks out of sync, cannot execute before others have been recieved")
							return 
						} else if verifyBlock(BlocksRecieved[i]) {
								executeAllNewTransactions(BlocksRecieved[i])
						}
					}
				} else {
					fmt.Println("blockNumber is out of order. Block rejected")
				}
			}
			}
		}
	


		// checks if transactions list is not emty
		if message.Transaction != nil {
			//fmt.Println("thisis peers:", connectedPeers)
			checkIfUserIsKnown(message, conn) // if the ip is know the PublicKey is added to the local accounts list
			incommingTransaction(message.Transaction) // handles incomming transaction
		}
				
		if (err == io.EOF) {
			fmt.Println("Connection has been closed by " + conn.RemoteAddr().String())
			return
		}

		// checks if the connections has been broadcasted
		if message.Broadcast != "" {
			mutex.Lock()
			_, exists := isBroadcasted[message.Broadcast]
			if !exists {
				isBroadcasted[message.Broadcast] = false
				fmt.Println("Is present: " + message.Broadcast)
				isBroadcasted[message.Broadcast] = true
				broadcast(myIp)
			} 
			mutex.Unlock()
		}
	} 
}

// checks connections of sender if it is known
func checkIfUserIsKnown(m *Message, conn net.Conn){
	exists, _ := findString(connectedPeers, conn.RemoteAddr().String())
	if exists {
		addNewPk(m)
		return 
	} else {
		return
	}
}

// add users pk to local ledger
func addNewPk(m *Message) {
	for key, element := range m.Ledger.Accounts {
		if _, exists1 := ledger.Accounts[key]; !exists1 {
			ledger.Accounts[key] = element
			sortKeyPair(key) // sorts the publicKey
			//fmt.Println("this the key: ",key)
		}
	}
	return
}

// handles outgoing transactions
func handleTransaction(t *Transaction) {
	mutex.Lock()
	exists := transactionIsUsed[t.Id]
	if exists {
		mutex.Unlock()
		return
	} else {
		// might be insufficient
		localMsg.Transaction = t
		///
		sendTransaction(t)
	}
	mutex.Unlock()
	return
}

// handles incomming transactions
func incommingTransaction(t *Transaction) {
	mutex.Lock()
	exists := transactionIsUsed[t.Id]
	if exists {
		mutex.Unlock()
		return
	} else {
		go sendTransaction(t)
		verifyTransaction(t)
		transactionIsUsed[t.Id] = true
	}
	mutex.Unlock()
	return 
}

func executeTransaction(t *Transaction) {
	fmt.Println("------------------------TRANSACTION RECIEVED----------------------------------")
	fmt.Println("[ ID:", t.Id, ": a transaction has been recieved from ", t.From, "to ", t.To, "with the amount of", t.Amount, "]")
	fmt.Println("------------------------------------------------------------------------------")

	// checks if there is an account with t.From                                  
	if worth1, exists1 := ledger.Accounts[t.To]; !exists1 { // checks if the account exists
		fmt.Println("account is not recognizable")
	} else if ledger.Accounts[t.To] - t.Amount < 0 { // checks if ballance becomes negative if the amount is withdrawn
		fmt.Println("Account does not have coverage")
	} else {
		ledger.Accounts[t.From] = worth1 - t.Amount
		ledger.Accounts[t.To] += t.Amount
	}
}

// THIS IS NEW FEATURE 
func executeAllNewTransactions(m *Block ) {
	fmt.Println("blocks has been recieved")
	tempIDList := make([]string,100)
	for i := 0; i<len(m.Transactions); i++ {
		tempIDList = append(tempIDList, m.IDlist[i])
	}
	for _, key := range ledger.Transactions {
		exists, _ := findString(tempIDList, key.Id)
		fmt.Println("this key might exist", key)
		if( exists ){
			executeTransaction(key)
		}
	}
}

// Sends a transactions to all peers
func sendTransaction(t *Transaction) {
	transactionToSend := &Message{}
	transactionToSend.Transaction = t
	for _, conn := range connections{
		go send(transactionToSend, conn)
	}
	transactionIsUsed[t.Id] = true
	return 
}

// sorts and checks peers
func checkForConnections(msg *Message) {
	mutex.Lock()
	sort.Strings(msg.Connections)
	for _, peer := range msg.Connections{
		exists, _ := findString(localMsg.Connections, peer)
		if !exists {
			localMsg.Connections = append(localMsg.Connections, peer)
			localMsg.Ledger = ledger
			addNewPk(msg)
		}
	}
	mutex.Unlock()
}

// connect to newcommers
func connectToNewPeers() {
	mutex.Lock()
	if len(connectedPeers)  > 10 { //if legnth of peers are greater than ten it will not connect to mere peers
		mutex.Unlock()
		return
	}
	if len(connectedPeers)  < 10 {
		for _, peer := range localMsg.Connections{
			connected, _ := findString(connectedPeers, peer)
			if !connected {
				conn, _ := net.Dial("tcp", peer)
				connections = append(connections, conn)
				connectedPeers = append(connectedPeers, conn.RemoteAddr().String())
			}
		}
	}
	mutex.Unlock()
	return 
}

// help function for comparing a string to an array
func findString(list []string, p string) (bool, int){
	for ind, l := range list{
		if l == p {
			return true, ind
		}
	}
	return false, -1
}

// sorts the PublicKeys from other connections and them to localPK list in a ledger containing Pks.
func sortKeyPair(pk string) *RSA.PublicKeyPair {
	u := new(RSA.PublicKeyPair)
	split := strings.Split(pk, ",")
	//fmt.Println(pk)
	n2 := new(big.Int)
	e2 := new(big.Int)
	n2 = stringToBigInt(split[0])
	e2 = stringToBigInt(split[1])
	//fmt.Println("n1:", n2)
	//fmt.Println("e1:", e2)
	u.N = n2
	u.E = e2
	PKlist = append(PKlist, u)
	//fmt.Println(PKlist)
	return u
}

// broadcast ip to all peers 
func broadcast(msg string) {
	isBroadcasted[msg] = false
	broadcastMsg := &Message{}
	broadcastMsg.Broadcast = msg
	for _, conn := range connections{
		go send(broadcastMsg, conn)
	}
}

// sets up the sequencer
func initateSequencer() {
	isDesignatedSequencer = true
	RSA.KeyGen(1024) // generates new keyPair
	SequencerKeyPair = RSA.GetKeyPair() // initiates this as keypair
	var e1 = convertBigIntToString(SequencerKeyPair.E)
	var n1 = convertBigIntToString(SequencerKeyPair.N)
	ledger.Sequencer = n1 + "," + e1
	fmt.Println("hello")
	ledger.NewBlock = new(Block)
	ledger.NewBlock.IDlist = make(map[int]string)
	ledger.NewBlock.BlockNr = LocalBlockNumber
	time.Sleep(20000 *time.Millisecond) // gives the peers 20 seconds to connect to the network
	ledger.Phase = 2
	localMsg.Ledger = ledger // updates the ledger in the localMessage 
	for _, conn := range connections{
		go send(localMsg, conn) // tells all its connections that it now stage two
	}
}

// dial up peer
func dial(addr string) {
	conn, err := net.Dial("tcp", addr)
	// Checks if there is an error when dialing the conncection and connecting if possible
	if err != nil {
		fmt.Println("The peer has not been found")
		fmt.Println("Using...", myIp, "...Instead")
		initateSequencer()
	} else {
		connections = append(connections, conn)
		connectedPeers = append(connectedPeers, conn.RemoteAddr().String())
		go handleConnection(conn)
	}
}

// listening for-loop
func listeningForConnections(ln net.Listener) {
	defer ln.Close()
	for {
		incomming, _ := ln.Accept()
		connections = append(connections, incomming)
		// addPk to localMessage
		localMsg.Ledger = ledger
		fmt.Println("Got a connection... with" + incomming.RemoteAddr().String())
		fmt.Println("Wait for transactions...")
		go send(localMsg, incomming)
		go handleConnection(incomming)
	}
}

// requests the peer for a transaction
func requestTransaction() {
	time.Sleep(12000 *time.Millisecond) // waits 20 seconds and then requests for a transaction
	// promt receivers with their following public kyes 
	fmt.Println("Please enter a transaction by following order: to,amount.")
	fmt.Println("------------------------------------------------------")
	fmt.Println("The disired reciever, to, can be chosen by typing the index on the left:")
	fmt.Println("------------------------------------------------------")
	

	var localMap = make(map[int]string)
	var i = 0

	for key, _ := range ledger.Accounts {
		localMap[i]	= key
		fmt.Println(i,": ", key)
		fmt.Println("------------------------------------------------------")
		i++
	}

	fmt.Println("->")
	
	var toSign string

	reader := bufio.NewReader(os.Stdin)
	trans, _ := reader.ReadString('\n')

	split := strings.Split(trans, ",")

	
	to, err := strconv.Atoi(strings.Trim(split[0], "\r \n"))
	amount, err := strconv.Atoi(strings.Trim(split[1], "\r \n")) // string to int
	if err != nil {
		fmt.Println("try again")
	} else {
		t := makeTransaction()
		t.From = PkString
		t.To = localMap[to]
		t.Amount = amount
		t.Id = myIp + t.From + t.To

		toSign = t.From + t.To + strconv.Itoa(amount)

		bigIntSign := RSA.Sign([]byte(toSign))

		bigSignString := convertBigIntToString(bigIntSign)

		t.Signature = bigSignString
		fmt.Println("----------------- TRANSACTION MADE ---------------")
		fmt.Println("[From: ", t.From, "To: ", t.To, "amount: ", t.Amount, "id: ", t.Id)
		fmt.Println("--------------------------------------------------")
		ledger.Transactions = append(ledger.Transactions, t)
		go handleTransaction(t)
	}
	
}

// executes transaction if the results come back true
func verifyTransaction(inc *Transaction) {
	if verifySignature(inc) {
		if(isDesignatedSequencer) {
			fmt.Println(inc.Id)
			ledger.NewBlock.IDlist[tempOrder] = inc.Id
			ledger.NewBlock.Transactions = append(ledger.NewBlock.Transactions, inc)
			tempOrder++
		}
		ledger.Transactions = append(ledger.Transactions, inc)
		return 
	} else {
		return
	}
}

// prepares the signature to be verified
func verifySignature(user *Transaction) bool {
	var testString string
	var strAmount string
	var testStringHash []byte
	signature := new(big.Int)
	signature = stringToBigInt(user.Signature) // recieved signature converted to big.int
	IncPK := new(RSA.PublicKeyPair) 
	strAmount = strconv.Itoa(user.Amount) 
	testString = user.From + user.To + strAmount // creates the string to get hashed
	testStringHash = []byte(testString) // created string from information
	testHash := RSA.Hash(testStringHash) // hashes the string
	//fmt.Println("this is the testHash ", testHash)
	IncPK = sortKeyPair(user.From) // sorts the keyPair to compare with signature
	if RSA.Verify(signature, testHash, IncPK) {
		fmt.Println("yolo191919")
		return true
	} else {
		fmt.Println("yolo11234")
		return false
	}
}

// THIS IS A NEW FEATURE
func verifyBlock( b *Block ) bool {
	fmt.Println("this far222")
	var sig = b.Signature
	var sigInt = stringToBigInt(sig)
	for _, key := range b.Transactions {
		var s = strconv.Itoa(LocalBlockNumber) + key.Id
		testHash := RSA.Hash([]byte(s)) 
		if(RSA.Verify(sigInt, testHash, SequencerKeyPair )) {
			return true
		} 
	}
	return false
}

func convertBigIntToString(b *big.Int) string {
	bString := b.String()
	return bString
}

func convertPKtoString() {
	PkString = convertBigIntToString(LocalPK.N) + "," + convertBigIntToString(LocalPK.E)
}

func stringToBigInt(s string) *big.Int {
	n := new(big.Int)
	n, ok := n.SetString(s, 10)
	if !ok {
			fmt.Println("error1")
			return n
	}
	return n
}

func main() {
	// initiating local variables
	localMsg = makeMessage()

	ledger = makeLedger()
	
	// generate keys for transactions
	RSA.KeyGen(1024)
	
	// initiates Public Key
	LocalPK = RSA.GetKeyPair()

	// convert pk to string
	convertPKtoString()
	
	// addPK to localledger
	ledger.Accounts[PkString] = 1000

	ledger.Phase = 1 // sets the phase to stage one

	// stores the broadcasts and is true if used
	isBroadcasted = make(map[string]bool)

	// stores the transactions and is true if used
	transactionIsUsed = make(map[string]bool)

	BlocksRecieved = make(map[int]*Block)

	// starting reader on terminal
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Write IP and port")
	fmt.Print(">")

	// address which to dial
	peerAddr, _ := reader.ReadString('\n')
	peerAddr = strings.TrimSpace(peerAddr) // returns peerAdrr without trailing suffix

	// Opening a listener with a random port
	ln, _ := net.Listen("tcp", ":" )
	myIp = ln.Addr().String()
	localMsg.Connections = append(localMsg.Connections, myIp)
	connectedPeers = append(connectedPeers, myIp)
	fmt.Println("Listening on...", myIp)

	// accepter for new connections
	go listeningForConnections(ln)

	// dialing up the peeradress
	go dial(peerAddr)

	// foreloop that keeps the system going
	for {
		if(ledger.Phase == 2) {
			fmt.Println("-----------PHASE-2-PROCEEDED-----------")
			if(!isDesignatedSequencer) {
				requestTransaction()
				//makeFakeTransaction()
			}
			if(isDesignatedSequencer){
				updateBlock() // sets the phase to stage two
			}
		}
	}
}


// only used when testing 2 different peers sending a transaction at the same time
func makeFakeTransaction() {
	time.Sleep(5000 *time.Millisecond)
	for i := 1; i < 100; i++ {
	t1 := makeTransaction()
	t2 := makeTransaction()

	var localMap = make(map[int]string)
	var i = 0

	for key, _ := range ledger.Accounts {
		localMap[i]	= key
		i++
	}

	t1.From = PkString
	t1.To = localMap[0]
	t1.Amount = i
	t1.Id = myIp + t1.From + t1.To

	toSign1 := t1.From + t1.To + strconv.Itoa(t1.Amount)

	bigIntSign1 := RSA.Sign([]byte(toSign1))

	bigSignString1 := convertBigIntToString(bigIntSign1)

	t1.Signature = bigSignString1

	t2.From = PkString
	t2.To = localMap[1]
	t2.Amount = i+2
	t2.Id = myIp + t2.From + t2.To

	toSign2 := t2.From + t2.To + strconv.Itoa(t2.Amount)

	bigIntSign2 := RSA.Sign([]byte(toSign2))

	bigSignString2 := convertBigIntToString(bigIntSign2)

	t2.Signature = bigSignString2

	go handleTransaction(t1)
	go handleTransaction(t2)
	time.Sleep(1000 *time.Millisecond)
}
}