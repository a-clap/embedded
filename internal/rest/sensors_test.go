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

type GetSensorMock struct {
	mock.Mock
}

type SensorsSuite struct {
	suite.Suite
	mocker *GetSensorMock
	srv    *rest.Server
	req    *http.Request
	resp   *httptest.ResponseRecorder
}

func (g *GetSensorMock) Sensors() ([]rest.Sensor, error) {
	args := g.Called()
	return args.Get(0).([]rest.Sensor), args.Error(1)
}

func TestSensorsSuite(t *testing.T) {
	suite.Run(t, new(SensorsSuite))
}

func (s *SensorsSuite) SetupTest() {
	s.mocker = new(GetSensorMock)
	s.srv, _ = rest.New(rest.WithFormat(rest.JSON), rest.WithSensorHandler(s.mocker))
	s.req, _ = http.NewRequest(http.MethodGet, rest.RoutesGetSensor, nil)
	s.resp = httptest.NewRecorder()
}

func (s *SensorsSuite) TestLackOfInterface() {
	s.srv, _ = rest.New(rest.WithFormat(rest.JSON))

	s.srv.ServeHTTP(s.resp, s.req)

	body, _ := io.ReadAll(s.resp.Body)

	s.Equal(http.StatusInternalServerError, s.resp.Code)
	s.JSONEq(rest.ErrNotImplemented.JSON(), string(body))
}

func (s *SensorsSuite) TestErrorOnInterfaceAccess() {
	s.mocker.On("Sensors").Return([]rest.Sensor{}, errors.New("lol nope"))

	s.srv.ServeHTTP(s.resp, s.req)

	body, _ := io.ReadAll(s.resp.Body)

	s.Equal(http.StatusInternalServerError, s.resp.Code)
	s.JSONEq(rest.ErrNotFound.JSON(), string(body))
}

func (s *SensorsSuite) TestCorrectSensors() {
	sensors := []rest.Sensor{
		{ID: "1", Name: "sensor_1", Temperature: 1.23},
		{ID: "blah", Name: "sensor_2", Temperature: 3.45},
		{ID: "hey you", Name: "sensor_3", Temperature: 5.63},
	}
	expected, _ := json.Marshal(sensors)

	s.mocker.On("Sensors").Return(sensors, nil)

	s.srv.ServeHTTP(s.resp, s.req)

	body, _ := io.ReadAll(s.resp.Body)

	s.Equal(http.StatusOK, s.resp.Code)

	s.JSONEq(string(expected), string(body))
}
