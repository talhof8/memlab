package communication

import (
	stdLibErrors "errors"
	"github.com/mdlayher/genetlink"
	"github.com/mdlayher/netlink"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"sync"
	"syscall"
)

// todo: find a better way of communicating than creating two separate generic-netlink families

type Communicator struct {
	logger            *zap.Logger
	waitGroup         sync.WaitGroup
	sendConn          *genetlink.Conn
	recvConn          *genetlink.Conn
	sendConnFamily    *genetlink.Family
	recvConnFamily    *genetlink.Family
	caughtSignalsChan chan *PayloadCaughtSignal
}

func connectToGenericNetlink(familyName string) (*genetlink.Conn, *genetlink.Family, error) {
	conn, err := genetlink.Dial(&netlink.Config{
		DisableNSLockThread: true,
	})
	if err != nil {
		return nil, nil, errors.WithMessage(err, "dial netlink connection")
	}

	family, err := conn.GetFamily(familyName)
	if err != nil {
		if stdLibErrors.Is(err, os.ErrNotExist) {
			return nil, nil, errors.Errorf("family '%s' does not exist", familyName)
		}
		return nil, nil, errors.WithMessagef(err, "get family '%s'", familyName)
	}
	return conn, &family, nil
}

func NewCommunicator(rootLogger *zap.Logger, recvFamilyName, sendFamilyName string) (*Communicator, error) {
	recvConn, recvConnFamily, err := connectToGenericNetlink(recvFamilyName)
	if err != nil {
		return nil, err
	}

	sendConn, sendConnFamily, err := connectToGenericNetlink(sendFamilyName)
	if err != nil {
		return nil, err
	}

	logger := rootLogger.Named("kernel-communicator")

	return &Communicator{
		logger:            logger,
		sendConn:          sendConn,
		sendConnFamily:    sendConnFamily,
		recvConn:          recvConn,
		recvConnFamily:    recvConnFamily,
		caughtSignalsChan: make(chan *PayloadCaughtSignal, 0),
	}, nil
}

func (c *Communicator) WatchProcess(pid uint32) error {
	c.logger.Debug("Watch process", zap.Uint32("Pid", pid))
	payload := &PayloadMonitorProcess{
		Pid:   pid,
		Watch: ActionWatchProcess,
	}
	return c.sendMonitorProcessMessage(CommandMonitorProcess, payload)
}

func (c *Communicator) UnwatchProcess(pid uint32) error {
	c.logger.Debug("Un-watch process", zap.Uint32("Pid", pid))
	payload := &PayloadMonitorProcess{
		Pid:   pid,
		Watch: ActionUnwatchProcess,
	}
	return c.sendMonitorProcessMessage(CommandMonitorProcess, payload)
}

func (c *Communicator) NotifyHandledSignal(pid uint32) error {
	c.logger.Debug("Notify kernel that signal was handled", zap.Uint32("Pid", pid))
	payload := &PayloadMonitorProcess{
		Pid: pid,
	}
	return c.sendMonitorProcessMessage(CommandHandledCaughtSignal, payload)
}

func (c *Communicator) sendMonitorProcessMessage(command uint8, payload *PayloadMonitorProcess) error {
	data, err := payload.Encode()
	if err != nil {
		return errors.WithMessage(err, "encode payload")
	}

	message := genetlink.Message{
		Header: genetlink.Header{
			Command: command,
		},
		Data: data,
	}

	_, err = c.sendConn.Send(message, c.sendConnFamily.ID, netlink.Request)
	if err != nil {
		return errors.WithMessage(err, "send message")
	}
	return nil
}

func (c *Communicator) ListenForCaughtSignals() error {
	c.waitGroup.Add(1)
	go func() {
		defer c.waitGroup.Done()

		c.logger.Debug("Join family groups")
		if !c.joinFamilyGroups() {
			return
		}

		c.logger.Debug("Listen for netlink messages")
		defer c.logger.Debug("Done listen for netlink messages")
		defer close(c.caughtSignalsChan)

		for {
			messages, _, err := c.recvConn.Receive()
			if err != nil {
				// Since there's not way to gracefully close the connection (e.g, via a context cancellation),
				// a call to close() is used to close it. A syscall.EBADF error is most likely caused by that close,
				// hence we assume it's ok and do not write any error log.
				if err == syscall.EBADF {
					return
				} // todo: make it work

				c.logger.Error("Failed to receive messages", zap.Error(err))
				continue
			}

			c.logger.Debug("Received messages", zap.Int("Count", len(messages)))
			c.handleMessages(messages)
		}
	}()

	return nil
}

func (c *Communicator) joinFamilyGroups() bool {
	if len(c.recvConnFamily.Groups) == 0 {
		c.logger.Debug("There are 0 groups for family", zap.String("FamilyName", c.sendConnFamily.Name))
		return true
	}

	for _, group := range c.recvConnFamily.Groups {
		c.logger.Debug("Joining family group", zap.String("GroupName", group.Name), zap.Uint32("GroupId",
			group.ID))

		err := c.recvConn.JoinGroup(group.ID)
		if err != nil {
			c.logger.Error("Failed to join group", zap.Error(err), zap.Uint32("GroupID", group.ID))
			return false
		}
	}

	return true
}

func (c *Communicator) handleMessages(messages []genetlink.Message) {
	for _, message := range messages {
		if message.Data == nil {
			c.logger.Debug("Message data is empty, continuing...")
			continue
		}

		caughtSignalPayload, err := DecodePayloadCaughtSignal(message.Data)
		if err != nil {
			c.logger.Error("Failed to decode 'caught-signal' payload", zap.Int("PayloadLen", len(message.Data)),
				zap.Error(err))
			continue
		}

		c.caughtSignalsChan <- caughtSignalPayload
	}
}

func (c *Communicator) CaughtSignalsChan() <-chan *PayloadCaughtSignal {
	return c.caughtSignalsChan
}

// todo: think how to restore/clean state when kernel module keeps running and agent stops and vice-versa
func (c *Communicator) Close() error {
	if err := c.sendConn.Close(); err != nil {
		return errors.WithMessage(err, "close netlink connection")
	}

	if err := c.recvConn.Close(); err != nil {
		return errors.WithMessage(err, "close netlink connection")
	}

	close(c.caughtSignalsChan)

	c.waitGroup.Wait()
	return nil
}
