package handlers

import (
	"sync"
)

// StateManager управляет состояниями пользователей
type StateManager struct {
	userStates  map[int64]string
	userData    map[int64]map[string]interface{}
	userHistory map[int64][]string
	mutex       sync.RWMutex
}

// NewStateManager создает новый менеджер состояний
func NewStateManager() *StateManager {
	return &StateManager{
		userStates:  make(map[int64]string),
		userData:    make(map[int64]map[string]interface{}),
		userHistory: make(map[int64][]string),
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

// DebugUserState выводит отладочную информацию о состоянии пользователя
func (sm *StateManager) DebugUserState(userID int64) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	state := sm.userStates[userID]
	data := sm.userData[userID]

	InfoLog.Printf("DebugUserState: user %d, state: %s, data: %+v", userID, state, data)
}

// GetAllUserData возвращает все данные пользователя для отладки
func (sm *StateManager) GetAllUserData(userID int64) map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if sm.userData[userID] == nil {
		return make(map[string]interface{})
	}

	// Создаем копию для безопасного использования
	result := make(map[string]interface{})
	for k, v := range sm.userData[userID] {
		result[k] = v
	}
	return result
}

// ClearUserDataByKey очищает конкретный ключ данных пользователя
func (sm *StateManager) ClearUserDataByKey(userID int64, key string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.userData[userID] != nil {
		delete(sm.userData[userID], key)
	}
}

// UserHasState проверяет, есть ли у пользователя состояние
func (sm *StateManager) UserHasState(userID int64) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	_, exists := sm.userStates[userID]
	return exists
}

// В state_manager.go добавьте:
func (sm *StateManager) PrintDebugInfo(userID int64) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	state := sm.userStates[userID]
	data := sm.userData[userID]

	InfoLog.Printf("StateManager Debug - User: %d, State: %s, Data: %+v", userID, state, data)
}

// PushState добавляет состояние в историю
func (sm *StateManager) PushState(userID int64, state string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.userHistory[userID] == nil {
		sm.userHistory[userID] = make([]string, 0)
	}

	// Добавляем текущее состояние в историю
	sm.userHistory[userID] = append(sm.userHistory[userID], state)

	// Ограничиваем историю последними 10 состояниями
	if len(sm.userHistory[userID]) > 10 {
		sm.userHistory[userID] = sm.userHistory[userID][1:]
	}
}

// PopState возвращает предыдущее состояние и удаляет его из истории
func (sm *StateManager) PopState(userID int64) (string, bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Убираем проверку на nil, так как len() для nil слайсов возвращает 0
	if len(sm.userHistory[userID]) == 0 {
		return "", false
	}

	// Берем последнее состояние
	lastIndex := len(sm.userHistory[userID]) - 1
	previousState := sm.userHistory[userID][lastIndex]

	// Удаляем его из истории
	sm.userHistory[userID] = sm.userHistory[userID][:lastIndex]

	return previousState, true
}

// GetPreviousState возвращает предыдущее состояние без удаления
func (sm *StateManager) GetPreviousState(userID int64) (string, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Убираем проверку на nil, так как len() для nil слайсов возвращает 0
	if len(sm.userHistory[userID]) == 0 {
		return "", false
	}

	lastIndex := len(sm.userHistory[userID]) - 1
	return sm.userHistory[userID][lastIndex], true
}

// ClearHistory очищает историю пользователя
func (sm *StateManager) ClearHistory(userID int64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.userHistory, userID)
}
