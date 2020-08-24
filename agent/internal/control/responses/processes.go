package responses

import "time"

type Process struct {
	RecordId       string     `json:"id"`
	MachineId      string     `json:"machine_id"`
	Pid            int32      `json:"pid"`
	Executable     string     `json:"executable"`
	CommandLine    string     `json:"command_line"`
	CreateTime     *time.Time `json:"create_time"`
	LastSeenAt     *time.Time `json:"last_seen_at"`
	Monitored      string     `json:"monitored"`
	MonitoredSince *time.Time `json:"monitored_since"`
}
