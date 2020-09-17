package detection

import "github.com/pkg/errors"

func errDetectorAlreadyExists(detectorName string) error {
	return errors.Errorf("detector '%s' already exists", detectorName)
}

func errDetectorDoesNotExist(detectorName string) error {
	return errors.Errorf("detector '%s' does not exist", detectorName)
}
