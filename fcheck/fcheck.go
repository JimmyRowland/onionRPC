package fchecker

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

////////////////////////////////////////////////////// DATA
// Define the message types fchecker has to use to communicate to other
// fchecker instances. We use Go's type declarations for this:
// https://golang.org/ref/spec#Type_declarations

// Heartbeat message.
type HBeatMessage struct {
	EpochNonce uint64 // Identifies this fchecker instance/epoch.
	SeqNum     uint64 // Unique for each heartbeat in an epoch.
}

// An ack message; response to a heartbeat.
type AckMessage struct {
	HBEatEpochNonce uint64 // Copy of what was received in the heartbeat.
	HBEatSeqNum     uint64 // Copy of what was received in the heartbeat.
}

// Notification of a failure, signal back to the client using this
// library.
type FailureDetected struct {
	UDPIpPort string    // The RemoteIP:RemotePort of the failed node.
	Timestamp time.Time // The time when the failure was detected.
}

////////////////////////////////////////////////////// API

type StartStruct struct {
	AckLocalIPAckLocalPort       string
	EpochNonce                   uint64
	HBeatLocalIPHBeatLocalPort   string
	HBeatRemoteIPHBeatRemotePort string
	LostMsgThresh                uint8
}

type Fcheck struct {
	allActive bool
	active    map[string]bool
	mu        sync.Mutex
	ch        chan FailureDetected
}

func NewFcheck() Fcheck {
	return Fcheck{
		active: make(map[string]bool),
	}
}

// Starts the fcheck library.
func (fc *Fcheck) Start(arg StartStruct) (notifyCh <-chan FailureDetected, err error) {
	fc.ch = make(chan FailureDetected)
	if arg.HBeatLocalIPHBeatLocalPort == "" {
		// ONLY arg.AckLocalIPAckLocalPort is set
		//
		// Start fcheck without monitoring any node, but responding to heartbeats.
		go fc.beginListen(arg.AckLocalIPAckLocalPort)

		return nil, nil
	}
	// Else: ALL fields in arg are set
	// Start the fcheck library by monitoring a single node and
	// also responding to heartbeats.

	go fc.beginListen(arg.AckLocalIPAckLocalPort)
	go fc.beginHeartbeat(arg)

	return fc.ch, nil
}

func (fc *Fcheck) BeginMonitoring(laddr, raddr string, lostMsgThresh int) error {
	start := StartStruct{
		HBeatLocalIPHBeatLocalPort:   laddr,
		HBeatRemoteIPHBeatRemotePort: raddr,
		EpochNonce:                   rand.Uint64(),
		LostMsgThresh:                uint8(lostMsgThresh),
	}
	fc.active[raddr] = true
	go fc.beginHeartbeat(start)
	return nil
}

func (fc *Fcheck) StopMonitoring(raddr string) {
	fc.mu.Lock()
	fc.active[raddr] = false
	fc.mu.Unlock()
}

func (fc *Fcheck) beginHeartbeat(arg StartStruct) {
	laddr, err := net.ResolveUDPAddr("udp", arg.HBeatLocalIPHBeatLocalPort)
	checkErr(err, "Error converting local UDP address: %v\n", err)

	raddr, err := net.ResolveUDPAddr("udp", arg.HBeatRemoteIPHBeatRemotePort)
	checkErr(err, "Error converting remote UDP address: %v\n", err)

	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		fc.ch <- FailureDetected{arg.HBeatRemoteIPHBeatRemotePort, time.Now()}
		return
	}
	defer conn.Close()

	lostMsgs := 0
	var seqNum uint64
	recvBuf := make([]byte, 1024)
	rtt := time.Duration(3) * time.Second
	for {
		fc.mu.Lock()
		doHeartBeat := fc.active[arg.HBeatRemoteIPHBeatRemotePort]
		fc.mu.Unlock()
		if !doHeartBeat {
			break
		}
		if lostMsgs >= int(arg.LostMsgThresh) {
			fc.ch <- FailureDetected{arg.HBeatRemoteIPHBeatRemotePort, time.Now()}
			break
		}

		hb := HBeatMessage{arg.EpochNonce, seqNum}
		conn.Write(encodeHB(&hb))
		t_start := time.Now()
		wait := rtt
		// Monitor connection for Ack, handle out-of-sequence Acks as per spec
		for {
			conn.SetReadDeadline(time.Now().Add(wait))
			len, err := conn.Read(recvBuf)
			if err != nil {
				// Read timeout
				fmt.Println("missing hb ack")
				lostMsgs += 1
				break
			}
			ack, err := decodeAck(recvBuf, len)
			if err != nil {
				// Malformed Ack
				checkErr(err, "Failed to decode fcheck Ack", err)
				break
			}
			if ack.HBEatSeqNum != hb.SeqNum && hb.SeqNum < seqNum {
				// Out-of-sequence Ack
				lostMsgs = 0
				continue
			}

			// Happy path:
			rtt = (rtt + time.Since(t_start)) / 2
			lostMsgs = 0

			// Set rtt to a *maximum* of 20ms, to avoid egregious overhead
			// when latency is low
			rttNs := int64(rtt)
			minRttNs := 20000000 // 20 milliseconds
			rtt = time.Duration(math.Max(float64(rttNs), float64(minRttNs))) * time.Nanosecond
			break
		}
		seqNum += 1
	}
}

func (fc *Fcheck) beginListen(addr string) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	checkErr(err, "Error converting UDP address: %v\n", err)

	conn, err := net.ListenUDP("udp", laddr)
	checkErr(err, "Couldn't start UDP-listen", err)
	defer conn.Close()
	recvBuf := make([]byte, 1024)

	fc.mu.Lock()
	fc.allActive = true
	doListen := true
	fc.mu.Unlock()

	fmt.Println("====> LISTENING AT ", conn.LocalAddr().String())

	for doListen {
		len, caddr, err := conn.ReadFrom(recvBuf)
		checkErr(err, "Error reading on UDP connection: %v\n", err)

		hb, err := decodeHB(recvBuf, len)
		checkErr(err, "Error decoding heartbeat: %v\n", err)

		ack := AckMessage{hb.EpochNonce, hb.SeqNum}
		go conn.WriteTo(encodeAck(&ack), caddr)

		fc.mu.Lock()
		doListen = fc.allActive
		fc.mu.Unlock()
	}
}

// Tells the library to stop monitoring/responding acks.
func (fc *Fcheck) Stop() {
	fc.mu.Lock()
	fc.allActive = false
	for k, _ := range fc.active {
		fc.active[k] = false
	}
	fc.mu.Unlock()
}

func checkErr(err error, errfmsg string, fargs ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, errfmsg, fargs...)
		os.Exit(1)
	}
}

func encodeHB(hb *HBeatMessage) []byte {
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(hb)
	return buf.Bytes()
}

func decodeAck(buf []byte, len int) (AckMessage, error) {
	var decoded AckMessage
	err := gob.NewDecoder(bytes.NewBuffer(buf[0:len])).Decode(&decoded)
	if err != nil {
		return AckMessage{}, err
	}
	return decoded, nil
}

func encodeAck(ack *AckMessage) []byte {
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(ack)
	return buf.Bytes()
}

func decodeHB(buf []byte, len int) (HBeatMessage, error) {
	var decoded HBeatMessage
	err := gob.NewDecoder(bytes.NewBuffer(buf[0:len])).Decode(&decoded)
	if err != nil {
		return HBeatMessage{}, err
	}
	return decoded, nil
}
