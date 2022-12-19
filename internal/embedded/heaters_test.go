package embedded_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/a-clap/iot/internal/embedded"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

type HeaterTestSuite struct {
	suite.Suite
	req  *http.Request
	resp *httptest.ResponseRecorder
	mock map[embedded.HardwareID]*HeaterMock
}

type HeaterMock struct {
	mock.Mock
}

func TestHeaterTestSuite(t *testing.T) {
	suite.Run(t, new(HeaterTestSuite))
}

func toJSON(obj any) string {
	b, _ := json.Marshal(obj)
	return string(b)
}

func (t *HeaterTestSuite) heaters() map[embedded.HardwareID]embedded.Heater {
	heater := make(map[embedded.HardwareID]embedded.Heater)

	for k, v := range t.mock {
		heater[k] = v
	}
	return heater
}

func (t *HeaterTestSuite) SetupTest() {
	gin.DefaultWriter = io.Discard
	t.mock = make(map[embedded.HardwareID]*HeaterMock)
	t.resp = httptest.NewRecorder()
}
func (t *HeaterTestSuite) TestHeater_PostHeaterAllGood_ReturnValuesFromInterface() {
	setHeater := embedded.HeaterConfig{
		HardwareID: "firstHeater",
		Enabled:    false,
		Power:      13,
	}
	returnHeater := embedded.HeaterConfig{
		HardwareID: setHeater.HardwareID,
		Enabled:    !setHeater.Enabled,
		Power:      uint(rand.Int() % 100),
	}

	heaterMock := new(HeaterMock)
	t.mock[embedded.HardwareID(setHeater.HardwareID)] = heaterMock

	heaterMock.On("Enable", setHeater.Enabled).Return(nil)
	heaterMock.On("SetPower", setHeater.Power).Return(nil)

	heaterMock.On("Enabled").Return(returnHeater.Enabled).Once()
	heaterMock.On("Power").Return(returnHeater.Power).Once()

	var body bytes.Buffer
	_ = json.NewEncoder(&body).Encode(setHeater)
	t.req, _ = http.NewRequest(http.MethodPut, "/api/config/heater/"+setHeater.HardwareID, &body)
	t.req.Header.Add("Content-Type", "application/json")

	h, _ := embedded.New(embedded.WithHeaters(t.heaters()))
	h.ServeHTTP(t.resp, t.req)
	b, _ := io.ReadAll(t.resp.Body)

	t.Equal(http.StatusOK, t.resp.Code)
	t.JSONEq(toJSON(returnHeater), string(b))
}

func (t *HeaterTestSuite) TestHeater_PostHeaterAllGoodTwice() {
	expectedHeater := embedded.HeaterConfig{
		HardwareID: "firstHeater",
		Enabled:    false,
		Power:      13,
	}

	heaterMock := new(HeaterMock)
	t.mock[embedded.HardwareID(expectedHeater.HardwareID)] = heaterMock
	{
		heaterMock.On("Enable", expectedHeater.Enabled).Return(nil).Once()
		heaterMock.On("SetPower", expectedHeater.Power).Return(nil).Once()

		heaterMock.On("Enabled").Return(expectedHeater.Enabled).Once()
		heaterMock.On("Power").Return(expectedHeater.Power).Once()

		var body bytes.Buffer
		_ = json.NewEncoder(&body).Encode(expectedHeater)
		t.req, _ = http.NewRequest(http.MethodPut, "/api/config/heater/"+expectedHeater.HardwareID, &body)
		t.req.Header.Add("Content-Type", "application/json")

		h, _ := embedded.New(embedded.WithHeaters(t.heaters()))
		h.ServeHTTP(t.resp, t.req)
		b, _ := io.ReadAll(t.resp.Body)

		t.Equal(http.StatusOK, t.resp.Code)
		t.JSONEq(toJSON(expectedHeater), string(b))
	}
	{
		newExpected := embedded.HeaterConfig{
			HardwareID: expectedHeater.HardwareID,
			Enabled:    !expectedHeater.Enabled,
			Power:      uint(rand.Uint64() % 100),
		}

		heaterMock.On("Enable", newExpected.Enabled).Return(nil).Once()
		heaterMock.On("SetPower", newExpected.Power).Return(nil).Once()
		heaterMock.On("Enabled").Return(newExpected.Enabled).Once()
		heaterMock.On("Power").Return(newExpected.Power).Once()

		var body bytes.Buffer
		_ = json.NewEncoder(&body).Encode(newExpected)
		t.req, _ = http.NewRequest(http.MethodPut, "/api/config/heater/"+newExpected.HardwareID, &body)
		t.req.Header.Add("Content-Type", "application/json")

		h, _ := embedded.New(embedded.WithHeaters(t.heaters()))
		h.ServeHTTP(t.resp, t.req)
		b, _ := io.ReadAll(t.resp.Body)

		t.Equal(http.StatusOK, t.resp.Code)
		t.JSONEq(toJSON(newExpected), string(b))
	}

}

func (t *HeaterTestSuite) TestHeater_PostHeaterInterfaceError() {
	args := []embedded.HeaterConfig{
		{
			HardwareID: "firstHeater",
			Enabled:    false,
			Power:      13,
		},
	}
	errOnSetPower := errors.New("nope, sorry")
	for _, arg := range args {
		heater := new(HeaterMock)
		heater.On("SetPower", mock.Anything).Return(errOnSetPower).Once()
		t.mock[embedded.HardwareID(arg.HardwareID)] = heater
	}

	var body bytes.Buffer
	_ = json.NewEncoder(&body).Encode(args[0])
	t.req, _ = http.NewRequest(http.MethodPut, "/api/config/heater/"+args[0].HardwareID, &body)
	t.req.Header.Add("Content-Type", "application/json")

	h, _ := embedded.New(embedded.WithHeaters(t.heaters()))
	h.ServeHTTP(t.resp, t.req)
	b, _ := io.ReadAll(t.resp.Body)

	t.Equal(http.StatusInternalServerError, t.resp.Code)
	t.Contains(string(b), errOnSetPower.Error())

}

