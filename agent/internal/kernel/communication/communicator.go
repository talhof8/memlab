package communication

import (
	"github.com/mdlayher/netlink"
	"github.com/memlab/agent/internal/errors"
	"github.com/memlab/agent/internal/logging"
	"go.uber.org/zap"
	"sync"
	"syscall"
)

type nlGroup int

const (
	nlGroupMonitorProcess nlGroup = iota + 1
	nlGroupSignals
)

const (
	nlGroupsAll = nlGroupMonitorProcess | nlGroupSignals
)

type Communicator struct {
	logger        *zap.Logger
	waitGroup     sync.WaitGroup
	conn          *netlink.Conn
	caughtSignals chan *PayloadCaughtSignal
}

func NewCommunicator(nlFamily int) (*Communicator, error) {
	conn, err := netlink.Dial(nlFamily, &netlink.Config{
		Groups:              uint32(nlGroupsAll),
		DisableNSLockThread: true,
	})
	if err != nil {
		return nil, errors.WrappedErrDialNetlinkConnection(err)
	}

	logger, err := logging.NewLogger("memlab-kernel-communicator")
	if err != nil {
		return nil, errors.WrappedErrNewLogger(err)
	}

	return &Communicator{
		logger:        logger,
		conn:          conn,
		caughtSignals: make(chan *PayloadCaughtSignal, 0),
	}, nil
}

func (c *Communicator) WatchProcess(pid uint32) error {
	payload := &PayloadMonitorProcess{
		Pid:    pid,
		Action: ActionWatchProcess,
	}
	_, err := c.sendMessage(payload)
	return err
}

func (c *Communicator) UnwatchProcess(pid uint32) error {
	payload := &PayloadMonitorProcess{
		Pid:    pid,
		Action: ActionUnwatchProcess,
	}
	_, err := c.sendMessage(payload)
	return err
}

func (c *Communicator) ListenForSignals() error {
	c.waitGroup.Add(1)

	go func() {
		defer c.waitGroup.Done()

		for {
			messages, err := c.conn.Receive()
			if err != nil {
				if err == syscall.EBADF { // Most likely caused by c.close() - simply stop execution
					return
				}

				c.logger.Error("Failed to receive messages", zap.Error(err))
				continue
			}

			c.handleMessages(messages)
		}
	}()

	return nil
}

func (c *Communicator) Signals() <-chan *PayloadCaughtSignal {
	return c.caughtSignals
}

// todo: think how to restore/clean state when kernel module keeps running and agent stops and vice-versa
func (c *Communicator) Close() error {
	if err := c.conn.Close(); err != nil {
		return err
	}

	c.waitGroup.Wait()
	return nil
}

func (c *Communicator) sendMessage(payload *PayloadMonitorProcess) (*netlink.Message, error) {
	data, err := encodePayload(payload)
	if err != nil {
		return nil, errors.WrappedErrEncodePayload(err)
	}

	message := netlink.Message{
		Header: netlink.Header{
			// Package netlink will automatically set header fields
			// which are set to zero
			Flags: netlink.Request,
		},
		Data: data,
	}

	reply, err := c.conn.Send(message)
	if err != nil {
		return nil, errors.WrappedErrSendMessage(err)
	}
	return &reply, nil
}

func (c *Communicator) handleMessages(messages []netlink.Message) {
	for _, message := range messages {
		if message.Data == nil {
			continue
		}

		payload, err := decodePayload(message.Data)
		if err != nil {
			c.logger.Error("Failed to decode payload", zap.Int("PayloadLen", len(message.Data)),
				zap.Error(err))
			continue
		}

		caughtSignalPayload, ok := payload.(*PayloadCaughtSignal)
		if !ok {
			continue
		}

		c.caughtSignals <- caughtSignalPayload
	}
}
