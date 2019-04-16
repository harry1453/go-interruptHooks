package interruptHooks

import (
	"fmt"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func runTest(hooks ...InterruptHook) error {
	signal.Ignore(syscall.SIGTERM)
	SetShouldShutdown(false)

	numberOfHooks := len(hooks)
	hooksRun := make([]bool, numberOfHooks)
	for i, hook := range hooks {
		iLocal := i
		hookLocal := hook
		AddHook(func() {
			hooksRun[iLocal] = true
			hookLocal()
		})
	}
	time.Sleep(time.Duration(numberOfHooks) * time.Millisecond)

	sigtermChannel <- syscall.SIGTERM

	time.Sleep(time.Duration(numberOfHooks) * time.Millisecond)

	for i, run := range hooksRun {
		if !run {
			return fmt.Errorf("hook at position %d did not run", i)
		}
	}
	return nil
}

func Test_EmptyHooks(t *testing.T) {
	// Test 5 empty hooks
	if err := runTest(func() {}, func() {}, func() {}, func() {}, func() {}); err != nil {
		t.Fatal(err)
	}
}

func Test_PanickingHooks(t *testing.T) {
	// Test 5 panicking hooks
	if err := runTest(func() {panic(nil)}, func() {panic(nil)}, func() {panic(nil)}, func() {panic(nil)}, func() {panic(nil)}); err != nil {
		t.Fatal(err)
	}
}
