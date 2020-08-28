package operators

//
//import (
//	"context"
//	"fmt"
//	"github.com/memlab/agent/internal/reports"
//	"github.com/memlab/agent/internal/reports/postdetection"
//	"github.com/memlab/agent/internal/types"
//	"github.com/pkg/errors"
//	"os/exec"
//)
//
//type ProcDumpOperator struct{}
//
//func (p *ProcDumpOperator) OperatorName() string {
//	return "proc-dump-operator"
//}
//
//func (p *ProcDumpOperator) Operate(ctx context.Context, pid types.Pid) (reports.Report, error) {
//	// todo: refactor to operate as following:
//	// 		1. take proc dump
//	// 		2. start a goroutine which dumps it in a background pool which uploads to backend/other destination
//	//		and keeps internal state persistently in case it didn't finish.
//
//	cmd := exec.Command("sudo", "procdump", "-p", fmt.Sprintf("%d", pid))
//	output, err := cmd.Output()
//	if err != nil {
//		return nil, errors.WithMessage(err, "run procdump command")
//	}
//
//	output
//	return postdetection.NewProcDumpReport(ps)
//}
//
//func (p *ProcDumpOperator) FailPipelineOnError() bool {
//	return false
//}
