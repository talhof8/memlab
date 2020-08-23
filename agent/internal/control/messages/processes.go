package messages

type ProcessListReport struct {
	Host string `json:"host"`
	Pid uint32 `json:"pid"`
	Executable string `json:"executable"`
	CommandLine string `json:"command_line"`
	IsActive bool `json:"is_active"`

	/**
	    command_line = models.CharField(max_length=1000, null=False, blank=False)
	    is_active = models.BooleanField(default=True)
	    monitored = models.BooleanField(default=False)
	    seen_at = models.DateTimeField(default=timezone.now, null=False, blank=False)
	    monitored_since = models.DateTimeField(null=True, blank=True)
	    disappeared_at = models.DateTimeField(null=True, blank=True)
	 */
}

func NewProcessListReport() (*ProcessListReport, error) {
	return nil, nil
}