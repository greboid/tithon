package services

import (
	"sync"
)

type InputHistoryService struct {
	inputHistory    []string
	historyPosition int
	historyLock     sync.RWMutex
}

func NewInputHistoryService() *InputHistoryService {
	return &InputHistoryService{
		inputHistory:    make([]string, 0),
		historyPosition: -1,
	}
}

func (ihs *InputHistoryService) AddToHistory(input string) {
	ihs.historyLock.Lock()
	defer ihs.historyLock.Unlock()

	if input == "" || (len(ihs.inputHistory) > 0 && ihs.inputHistory[len(ihs.inputHistory)-1] == input) {
		return
	}

	ihs.inputHistory = append(ihs.inputHistory, input)
	ihs.historyPosition = -1
}

func (ihs *InputHistoryService) GetHistoryItem(position int) string {
	ihs.historyLock.Lock()
	defer ihs.historyLock.Unlock()

	if position < 0 || position >= len(ihs.inputHistory) {
		return ""
	}

	reverseIndex := len(ihs.inputHistory) - 1 - position
	if reverseIndex < 0 || reverseIndex >= len(ihs.inputHistory) {
		return ""
	}

	return ihs.inputHistory[reverseIndex]
}

func (ihs *InputHistoryService) GetHistoryLength() int {
	ihs.historyLock.RLock()
	defer ihs.historyLock.RUnlock()
	return len(ihs.inputHistory)
}

func (ihs *InputHistoryService) ResetPosition() {
	ihs.historyLock.RLock()
	defer ihs.historyLock.RUnlock()
	ihs.historyPosition = -1
}

func (ihs *InputHistoryService) GetCurrentPosition() int {
	ihs.historyLock.RLock()
	defer ihs.historyLock.RUnlock()
	return ihs.historyPosition
}

func (ihs *InputHistoryService) NavigateUp(currentInput string) string {
	ihs.historyLock.Lock()
	defer ihs.historyLock.Unlock()

	if len(ihs.inputHistory) == 0 {
		return ""
	}

	if currentInput != "" && (len(ihs.inputHistory) == 0 || ihs.inputHistory[len(ihs.inputHistory)-1] != currentInput) {
		ihs.inputHistory = append(ihs.inputHistory, currentInput)
		ihs.historyPosition = len(ihs.inputHistory) - 1
		return currentInput
	}

	if ihs.historyPosition == -1 {
		ihs.historyPosition = len(ihs.inputHistory) - 1
	} else if ihs.historyPosition > 0 {
		ihs.historyPosition--
	}

	if ihs.historyPosition >= 0 && ihs.historyPosition < len(ihs.inputHistory) {
		return ihs.inputHistory[ihs.historyPosition]
	}
	return ""
}

func (ihs *InputHistoryService) NavigateDown() string {
	ihs.historyLock.Lock()
	defer ihs.historyLock.Unlock()

	if len(ihs.inputHistory) == 0 || ihs.historyPosition == -1 {
		return ""
	}

	if ihs.historyPosition < len(ihs.inputHistory)-1 {
		ihs.historyPosition++
		if ihs.historyPosition < len(ihs.inputHistory) {
			return ihs.inputHistory[ihs.historyPosition]
		}
	} else {
		ihs.historyPosition = -1
		return ""
	}

	return ""
}
