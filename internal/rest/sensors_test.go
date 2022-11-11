package rest_test

import (
	"encoding/json"
	"errors"
	"github.com/a-clap/iot/internal/rest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type SensorsSuite struct {
	suite.Suite
}

type GetSensorMock struct {
	mock.Mock
}

func (g *GetSensorMock) Sensors() ([]rest.Sensor, error) {
	args := g.Called()
	return args.Get(0).([]rest.Sensor), args.Error(1)
}

var (
	mocker *GetSensorMock
	srv    *rest.Server
	req    *http.Request
	resp   *httptest.ResponseRecorder
)

func TestSensorsSuite(t *testing.T) {
	suite.Run(t, new(SensorsSuite))
}

func (s *SensorsSuite) SetupTest() {
	mocker = new(GetSensorMock)
	srv = rest.New(rest.JSON, mocker)
	req, _ = http.NewRequest(http.MethodGet, rest.RoutesGetSensor, nil)
	resp = httptest.NewRecorder()
}

func (s *SensorsSuite) TestLackOfInterface() {
	srv = rest.New(rest.JSON)

	srv.ServeHTTP(resp, req)

	body, _ := io.ReadAll(resp.Body)

	s.Equal(http.StatusInternalServerError, resp.Code)
	s.JSONEq(rest.ErrNotImplemented.JSON(), string(body))
}

func (s *SensorsSuite) TestErrorOnInterfaceAccess() {
	mocker.On("Sensors").Return([]rest.Sensor{}, errors.New("lol nope"))

	srv.ServeHTTP(resp, req)

	body, _ := io.ReadAll(resp.Body)

	s.Equal(http.StatusInternalServerError, resp.Code)
	s.JSONEq(rest.ErrNotFound.JSON(), string(body))
}

func (s *SensorsSuite) TestCorrectSensors() {
	sensors := []rest.Sensor{
		{ID: "1", Name: "sensor_1", Temperature: 1.23},
		{ID: "blah", Name: "sensor_2", Temperature: 3.45},
		{ID: "hey you", Name: "sensor_3", Temperature: 5.63},
	}
	expected, _ := json.Marshal(sensors)

	mocker.On("Sensors").Return(sensors, nil)

	srv.ServeHTTP(resp, req)

	body, _ := io.ReadAll(resp.Body)

	s.Equal(http.StatusOK, resp.Code)

	s.JSONEq(string(expected), string(body))
}
