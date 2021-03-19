package reports

import (
	"encoding/json"
	"github.com/pkg/errors"
)

func MergeReports(reports ...Report) (map[string]interface{}, error) {
	merged := make(map[string]interface{}, 0)

	for _, report := range reports {
		reportDump, err := report.DumpReport()
		reportName := report.ReportName()

		if err != nil {
			return nil, errors.WithMessagef(err, "dump report '%s'", reportName)
		}

		if err := json.Unmarshal(reportDump, &merged); err != nil {
			return nil, errors.WithMessagef(err, "merge with report '%s'", reportName)
		}
	}

	return merged, nil
}
