package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

const fahAPI = "https://stats.foldingathome.org/api"

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
		if strings.Contains(t, "PyON") {
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

// ReadFAH sends command to FAH client and unmarshals data into struct
func ReadFAH(conn net.Conn, cmd string, target interface{}) error {
	out, err := readPyON(conn, cmd)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(out), target)
}

// ReadAPI sends GET request to FAH API and unmarshals data into struct
func ReadAPI(endpoint string, target interface{}) error {
	resp, err := myClient.Get(fmt.Sprintf("%s/%s", fahAPI, endpoint))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
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
	ID          string          `json:"id"`
	Status      string          `json:"status"`
	Description string          `json:"description"`
	Options     SlotInfoOptions `json:"options"`
	Reason      string          `json:"reason"`
	Idle        bool            `json:"idle"`
}

// SlotInfoOptions from FAH response
type SlotInfoOptions struct {
	Paused bool `json:"paused"`
}

// Options output from options command
type Options struct {
	Power string `json:"power"`
	Team  string `json:"team"`
	User  string `json:"user"`
}

// DonorAPI from https://stats.foldingathome.org/api/donor/<user>
type DonorAPI struct {
	Rank   int       `json:"rank"`
	ID     int       `json:"id"`
	Name   string    `json:"name"`
	Teams  []TeamAPI `json:"teams"`
	Credit int       `json:"credit"`
}

// TeamAPI from https://stats.foldingathome.org/api/team/<team>
type TeamAPI struct {
	Credit int    `json:"credit"`
	Team   int    `json:"team"`
	Name   string `json:"name"`
}
