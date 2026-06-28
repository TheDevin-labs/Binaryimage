
package signal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type FaultInfo struct {
	Signal  string
	Code    int
	Address uintptr
}

func (f *FaultInfo) Error() string {
	return fmt.Sprintf("hardware fault: signal=%s code=%d addr=0x%x",
		f.Signal, f.Code, f.Address)
}

func Arm(onFault func(*FaultInfo)) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch,
		syscall.SIGSEGV,
		syscall.SIGBUS,
		syscall.SIGILL,
		syscall.SIGFPE,
		syscall.SIGTRAP,
	)

	go func() {
		sig := <-ch
		info := &FaultInfo{
			Signal: sig.String(),
		}

		if si, ok := sig.(syscall.Signal); ok {
			info.Code = int(si)
		}

		onFault(info)
	}()
}

func Disarm() {
	signal.Reset(
		syscall.SIGSEGV,
		syscall.SIGBUS,
		syscall.SIGILL,
		syscall.SIGFPE,
		syscall.SIGTRAP,
	)
}
