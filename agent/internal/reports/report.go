package reports

type Report interface {
	ReportName() string
	DumpReport() ([]byte, error)
}
