package host

import (
	"github.com/denisbrodbeck/machineid"
	"github.com/pkg/errors"
)

func MachineId() (string, error) {
	machineId, err := machineid.ID()
	if err != nil { // todo: find a fallback on error (should be a constant identifier)
		return "", errors.WithMessage(err, "get machine id")
	}
	return machineId, nil
}
