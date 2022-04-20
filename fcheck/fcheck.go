/*

This package specifies the API to the failure checking library to be
used in assignment 2 of UBC CS 416 2021W2.

You are *not* allowed to change the API below. For example, you can
modify this file by adding an implementation to Stop, but you cannot
change its API.

*/

package fchecker

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
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
	StartStruct StartStruct
	AckConn     net.Conn
	MonitorConn net.Conn
	RTT         time.Duration
	monitorErr  chan error
	responseErr chan error
	done        bool
}

const MAX_ATTEMPTS_ERROR = "max consecutive heartbeats lost"
const CLOSE_ERROR = "connection closed"

func (fcheck *Fcheck) RespondToHeartbeat() error {
	addr, err := net.ResolveUDPAddr("udp", fcheck.StartStruct.AckLocalIPAckLocalPort)
	CheckErr(err, "Error resolving UDP address: %v\n", err)
	conn, err := net.ListenUDP("udp", addr)
	CheckErr(err, "Error listening on UDP address: %v\n", err)
	fcheck.AckConn = conn
	for {
		heartBeatBuf := make([]byte, 1024)
		len, addr, err := conn.ReadFromUDP(heartBeatBuf)
		if fcheck.done {
			return nil
		}
		CheckErr(err, "Error reading from UDP: %v\n", err)
		heartBeat := HBeatMessage{}
		err = gob.NewDecoder(bytes.NewBuffer(heartBeatBuf[0:len])).Decode(&heartBeat)
		if err != nil {
			fcheck.responseErr <- err
			return err
		}
		ack := AckMessage{
			HBEatEpochNonce: heartBeat.EpochNonce,
			HBEatSeqNum:     heartBeat.SeqNum,
		}
		var ackBuf bytes.Buffer
		gob.NewEncoder(&ackBuf).Encode(&ack)
		//TODO remove
		time.Sleep(time.Millisecond * 500)
		_, err = conn.WriteToUDP(ackBuf.Bytes(), addr)
		if err != nil {
			fcheck.responseErr <- err
			return err
		}
	}
}

func (fcheck *Fcheck) MonitorHeartbeats() error {
	laddr, monitorErr := net.ResolveUDPAddr("udp", fcheck.StartStruct.HBeatLocalIPHBeatLocalPort)
	CheckErr(monitorErr, "Error converting UDP address: %v\n", monitorErr)
	raddr, monitorErr := net.ResolveUDPAddr("udp", fcheck.StartStruct.HBeatRemoteIPHBeatRemotePort)
	CheckErr(monitorErr, "Error converting UDP address: %v\n", monitorErr)
	// setup UDP connection
	conn, err := net.DialUDP("udp", laddr, raddr)

	CheckErr(err, "Couldn't connect to the server", err)
	fcheck.MonitorConn = conn
	heartbeat := HBeatMessage{
		EpochNonce: fcheck.StartStruct.EpochNonce,
		SeqNum:     0,
	}
	seqTimestamps := make(map[uint64]time.Time)
	numberOfLostMsg := 0
	for {
		heartbeat.SeqNum = (heartbeat.SeqNum + 1) % math.MaxUint64
		var heartbeatBuf bytes.Buffer
		gob.NewEncoder(&heartbeatBuf).Encode(&heartbeat)
		seqTimestamps[heartbeat.SeqNum] = time.Now()
		_, err := fcheck.MonitorConn.Write(heartbeatBuf.Bytes())
		if err != nil {
			return err
		}

		//Prevent infinite loop
		fcheck.MonitorConn.SetReadDeadline(time.Now().Add(fcheck.RTT))
		for i := 0; i < 100; i++ {
			ackBuf := make([]byte, 1024)
			len, err := fcheck.MonitorConn.Read(ackBuf)
			if err != nil {
				if fcheck.done {
					err := errors.New(CLOSE_ERROR)
					fcheck.monitorErr <- err
					return err
				}
				if os.IsTimeout(err) {

					if _, found := seqTimestamps[heartbeat.SeqNum]; found {
						numberOfLostMsg += 1
					}
					if numberOfLostMsg >= int(fcheck.StartStruct.LostMsgThresh) {
						err := errors.New(MAX_ATTEMPTS_ERROR)
						fcheck.monitorErr <- err
						return err
					}
					break
				} else {

					fcheck.monitorErr <- err
					return err
				}
			}
			ack := AckMessage{}
			err = gob.NewDecoder(bytes.NewBuffer(ackBuf[0:len])).Decode(&ack)
			if err != nil {
				fcheck.monitorErr <- err
				return err
			}
			if reqTime, found := seqTimestamps[ack.HBEatSeqNum]; ack.HBEatEpochNonce == heartbeat.EpochNonce && found {
				fcheck.RTT = (fcheck.RTT + time.Now().Sub(reqTime)) / 2
				numberOfLostMsg = 0
				delete(seqTimestamps, ack.HBEatSeqNum)
				continue
			}
		}
	}
}

// Starts the Fcheck library.

func (fcheck *Fcheck) Start(arg StartStruct) (notifyCh <-chan FailureDetected, err error) {
	fcheck.done = false

	fcheck.StartStruct = arg
	fcheck.RTT = 3000 * time.Millisecond
	channel := make(chan FailureDetected)

	if arg.HBeatLocalIPHBeatLocalPort == "" {
		// ONLY arg.AckLocalIPAckLocalPort is set
		//
		// Start Fcheck without monitoring any node, but responding to heartbeats.
		fcheck.responseErr = make(chan error)

		go fcheck.RespondToHeartbeat()
		go func() {
			err = <-fcheck.responseErr
			defer fcheck.AckConn.Close()
			defer close(fcheck.responseErr)
		}()

		return nil, err
	} else {
		go fcheck.MonitorHeartbeats()

		// Else: ALL fields in arg are set
		// Start the Fcheck library by monito
		//ring a single node and
		// also responding to heartbeats.
		go func() {
			fcheck.monitorErr = make(chan error)
			err = <-fcheck.monitorErr
			defer close(fcheck.monitorErr)
			defer fcheck.MonitorConn.Close()
			if err.Error() == MAX_ATTEMPTS_ERROR {
				failure := FailureDetected{
					UDPIpPort: fcheck.StartStruct.HBeatRemoteIPHBeatRemotePort,
					Timestamp: time.Now(),
				}
				go func() {
					channel <- failure
				}()
			} else {
				//	Undefined behavior
				failure := FailureDetected{
					UDPIpPort: fcheck.StartStruct.HBeatRemoteIPHBeatRemotePort,
					Timestamp: time.Now(),
				}
				go func() {
					channel <- failure
				}()
			}
		}()
	}

	return channel, err

}

// Tells the library to stop monitoring/responding acks.
func (fcheck *Fcheck) Stop() {
	fcheck.done = true
	if fcheck.AckConn != nil {
		fcheck.AckConn.Close()
	}
	if fcheck.MonitorConn != nil {
		fcheck.MonitorConn.Close()
	}
	return
}

func (fcheck *Fcheck) StopRespondToHeartbeat() {
	fcheck.done = true
}

func CheckErr(err error, errfmsg string, fargs ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, errfmsg, fargs...)
		os.Exit(1)
	}
}