func (t *HeaterTestSuite) TestHeater_PostHeaterDoesntExist() {
	t.req, _ = http.NewRequest(http.MethodPut, "/api/config/heater/blah", nil)
	t.req.Header.Add("Content-Type", "application/json")

	h, _ := embedded.New(embedded.WithHeaters(t.heaters()))
	h.ServeHTTP(t.resp, t.req)
	b, _ := io.ReadAll(t.resp.Body)

	t.Equal(http.StatusNotFound, t.resp.Code)
	t.Equal(toJSON(embedded.ErrHeaterDoesntExist), string(b))
}

func (t *HeaterTestSuite) TestHeater_GetHeater() {
	args := []embedded.HeaterConfig{
		{
			HardwareID: "firstHeater",
			Enabled:    false,
			Power:      13,
		},
		{
			HardwareID: "second",
			Enabled:    true,
			Power:      71,
		},
	}

	for _, arg := range args {
		heater := new(HeaterMock)
		heater.On("Enabled").Return(arg.Enabled).Once()
		heater.On("Power").Return(arg.Power).Once()
		t.mock[embedded.HardwareID(arg.HardwareID)] = heater
	}

	t.req, _ = http.NewRequest(http.MethodGet, embedded.RoutesGetHeaters, nil)

	h, _ := embedded.New(embedded.WithHeaters(t.heaters()))
	h.ServeHTTP(t.resp, t.req)

	b, _ := io.ReadAll(t.resp.Body)

	t.Equal(http.StatusOK, t.resp.Code)
	t.JSONEq(toJSON(args), string(b))
}

func (t *HeaterTestSuite) TestHeater_GetZeroHeaters() {

	t.req, _ = http.NewRequest(http.MethodGet, embedded.RoutesGetHeaters, nil)
	h, _ := embedded.New()
	h.ServeHTTP(t.resp, t.req)

	b, _ := io.ReadAll(t.resp.Body)

	t.Equal(http.StatusInternalServerError, t.resp.Code)
	t.JSONEq(toJSON(embedded.ErrNotImplemented), string(b))
}

func (t *HeaterTestSuite) TestHeater_MultipleHeaters() {
	type args_t struct {
		name    embedded.HardwareID
		power   uint
		enabled bool
	}
	args := []args_t{
		{
			name:    "firstHeater",
			power:   13,
			enabled: false,
		},
		{
			name:    "second",
			power:   71,
			enabled: true,
		},
		{
			name:    "third",
			power:   0,
			enabled: true,
		},
		{
			name:    "the last one",
			power:   91,
			enabled: false,
		},
	}

	for _, arg := range args {
		heater := new(HeaterMock)
		heater.On("Enable", arg.enabled).Once()
		heater.On("SetPower", arg.power).Return(nil).Once()
		heater.On("Enabled").Return(arg.enabled).Once()
		heater.On("Power").Return(arg.power).Once()
		t.mock[arg.name] = heater
	}

	handler, _ := embedded.New(embedded.WithHeaters(t.heaters()))
	for _, arg := range args {
		t.Nil(handler.HeaterEnable(arg.name, arg.enabled))
		t.Nil(handler.HeaterPower(arg.name, arg.power))
	}

	stat := handler.HeatersStatus()
	t.Len(stat, len(args))
	for _, elem := range stat {
		correct := false
		for _, arg := range args {
			if string(arg.name) == elem.HardwareID {
				correct = arg.enabled == elem.Enabled && arg.power == elem.Power
				if correct {
					break
				}
			}
		}
		t.True(correct, "expected elem not found")
	}

}

func (t *HeaterTestSuite) TestHeater_SingleHeater() {
	firstMock := new(HeaterMock)

	firstMock.On("Enable", true).Once()
	firstMock.On("SetPower", uint(16)).Return(nil).Once()
	firstMock.On("Enabled").Return(true).Once()
	firstMock.On("Power").Return(uint(16)).Once()

	t.mock["firstMock"] = firstMock

	handler, _ := embedded.New(embedded.WithHeaters(t.heaters()))

	err := handler.HeaterEnable("firstMock", true)
	t.Nil(err)
	err = handler.HeaterPower("firstMock", uint(16))
	t.Nil(err)

	stat := handler.HeatersStatus()
	t.Len(stat, 1)
	t.EqualValues(true, stat[0].Enabled)
	t.EqualValues(16, stat[0].Power)
	t.EqualValues("firstMock", stat[0].HardwareID)

	firstMock.AssertExpectations(t.T())
}

func (t *HeaterTestSuite) TestHeater_NoHeaters() {
	handler, err := embedded.New()
	t.Nil(err)
	t.NotNil(handler)

	err = handler.HeaterPower("b", 1)
	t.ErrorIs(err, embedded.ErrHeaterDoesntExist)

	err = handler.HeaterEnable("", true)
	t.ErrorIs(err, embedded.ErrHeaterDoesntExist)

	stat := handler.HeatersStatus()
	t.Len(stat, 0)
}

func (h *HeaterMock) Enable(ena bool) {
	h.Called(ena)
}

func (h *HeaterMock) SetPower(pwr uint) error {
	return h.Called(pwr).Error(0)
}

func (h *HeaterMock) Enabled() bool {
	return h.Called().Bool(0)
}

func (h *HeaterMock) Power() uint {
	args := h.Called()
	return args.Get(0).(uint)
}
