package rest_test

import (
	"bytes"
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

type WifiHandlerMock struct {
	mock.Mock
}

func (w *WifiHandlerMock) Status() (rest.WIFIStatus, error) {
	args := w.Called()
	return args.Get(0).(rest.WIFIStatus), args.Error(1)
}

func (w *WifiHandlerMock) Disconnect() error {
	args := w.Called()
	return args.Error(0)
}

var _ rest.WIFIHandler = (*WifiHandlerMock)(nil)

type WifiRestTestSuite struct {
	suite.Suite
	mocker *WifiHandlerMock
	srv    *rest.Server
	req    *http.Request
	resp   *httptest.ResponseRecorder
}

func (w *WifiHandlerMock) Connect(n rest.WIFINetwork) error {
	args := w.Called(n)
	return args.Error(0)

}

func (w *WifiHandlerMock) APs() ([]rest.WIFINetwork, error) {
	args := w.Called()
	aps, _ := args.Get(0).([]rest.WIFINetwork)
	return aps, args.Error(1)
}

func TestWifiSuite(t *testing.T) {
	suite.Run(t, new(WifiRestTestSuite))
}

func (t *WifiRestTestSuite) SetupTest() {
	t.mocker = new(WifiHandlerMock)
	t.srv, _ = rest.New(rest.WithFormat(rest.JSON), rest.WithWifiHandler(t.mocker))
	t.resp = httptest.NewRecorder()
}

func (t *WifiRestTestSuite) TestLackOfInterface() {
	t.srv, _ = rest.New(rest.WithFormat(rest.JSON))
	t.req, _ = http.NewRequest(http.MethodGet, rest.RoutesGetWifiAps, nil)

	t.srv.ServeHTTP(t.resp, t.req)

	body, _ := io.ReadAll(t.resp.Body)

	t.Equal(http.StatusInternalServerError, t.resp.Code)
	t.JSONEq(rest.ErrNotImplemented.JSON(), string(body))
}

func (t *WifiRestTestSuite) TestGetAps() {
	t.req, _ = http.NewRequest(http.MethodGet, rest.RoutesGetWifiAps, nil)
	t.srv, _ = rest.New(rest.WithFormat(rest.JSON), rest.WithWifiHandler(t.mocker))

	aps := []rest.WIFINetwork{{SSID: "hello"}}
	t.mocker.On("APs").Return(aps, nil).Once()

	t.srv.ServeHTTP(t.resp, t.req)
	body, _ := io.ReadAll(t.resp.Body)

	t.Equal(http.StatusOK, t.resp.Code)
	jAps, _ := json.Marshal(aps)
	t.JSONEq(string(jAps), string(body))
}

func (t *WifiRestTestSuite) TestHandleError() {
	t.req, _ = http.NewRequest(http.MethodGet, rest.RoutesGetWifiAps, nil)
	t.srv, _ = rest.New(rest.WithFormat(rest.JSON), rest.WithWifiHandler(t.mocker))

	err := errors.New("broken aps")
	t.mocker.On("APs").Return(nil, err).Once()

	t.srv.ServeHTTP(t.resp, t.req)
	body, _ := io.ReadAll(t.resp.Body)
	t.Equal(http.StatusInternalServerError, t.resp.Code)

	e := rest.Error{}
	_ = json.Unmarshal(body, &e)
	t.Equal(rest.ErrInterface.ErrorCode, e.ErrorCode)
	t.ErrorContains(e, err.Error())
}

func (t *WifiRestTestSuite) TestConnectToAp() {

	t.srv, _ = rest.New(rest.WithFormat(rest.JSON), rest.WithWifiHandler(t.mocker))
	ap := rest.WIFINetwork{
		SSID:     "blah",
		Password: "123",
	}
	t.mocker.On("Connect", ap).Return(nil)
	var b bytes.Buffer
	_ = json.NewEncoder(&b).Encode(ap)

	t.req, _ = http.NewRequest(http.MethodPost, rest.RoutesConnectWifi, &b)
	t.req.Header.Add("Content-Type", "application/json")

	t.srv.ServeHTTP(t.resp, t.req)
	_, _ = io.ReadAll(t.resp.Body)
	t.Equal(http.StatusOK, t.resp.Code)
}
