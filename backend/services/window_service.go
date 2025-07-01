package services

import (
	"github.com/greboid/tithon/irc"
)

type UpdateTrigger interface {
	SetPendingUpdate()
}

type WindowChangedTrigger interface {
	SetWindowChanged()
}

func NewWindowService(updateTrigger UpdateTrigger, windowChangedTrigger WindowChangedTrigger) *WindowService {
	return &WindowService{
		pendingUpdate: updateTrigger,
		windowChanged: windowChangedTrigger,
	}
}

func (ws *WindowService) SetActiveWindow(window *irc.Window) {
	ws.activeLock.Lock()
	defer ws.activeLock.Unlock()

	if ws.activeWindow != nil {
		ws.activeWindow.SetActive(false)
	}
	if window != nil {
		window.SetActive(true)
	}
	ws.activeWindow = window
	ws.pendingUpdate.SetPendingUpdate()
	ws.windowChanged.SetWindowChanged()
}

func (ws *WindowService) GetActiveWindow() *irc.Window {
	ws.activeLock.RLock()
	defer ws.activeLock.RUnlock()
	return ws.activeWindow
}

func (ws *WindowService) OnWindowRemoved(removedWindow *irc.Window, serverList *ServerList) {
	currentActive := ws.GetActiveWindow()
	if currentActive == removedWindow {
		var newActiveWindow *irc.Window
		if serverList != nil && len(serverList.OrderedList) > 0 {
			for _, item := range serverList.OrderedList {
				if item.Window != removedWindow {
					newActiveWindow = item.Window
					break
				}
			}
		}
		ws.SetActiveWindow(newActiveWindow)
	}
}
