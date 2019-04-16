// TODO Package documentation
package interruptHooks

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type InterruptHook func()

var hookAddChannel = make(chan InterruptHook)
var setShutdownChannel = make(chan bool)
var sigtermChannel = make(chan os.Signal, 1)

func init() {
	signal.Notify(sigtermChannel, os.Interrupt, syscall.SIGTERM)
	go listenForSignalsOrHooks()
}

func listenForSignalsOrHooks() {
	shouldShutdown := true
	for {
		var interruptHooks []InterruptHook
	listenToChannels:
		for {
			select {
			case newHook := <-hookAddChannel:
				interruptHooks = append(interruptHooks, newHook)
				break
			case newShutdownValue := <-setShutdownChannel:
				shouldShutdown = newShutdownValue
				break
			case <-sigtermChannel:
				break listenToChannels
			}
		}
		callInterruptHooks(interruptHooks)
		if shouldShutdown {
			os.Exit(1)
			return
		}
	}
}

func callInterruptHooks(interruptHooks []InterruptHook) {
	var waitGroup sync.WaitGroup
	for _, hook := range interruptHooks {
		hookLocal := hook
		go func() {
			waitGroup.Add(1)
			defer waitGroup.Done()
			defer func() {
				recover()
			}()
			hookLocal()
		}() // Run in separate goroutine in order to prevent a hook from panicking or blocking and thereby preventing others from running
	}
	waitGroup.Wait()
}

func AddHook(hook InterruptHook) {
	hookAddChannel <- hook
}

func SetShouldShutdown(shouldShutdown bool) {
	setShutdownChannel <- shouldShutdown
}
