package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

// readPyON Send command to connection and read response
// PyON is converted to JSON using text replacement
func readPyON(conn net.Conn, command string) (outJSON string, err error) {

	p := []byte(command + "\r\n")
	n, err := conn.Write(p)
	if err != nil {
		return
	}
	if expected, actual := len(p), n; expected != actual {
		err = fmt.Errorf("transmission problem: tried sending %d bytes, but actually only sent %d bytes", expected, actual)
		return
	}
	var rawOut string
	var reading bool
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		t := scanner.Text()
		if t == "---" {
			break
		}
		if strings.HasPrefix(t, "PyON") {
			reading = true
			continue
		}
		if reading {
			rawOut += t
		}
	}
	outJSON = strings.ReplaceAll(rawOut, ": None", ": null")
	outJSON = strings.ReplaceAll(outJSON, ": True", ": true")
	outJSON = strings.ReplaceAll(outJSON, ": False", ": false")
	return
}

// ReadQueueInfo sends command to FAH client and returns a list of QueueInfo structs
func ReadQueueInfo(conn net.Conn) (q []QueueInfo, err error) {
	out, err := readPyON(conn, "queue-info")
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(out), &q)
	if err != nil {
		return
	}
	return
}

// ReadSlotInfo sends command to FAH client and returns a list of SlotInfo structs
func ReadSlotInfo(conn net.Conn) (q []SlotInfo, err error) {
	out, err := readPyON(conn, "slot-info")
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(out), &q)
	if err != nil {
		return
	}
	return
}

// QueueInfo is the data from the queue-info command
type QueueInfo struct {
	ID             string `json:"id"`
	State          string `json:"state"`
	Error          string `json:"error"`
	Project        int    `json:"project"`
	Run            int    `json:"run"`
	Clone          int    `json:"clone"`
	Gen            int    `json:"gen"`
	Core           string `json:"core"`
	Unit           string `json:"unit"`
	PercentDone    string `json:"percentdone"`
	Eta            string `json:"eta"`
	Ppd            string `json:"ppd"`
	CreditEstimate string `json:"creditestimate"`
	WaitingOn      string `json:"waitingon"`
	NextAttempt    string `json:"nextattempt"`
	TimeRemaining  string `json:"timeremaining"`
	TotalFrames    int    `json:"totalframes"`
	FramesDone     int    `json:"framesdone"`
	Assigned       string `json:"assigned"`
	Timeout        string `json:"timeout"`
	Deadline       string `json:"deadline"`
	Ws             string `json:"ws"`
	Cs             string `json:"cs"`
	Attempts       int    `json:"attempts"`
	Slot           string `json:"slot"`
	Tpf            string `json:"tpf"`
	BaseCredit     string `json:"basecredit"`
}

// SlotInfo output from slot-info command
type SlotInfo struct {
	ID          string  `json:"id"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
	Options     Options `json:"options"`
	Reason      string  `json:"reason"`
	Idle        bool    `json:"idle"`
}

// Options from FAH response
type Options struct {
	Paused bool `json:"paused"`
}
