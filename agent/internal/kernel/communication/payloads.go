package communication

import (
	"github.com/mdlayher/netlink"
	"github.com/pkg/errors"
)

// Netlink commands enum (see kernel/communication.h)
const (
	CommandMonitorProcess = iota
	//_                     // CommandNotifyCaughtSignal
	CommandHandledCaughtSignal
)

const (
	//CommandMonitorProcess = iota
	CommandNotifyCaughtSignal = iota
	//CommandHandledCaughtSignal
)

// Netlink attributes enum (see kernel/communication.h)
const (
	AttributePid uint16 = iota + 1 // Starts from 1
	AttributeDoWatch
	AttributeSignalNotificationSignal
)

const (
	ActionUnwatchProcess uint8 = 0
	ActionWatchProcess   uint8 = 1
)

type PayloadMonitorProcess struct {
	Pid   uint32
	Watch uint8
}

func (p *PayloadMonitorProcess) Encode() ([]byte, error) {
	encoder := netlink.NewAttributeEncoder()
	encoder.Uint32(AttributePid, p.Pid)
	encoder.Uint8(AttributeDoWatch, p.Watch)
	return encoder.Encode()
}

type PayloadCaughtSignal struct {
	Pid    uint32
	Signal uint32
}

func (p *PayloadCaughtSignal) Encode() ([]byte, error) {
	encoder := netlink.NewAttributeEncoder()
	encoder.Uint32(AttributePid, p.Pid)
	encoder.Uint32(AttributeSignalNotificationSignal, p.Signal)
	return encoder.Encode()
}

func DecodePayloadCaughtSignal(data []byte) (*PayloadCaughtSignal, error) {
	decoder, err := netlink.NewAttributeDecoder(data)
	if err != nil {
		return nil, err
	}

	payload := &PayloadCaughtSignal{}
	for decoder.Next() {
		switch decoder.Type() {
		case AttributePid:
			payload.Pid = decoder.Uint32()
		case AttributeSignalNotificationSignal:
			payload.Signal = decoder.Uint32()
		default:
			return nil, errors.Errorf("invalid attribute type ('%d')", decoder.Type())
		}
	}

	if err := decoder.Err(); err != nil {
		return nil, errors.WithMessage(err, "malformed attributes")
	}
	return payload, nil
}
