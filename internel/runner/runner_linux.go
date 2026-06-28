
package runner

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/TheDevin-Labs/binaryimage/internal/signal"
	"github.com/TheDevin-Labs/binaryimage/pkg/sbimg"
)

type Result struct {
	Exited bool
	Fault  *signal.FaultInfo
}

func Execute(img *sbimg.Image) (*Result, error) {
	size := len(img.Code)
	if size == 0 {
		return nil, fmt.Errorf("runner: image contains no executable code")
	}

	mem, err := syscall.Mmap(
		-1, 0, size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_PRIVATE|syscall.MAP_ANONYMOUS,
	)
	if err != nil {
		return nil, fmt.Errorf("runner: mmap failed: %w", err)
	}

	copy(mem, img.Code)

	if err := syscall.Mprotect(mem, syscall.PROT_READ|syscall.PROT_EXEC); err != nil {
		_ = syscall.Munmap(mem)
		return nil, fmt.Errorf("runner: mprotect failed: %w", err)
	}

	result := &Result{}
	faultCh := make(chan *signal.FaultInfo, 1)
	doneCh := make(chan struct{})

	signal.Arm(func(f *signal.FaultInfo) {
		faultCh <- f
		<-doneCh
	})

	execCh := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				execCh <- fmt.Errorf("runner: panic during execution: %v", r)
			}
		}()

		fn := *(*func())(unsafe.Pointer(&mem))
		fn()
		execCh <- nil
	}()

	select {
	case fault := <-faultCh:
		result.Fault = fault
		close(doneCh)
		fmt.Fprintf(os.Stderr, "\nrunner: hardware exception intercepted\n")
		fmt.Fprintf(os.Stderr, "  signal  : %s\n", fault.Signal)
		fmt.Fprintf(os.Stderr, "  code    : %d\n", fault.Code)
		fmt.Fprintf(os.Stderr, "  address : 0x%x\n", fault.Address)
	case err := <-execCh:
		signal.Disarm()
		result.Exited = true
		if err != nil {
			return result, err
		}
	}

	_ = syscall.Munmap(mem)
	return result, nil
}
