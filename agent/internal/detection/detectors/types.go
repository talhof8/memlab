package detectors

type DetectorType int

const (
	DetectorTypeSignals DetectorType = iota
)

var detectorNames = map[DetectorType]string{
	DetectorTypeSignals: "signal-detector",
}

func (dt DetectorType) Name() string {
	name, found := detectorNames[dt]
	if !found {
		return ""
	}
	return name
}
