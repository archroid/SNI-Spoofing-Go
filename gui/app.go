package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"sni-spoofing-gui/autostart"
	"sni-spoofing-go/guiapi"
	"sni-spoofing-go/helper"
	"sni-spoofing-go/helper/spawn"
	"sni-spoofing-go/proxy"
)

type ProxyConfig = guiapi.ProxyConfig
type ProxyStatus = guiapi.ProxyStatus
type LogEvent = guiapi.LogEvent
type TestResult = guiapi.TestResult
type TestSummary = guiapi.TestSummary
type TestPreflight = guiapi.TestPreflight

// App is the Wails-bound application object. Privileged work runs in a separate
// elevated helper process; this struct is the unprivileged GUI front-end.
type App struct {
	ctx context.Context

	mu             sync.Mutex
	manager        *helper.Manager
	status         ProxyStatus
	minimizeToTray bool
	isQuitting     bool
	windowVisible  bool
	lastConfig     ProxyConfig
	trayManager    *TrayManager
}

func NewApp() *App {
	a := &App{
		minimizeToTray: true,
		windowVisible:  true,
	}
	a.lastConfig = a.GetDefaultConfig()
	a.trayManager = NewTrayManager(a)
	return a
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	exe, err := spawn.SelfExe()
	if err != nil {
		a.emitLog("error", fmt.Sprintf("locate executable: %v", err))
		return
	}
	a.manager = helper.NewManager(exe, helper.EventHandler{
		OnLog:        a.onHelperLog,
		OnStatus:     a.onHelperStatus,
		OnTestResult: a.onHelperTestResult,
		OnDisconnect: a.onHelperDisconnect,
	})
}

func (a *App) shutdown(ctx context.Context) {
	if a.manager != nil {
		_ = a.manager.Close()
	}
}

func (a *App) onBeforeClose(ctx context.Context) bool {
	a.mu.Lock()
	minToTray := a.minimizeToTray
	quitting := a.isQuitting
	a.mu.Unlock()

	if minToTray && !quitting {
		runtime.WindowHide(ctx)
		a.mu.Lock()
		a.windowVisible = false
		a.mu.Unlock()
		if a.trayManager != nil {
			a.trayManager.UpdateShowHideText(false)
		}
		return true
	}
	return false
}

func (a *App) onHelperLog(ev LogEvent) {
	a.emitLog(ev.Level, ev.Message)
}

func (a *App) onHelperStatus(st ProxyStatus) {
	a.mu.Lock()
	a.status = st
	a.mu.Unlock()
	a.emitStatus(st)
	if a.trayManager != nil {
		a.trayManager.UpdateStatus(st)
	}
}

func (a *App) onHelperTestResult(row TestResult) {
	a.emitTestResult(row)
}

func (a *App) onHelperDisconnect(err error) {
	a.mu.Lock()
	active := a.status.Running || a.status.Testing
	a.mu.Unlock()
	if active && err != nil && !errors.Is(err, io.EOF) {
		a.emitLog("warn", fmt.Sprintf("helper disconnected: %v", err))
	}
	if active {
		a.clearStatus()
	}
}

func (a *App) clearStatus() {
	a.mu.Lock()
	a.status = ProxyStatus{}
	st := a.status
	a.mu.Unlock()
	a.emitStatus(st)
	if a.trayManager != nil {
		a.trayManager.UpdateStatus(st)
	}
}

func (a *App) GetDefaultConfig() ProxyConfig {
	return guiapi.DefaultConfig(string(proxy.DefaultInjectorMode()))
}

func (a *App) UTLSPresets() []string {
	return []string{
		"none", "firefox", "chrome", "edge", "safari", "ios", "qq", "360browser",
	}
}

func (a *App) InjectorModes() []string {
	return []string{"active", "passive"}
}

func (a *App) Status() ProxyStatus {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.status
}

func (a *App) helperProgress() helper.ProgressFunc {
	return func(message string) {
		a.emitLog("info", message)
	}
}

