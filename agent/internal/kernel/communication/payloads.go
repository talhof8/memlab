package communication

import (
	"bytes"
	"encoding/binary"
)

type MonitorProcessAction uint32

const (
	ActionWatchProcess MonitorProcessAction = iota + 1
	ActionUnwatchProcess
)

// A container around the Payload interface, so that binary.Read() can be used.
type PayloadContainer struct {
	Payload Payload
}

type Payload interface {
	Type() string
}

type PayloadMonitorProcess struct {
	Pid    uint32
	Action MonitorProcessAction
}

func (p *PayloadMonitorProcess) Type() string {
	return "monitor_process"
}

type PayloadCaughtSignal struct {
	Pid            uint32
	ExecutablePath []byte
	Signal         uint32
}

func (p *PayloadCaughtSignal) Type() string {
	return "caught_signal"
}

func encodePayload(payload Payload) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, &PayloadContainer{payload})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decodePayload(data []byte) (Payload, error) {
	var payloadContainer PayloadContainer
	buf := bytes.NewBuffer(data)

	err := binary.Read(buf, binary.LittleEndian, payloadContainer)
	if err != nil {
		return nil, err
	}

	return payloadContainer.Payload, nil
}
