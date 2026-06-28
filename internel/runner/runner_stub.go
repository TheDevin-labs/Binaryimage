
package runner

import (
	"errors"

	"github.com/TheDevin-Labs/binaryimage/internal/signal"
	"github.com/TheDevin-Labs/binaryimage/pkg/sbimg"
)

type Result struct {
	Exited bool
	Fault  *signal.FaultInfo
}

func Execute(img *sbimg.Image) (*Result, error) {
	return nil, errors.New("runner: direct execution is only supported on linux/amd64 and linux/arm64")
}