func (a *App) helperClient(ctx context.Context) (*helper.Client, error) {
	if a.manager == nil {
		return nil, errors.New("helper manager not initialized")
	}
	return a.manager.Ensure(ctx, a.helperProgress())
}

func (a *App) Start(cfg ProxyConfig) error {
	if err := guiapi.ValidateConfig(cfg); err != nil {
		return err
	}
	a.mu.Lock()
	a.lastConfig = cfg
	a.mu.Unlock()

	a.emitLog("info", "Starting proxy…")
	client, err := a.helperClient(a.ctx)
	if err != nil {
		return err
	}
	return client.Start(a.ctx, cfg)
}

func (a *App) Stop() error {
	a.mu.Lock()
	active := a.status.Running || a.status.Testing
	mgr := a.manager
	a.mu.Unlock()
	if !active {
		return nil
	}
	if mgr == nil {
		a.clearStatus()
		return nil
	}
	if err := mgr.Stop(a.ctx); err != nil {
		if errors.Is(err, helper.ErrNotConnected) {
			a.clearStatus()
			return nil
		}
		a.emitLog("warn", fmt.Sprintf("stop: %v", err))
	}
	return nil
}

func (a *App) RunTest(cfg ProxyConfig) (TestSummary, error) {
	if err := guiapi.ValidateConfig(cfg); err != nil {
		return TestSummary{}, err
	}
	a.mu.Lock()
	a.lastConfig = cfg
	a.mu.Unlock()

	a.emitLog("info", "Running test matrix…")
	client, err := a.helperClient(a.ctx)
	if err != nil {
		return TestSummary{}, err
	}
	return client.RunTest(a.ctx, cfg)
}

func (a *App) GetAutoStart() (bool, error) {
	return autostart.IsEnabled()
}

func (a *App) SetAutoStart(enabled bool) error {
	var err error
	if enabled {
		err = autostart.Enable()
	} else {
		err = autostart.Disable()
	}
	if err == nil && a.trayManager != nil {
		a.trayManager.UpdateAutostart(enabled)
	}
	return err
}

func (a *App) ToggleAutoStart() error {
	curr, err := a.GetAutoStart()
	if err != nil {
		return err
	}
	return a.SetAutoStart(!curr)
}

func (a *App) GetMinimizeToTray() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.minimizeToTray
}

func (a *App) SetMinimizeToTray(enabled bool) {
	a.mu.Lock()
	a.minimizeToTray = enabled
	a.mu.Unlock()
	if a.trayManager != nil {
		a.trayManager.UpdateMinimizeToTray(enabled)
	}
}

func (a *App) ToggleMinimizeToTray() {
	curr := a.GetMinimizeToTray()
	a.SetMinimizeToTray(!curr)
}

func (a *App) ToggleWindowVisibility() {
	if a.ctx == nil {
		return
	}
	a.mu.Lock()
	visible := a.windowVisible
	a.mu.Unlock()

	if visible {
		runtime.WindowHide(a.ctx)
		a.mu.Lock()
		a.windowVisible = false
		a.mu.Unlock()
	} else {
		runtime.WindowShow(a.ctx)
		runtime.WindowUnminimise(a.ctx)
		a.mu.Lock()
		a.windowVisible = true
		a.mu.Unlock()
	}
	if a.trayManager != nil {
		a.trayManager.UpdateShowHideText(!visible)
	}
}

func (a *App) ToggleProxy() {
	a.mu.Lock()
	st := a.status
	cfg := a.lastConfig
	a.mu.Unlock()

	if st.Running || st.Testing {
		_ = a.Stop()
	} else {
		_ = a.Start(cfg)
	}
}

func (a *App) Quit() {
	a.mu.Lock()
	a.isQuitting = true
	a.mu.Unlock()
	_ = a.Stop()
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
}

func (a *App) emitLog(level, message string) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "log", LogEvent{Level: level, Message: message})
}

func (a *App) emitStatus(s ProxyStatus) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "status", s)
}

func (a *App) emitTestResult(r TestResult) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "test_result", r)
}
