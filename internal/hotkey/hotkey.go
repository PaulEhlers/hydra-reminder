package hotkey

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                 = windows.NewLazySystemDLL("user32.dll")
	procRegisterHotKey     = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey   = user32.NewProc("UnregisterHotKey")
	procGetMessageW        = user32.NewProc("GetMessageW")
	procPostThreadMessageW = user32.NewProc("PostThreadMessageW")
)

const (
	WM_HOTKEY = 0x0312
	WM_QUIT   = 0x0012

	hotkeyID = 1
)

var (
	mu             sync.Mutex
	loopThreadId   uint32
	loopDoneCh     chan struct{}
	hotkeyCallback func()
)

type msg struct {
	Hwnd     windows.Handle
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       point
	LPrivate uint32
}

type point struct {
	X int32
	Y int32
}

func Init(callback func()) {
	hotkeyCallback = callback
}

type registerResult struct {
	threadId uint32
	doneCh   chan struct{}
	err      error
}

func Register(modifiers uint32, key uint32) error {
	mu.Lock()
	defer mu.Unlock()

	// Stop any existing loop fully before proceeding
	if loopThreadId != 0 {
		procPostThreadMessageW.Call(uintptr(loopThreadId), WM_QUIT, 0, 0)
		<-loopDoneCh // Wait for thread to exit
		loopThreadId = 0
		loopDoneCh = nil
	}

	resCh := make(chan registerResult, 1)

	go func() {
		// Lock goroutine to an OS thread to run standard Windows message loop
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		tid := windows.GetCurrentThreadId()

		ret, _, err := procRegisterHotKey.Call(
			0,                  // hwnd
			uintptr(hotkeyID),  // id
			uintptr(modifiers), // fsModifiers
			uintptr(key),       // vk
		)

		done := make(chan struct{})

		if ret == 0 {
			close(done)
			resCh <- registerResult{err: fmt.Errorf("RegisterHotKey failed: %v", err)}
			return
		}

		resCh <- registerResult{threadId: tid, doneCh: done, err: nil}

		defer func() {
			procUnregisterHotKey.Call(0, uintptr(hotkeyID))
			close(done)
		}()

		var m msg
		for {
			// Peek/Get message.
			ret, _, _ := procGetMessageW.Call(
				uintptr(unsafe.Pointer(&m)),
				0,
				0,
				0,
			)
			if int32(ret) == -1 || ret == 0 || m.Message == WM_QUIT {
				return
			}

			if m.Message == WM_HOTKEY && m.WParam == hotkeyID {
				if hotkeyCallback != nil {
					go hotkeyCallback()
				}
			}
		}
	}()

	res := <-resCh
	if res.err != nil {
		return res.err
	}

	loopThreadId = res.threadId
	loopDoneCh = res.doneCh

	return nil
}

func Unregister() {
	mu.Lock()
	defer mu.Unlock()

	if loopThreadId != 0 {
		procPostThreadMessageW.Call(uintptr(loopThreadId), WM_QUIT, 0, 0)
		<-loopDoneCh // Wait for thread to cleanly unregister
		loopThreadId = 0
		loopDoneCh = nil
	}
}
