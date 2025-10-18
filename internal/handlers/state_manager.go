package handlers

import (
	"sync"
)

// StateManager управляет состояниями пользователей
type StateManager struct {
	userStates map[int64]string
	userData   map[int64]map[string]interface{}
	mutex      sync.RWMutex
}

// NewStateManager создает новый менеджер состояний
func NewStateManager() *StateManager {
	return &StateManager{
		userStates: make(map[int64]string),
		userData:   make(map[int64]map[string]interface{}),
	}
}

// SetUserState устанавливает состояние пользователя
func (sm *StateManager) SetUserState(userID int64, state string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.userStates[userID] = state
}

// GetUserState возвращает состояние пользователя
func (sm *StateManager) GetUserState(userID int64) string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.userStates[userID]
}

// ClearUserState очищает состояние пользователя
func (sm *StateManager) ClearUserState(userID int64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.userStates, userID)
	delete(sm.userData, userID)
}

// SetUserData сохраняет данные пользователя
func (sm *StateManager) SetUserData(userID int64, key string, value interface{}) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.userData[userID] == nil {
		sm.userData[userID] = make(map[string]interface{})
	}
	sm.userData[userID][key] = value
}

// GetUserData возвращает данные пользователя
func (sm *StateManager) GetUserData(userID int64, key string) interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if userData, exists := sm.userData[userID]; exists {
		return userData[key]
	}
	return nil
}

// GetUserDataInt возвращает данные пользователя как int
func (sm *StateManager) GetUserDataInt(userID int64, key string) (int, bool) {
	value := sm.GetUserData(userID, key)
	if value == nil {
		return 0, false
	}

	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	default:
		return 0, false
	}
}

// ClearUserData очищает все данные пользователя
func (sm *StateManager) ClearUserData(userID int64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.userData, userID)
}
