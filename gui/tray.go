package main

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/energye/systray"
	"sni-spoofing-go/guiapi"
)

//go:embed build/appicon.png
var trayIconBytes []byte

type TrayManager struct {
	app *App

	mu              sync.Mutex
	mShowHide       *systray.MenuItem
	mStatus         *systray.MenuItem
	mToggleProxy    *systray.MenuItem
	mAutostart      *systray.MenuItem
	mMinimizeToTray *systray.MenuItem
	mQuit           *systray.MenuItem
	ready           bool
	stopFunc        func()
}

func NewTrayManager(app *App) *TrayManager {
	return &TrayManager{app: app}
}

func (tm *TrayManager) Start() {
	start, end := systray.RunWithExternalLoop(tm.onReady, tm.onExit)
	tm.stopFunc = end
	start()
}

func (tm *TrayManager) Stop() {
	if tm.stopFunc != nil {
		tm.stopFunc()
	}
}

func (tm *TrayManager) onReady() {
	systray.SetIcon(trayIconBytes)
	systray.SetTitle("SNI Spoofing")
	systray.SetTooltip("SNI Spoofing - DPI Bypass Proxy")

	tm.mShowHide = systray.AddMenuItem("Show Window", "Show or hide the SNI Spoofing window")
	tm.mShowHide.Click(func() {
		tm.app.ToggleWindowVisibility()
	})

	systray.AddSeparator()

	tm.mStatus = systray.AddMenuItem("Status: Stopped", "")
	tm.mStatus.Disable()

	tm.mToggleProxy = systray.AddMenuItem("Start Proxy", "Start or stop the proxy")
	tm.mToggleProxy.Click(func() {
		tm.app.ToggleProxy()
	})

	systray.AddSeparator()

	tm.mAutostart = systray.AddMenuItemCheckbox("Autostart on Boot", "Start automatically when system boots", false)
	tm.mAutostart.Click(func() {
		_ = tm.app.ToggleAutoStart()
	})

	tm.mMinimizeToTray = systray.AddMenuItemCheckbox("Minimize to Tray", "Minimize to system tray on close", true)
	tm.mMinimizeToTray.Click(func() {
		tm.app.ToggleMinimizeToTray()
	})

	systray.AddSeparator()

	tm.mQuit = systray.AddMenuItem("Quit", "Quit SNI Spoofing")
	tm.mQuit.Click(func() {
		tm.app.Quit()
	})

	systray.SetOnClick(func(menu systray.IMenu) {
		tm.app.ToggleWindowVisibility()
	})
	systray.SetOnDClick(func(menu systray.IMenu) {
		tm.app.ToggleWindowVisibility()
	})

	tm.mu.Lock()
	tm.ready = true
	tm.mu.Unlock()

	// Sync initial state
	if autostartEnabled, err := tm.app.GetAutoStart(); err == nil {
		tm.UpdateAutostart(autostartEnabled)
	}
	tm.UpdateMinimizeToTray(tm.app.GetMinimizeToTray())
	tm.UpdateStatus(tm.app.Status())
}

func (tm *TrayManager) onExit() {
	fmt.Println("[Tray] onExit invoked!")
}

func (tm *TrayManager) UpdateStatus(st guiapi.ProxyStatus) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if !tm.ready {
		return
	}

	if st.Testing {
		tm.mStatus.SetTitle("Status: Testing matrix…")
		tm.mToggleProxy.SetTitle("Stop Proxy")
		tm.mToggleProxy.Disable()
	} else if st.Running {
		tm.mStatus.SetTitle(fmt.Sprintf("Status: Running (%s)", st.ListenAddr))
		tm.mToggleProxy.SetTitle("Stop Proxy")
		tm.mToggleProxy.Enable()
	} else {
		tm.mStatus.SetTitle("Status: Stopped")
		tm.mToggleProxy.SetTitle("Start Proxy")
		tm.mToggleProxy.Enable()
	}
}

func (tm *TrayManager) UpdateAutostart(enabled bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if !tm.ready || tm.mAutostart == nil {
		return
	}
	if enabled {
		tm.mAutostart.Check()
	} else {
		tm.mAutostart.Uncheck()
	}
}

func (tm *TrayManager) UpdateMinimizeToTray(enabled bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if !tm.ready || tm.mMinimizeToTray == nil {
		return
	}
	if enabled {
		tm.mMinimizeToTray.Check()
	} else {
		tm.mMinimizeToTray.Uncheck()
	}
}

func (tm *TrayManager) UpdateShowHideText(windowVisible bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if !tm.ready || tm.mShowHide == nil {
		return
	}
	if windowVisible {
		tm.mShowHide.SetTitle("Hide Window")
	} else {
		tm.mShowHide.SetTitle("Show Window")
	}
}
