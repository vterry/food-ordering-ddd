package saga

import (
	"encoding/json"
	"time"
)

type SagaStatus string

const (
	SagaRunning      SagaStatus = "RUNNING"
	SagaCompensating SagaStatus = "COMPENSATING"
	SagaCompleted    SagaStatus = "COMPLETED"
	SagaFailed       SagaStatus = "FAILED"
)

// SagaState represents the state of a saga instance.
type SagaState struct {
	orderID     string
	currentStep string
	status      SagaStatus
	data        map[string]interface{}
	updatedAt   time.Time
}

func NewSagaState(orderID string, initialStep string) *SagaState {
	return &SagaState{
		orderID:     orderID,
		currentStep: initialStep,
		status:      SagaRunning,
		data:        make(map[string]interface{}),
		updatedAt:   time.Now(),
	}
}

func (s *SagaState) OrderID() string { return s.orderID }
func (s *SagaState) CurrentStep() string { return s.currentStep }
func (s *SagaState) Status() SagaStatus { return s.status }
func (s *SagaState) Data() map[string]interface{} { return s.data }
func (s *SagaState) UpdatedAt() time.Time { return s.updatedAt }

func (s *SagaState) SetStep(step string) {
	s.currentStep = step
	s.updatedAt = time.Now()
}

func (s *SagaState) SetStatus(status SagaStatus) {
	s.status = status
	s.updatedAt = time.Now()
}

func (s *SagaState) AddData(key string, value interface{}) {
	s.data[key] = value
	s.updatedAt = time.Now()
}

func (s *SagaState) DataJSON() ([]byte, error) {
	return json.Marshal(s.data)
}

// MapFromPersistence reconstructs the saga state from persistence.
func MapFromPersistence(orderID string, currentStep string, status SagaStatus, dataJSON []byte, updatedAt time.Time) (*SagaState, error) {
	data := make(map[string]interface{})
	if len(dataJSON) > 0 {
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			return nil, err
		}
	}
	
	return &SagaState{
		orderID:     orderID,
		currentStep: currentStep,
		status:      status,
		data:        data,
		updatedAt:   updatedAt,
	}, nil
}
