package operators

import "context"

type Operator interface {
	Name() string
	Operate(ctx context.Context) error
	StopOnFailure() bool
}
