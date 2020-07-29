package errors

import "github.com/pkg/errors"

func WrappedErrNewLogger(err error) error {
	return errors.WithMessage(err, "new logger")
}

func WrappedErrDialNetlinkConnection(err error) error {
	return errors.WithMessage(err, "dial netlink connection")
}

func WrappedErrNewCommunicator(err error) error {
	return errors.WithMessage(err, "new communicator")
}

func WrappedErrCloseCommunicator(err error) error {
	return errors.WithMessage(err, "close communicator")
}

func WrappedErrEncodePayload(err error) error {
	return errors.WithMessage(err, "encode payload")
}

func WrappedErrSendMessage(err error) error {
	return errors.WithMessage(err, "send message")
}
