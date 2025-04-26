package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Arlan-Z/def-betting-api/internal/data"
	eventhandler "github.com/Arlan-Z/def-betting-api/internal/deliveries/event/http"
	svcmocks "github.com/Arlan-Z/def-betting-api/internal/services/mocks"
	eventuc "github.com/Arlan-Z/def-betting-api/internal/usecases/event"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestEventHandler_GetActiveEvents_Success(t *testing.T) {
	mockService := svcmocks.NewEventService(t)
	logger := zap.NewNop()
	handler := eventhandler.NewHandler(mockService, logger)

	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	expectedEvents := []data.Event{
		{ID: uuid.NewString(), EventName: "Event 1"},
	}
	expectedDTOs := data.MapEventsToDTOs(expectedEvents)

	mockService.On("GetActiveEvents", mock.Anything).Return(expectedEvents, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Expected status 200 OK")

	var responseDTOs []data.EventDTO
	err := json.Unmarshal(rr.Body.Bytes(), &responseDTOs)
	require.NoError(t, err, "Error decoding JSON response")
	require.Equal(t, expectedDTOs, responseDTOs, "Response body does not match expected DTO")

	mockService.AssertExpectations(t)
}

func TestEventHandler_GetActiveEvents_ServiceError(t *testing.T) {
	mockService := svcmocks.NewEventService(t)
	logger := zap.NewNop()
	handler := eventhandler.NewHandler(mockService, logger)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	serviceError := errors.New("service failed")
	mockService.On("GetActiveEvents", mock.Anything).Return(nil, serviceError).Once()

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code, "Expected status 500 Internal Server Error")
	assert.Contains(t, rr.Body.String(), "Internal Server Error", "Response body should contain internal server error message")

	mockService.AssertExpectations(t)
}

func TestEventHandler_FinalizeEvent_Success(t *testing.T) {
	mockService := svcmocks.NewEventService(t)
	logger := zap.NewNop()
	handler := eventhandler.NewHandler(mockService, logger)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	eventID := uuid.NewString()
	result := data.HomeWin
	requestBody := map[string]string{"result": string(result)}
	jsonBody, _ := json.Marshal(requestBody)

	mockService.On("FinalizeEvent", mock.Anything, eventID, result).Return(nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID+"/finalize", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, "Expected status 200 OK")

	mockService.AssertExpectations(t)
}

func TestEventHandler_FinalizeEvent_ValidationError(t *testing.T) {
	mockService := svcmocks.NewEventService(t)
	logger := zap.NewNop()
	handler := eventhandler.NewHandler(mockService, logger)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	eventID := uuid.NewString()
	requestBody := map[string]string{"result": "InvalidResult"}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID+"/finalize", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code, "Expected status 400 Bad Request")
	assert.Contains(t, rr.Body.String(), "Validation error", "Response body should contain validation error message")

	mockService.AssertNotCalled(t, "FinalizeEvent", mock.Anything, mock.Anything, mock.Anything)
}

func TestEventHandler_FinalizeEvent_BadJson(t *testing.T) {
	mockService := svcmocks.NewEventService(t)
	logger := zap.NewNop()
	handler := eventhandler.NewHandler(mockService, logger)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	eventID := uuid.NewString()
	badJsonBody := []byte(`{"result":"HomeWin"`)

	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID+"/finalize", bytes.NewBuffer(badJsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code, "Expected status 400 Bad Request")
	assert.Contains(t, rr.Body.String(), "Invalid request body", "Response body should contain JSON error message")

	mockService.AssertNotCalled(t, "FinalizeEvent", mock.Anything, mock.Anything, mock.Anything)
}

func TestEventHandler_FinalizeEvent_NotFound(t *testing.T) {
	mockService := svcmocks.NewEventService(t)
	logger := zap.NewNop()
	handler := eventhandler.NewHandler(mockService, logger)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	eventID := uuid.NewString()
	result := data.AwayWin
	requestBody := map[string]string{"result": string(result)}
	jsonBody, _ := json.Marshal(requestBody)

	notFoundError := eventuc.ErrEventNotFound
	mockService.On("FinalizeEvent", mock.Anything, eventID, result).Return(notFoundError).Once()

	req := httptest.NewRequest(http.MethodPost, "/events/"+eventID+"/finalize", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code, "Expected status 404 Not Found")
	assert.Contains(t, rr.Body.String(), "Event not found", "Response body should contain 'Event not found' message")

	mockService.AssertExpectations(t)
}
