/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package wifi_test

import (
	"errors"
	"testing"
	"time"

	"github.com/a-clap/iot/internal/wifi"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type WifiSuite struct {
	suite.Suite
}

type WirelessMock struct {
	mock.Mock
}

type ClientMock struct {
	mock.Mock
}

var (
	_            wifi.Wireless = (*WirelessMock)(nil)
	_            wifi.Client   = (*ClientMock)(nil)
	wirelessMock *WirelessMock
	clientMock   *ClientMock
)

func TestWifiSuite(t *testing.T) {
	suite.Run(t, new(WifiSuite))
}

func (t *WifiSuite) SetupTest() {
	wirelessMock = new(WirelessMock)
	clientMock = new(ClientMock)
}

func (t *WifiSuite) TestNewWithInterface_ErrorHandling() {
	wirelessMock.On("Interfaces", mock.Anything).Return([]string{}).Once()
	w, err := wifi.NewWithInterface(wirelessMock)
	t.Nil(w)
	t.ErrorIs(wifi.ErrLackOfInterface, err)

	returnErr := errors.New("broken")
	wirelessMock.On("Interfaces", mock.Anything).Return([]string{"interface 1"}).Once()
	wirelessMock.On("Client", "interface 1").Return(wifi.Client(nil), returnErr).Once()
	w, err = wifi.NewWithInterface(wirelessMock)
	t.Nil(w)
	t.ErrorIs(returnErr, err)

	returnErr2 := errors.New("broken2")
	wirelessMock.On("Interfaces", mock.Anything).Return([]string{"first", "second"}).Once()
	wirelessMock.On("Client", "first").Return(wifi.Client(nil), returnErr2).Once()
	wirelessMock.On("Client", "second").Return(wifi.Client(nil), returnErr2).Once()
	w, err = wifi.NewWithInterface(wirelessMock)
	t.Nil(w)
	t.ErrorIs(returnErr2, err)
}

func (t *WifiSuite) TestNewWithInterface_TryDifferentInterfaces() {
	returnErr2 := errors.New("broken2")
	wirelessMock.On("Interfaces", mock.Anything).Return([]string{"first", "second"}).Once()
	wirelessMock.On("Client", "first").Return(wifi.Client(nil), returnErr2).Once()
	wirelessMock.On("Client", "second").Return(wifi.Client(nil), returnErr2).Once()
	w, err := wifi.NewWithInterface(wirelessMock)
	t.Nil(w)
	t.ErrorIs(returnErr2, err)

	wirelessMock.AssertExpectations(t.T())
}

func (t *WifiSuite) TestChoose() {
	wirelessMock.On("Interfaces", mock.Anything).Return([]string{"first", "second"}).Once()
	wirelessMock.On("Client", "first").Return(wifi.Client(clientMock), nil).Once()
	clientMock.On("SetScanTimeout", mock.Anything).Return()
	w, err := wifi.NewWithInterface(wirelessMock)
	t.NotNil(w)
	t.Nil(err)
	// Correct choose
	clientMock.On("Close").Return(nil).Once()
	wirelessMock.On("Client", "second").Return(clientMock, nil).Once()
	err = w.Choose("second")
	t.Nil(err)

	// Wrong choose
	err = w.Choose("second2")
	t.NotNil(err)
	t.ErrorIs(wifi.ErrNoSuchInterface, err)

	wirelessMock.AssertExpectations(t.T())
	clientMock.AssertExpectations(t.T())
}

func (t *WifiSuite) TestConnect_Errors() {
	wirelessMock.On("Interfaces", mock.Anything).Return([]string{"first"}).Once()
	wirelessMock.On("Client", "first").Return(clientMock, nil).Once()
	clientMock.On("SetScanTimeout", mock.Anything).Return()
	w, _ := wifi.NewWithInterface(wirelessMock)
	t.NotNil(w)
	statusErr := errors.New("status err")
	{
		clientMock.On("Status").Return(wifi.Status{}, statusErr).Once()
		err := w.Connect(wifi.Network{})
		t.NotNil(err)
		t.ErrorIs(err, statusErr)
	}
	{
		status := wifi.Status{
			Connected: true,
			SSID:      "very interesting ssid",
		}
		clientMock.On("Status").Return(status, nil).Once()
		err := w.Connect(wifi.Network{})
		t.NotNil(err)
		t.ErrorIs(err, wifi.ErrAlreadyConnected)
		t.ErrorContains(err, status.SSID)
	}
	{
		status := wifi.Status{
			Connected: false,
		}
		errConnect := errors.New("err connect broken")
		clientMock.On("Status").Return(status, nil).Once()
		clientMock.On("ConnectWithEvents", mock.Anything, mock.Anything).Return(nil, errConnect).Once()
		err := w.Connect(wifi.Network{})
		t.NotNil(err)
		t.ErrorIs(err, errConnect)
	}
}
func (t *WifiSuite) TestDisconnect_Errors() {
	wirelessMock.On("Interfaces", mock.Anything).Return([]string{"first"}).Once()
	wirelessMock.On("Client", "first").Return(clientMock, nil).Once()
	clientMock.On("SetScanTimeout", mock.Anything).Return()
	w, _ := wifi.NewWithInterface(wirelessMock)
	{
		clientMock.On("Status").Return(wifi.Status{Connected: false}, nil).Once()
		err := w.Disconnect()
		t.NotNil(err)
		t.ErrorIs(err, wifi.ErrNotConnected)
	}
	{
		discErr := errors.New("broken")
		clientMock.On("Status").Return(wifi.Status{Connected: true}, nil).Once()
		clientMock.On("Disconnect").Return(discErr).Once()
		err := w.Disconnect()
		t.NotNil(err)
		t.ErrorIs(err, discErr)
	}
}

func (t *WifiSuite) TestScan() {
	wirelessMock.On("Interfaces", mock.Anything).Return([]string{"first"}).Once()
	wirelessMock.On("Client", "first").Return(clientMock, nil).Once()
	clientMock.On("SetScanTimeout", mock.Anything).Return()
	w, _ := wifi.NewWithInterface(wirelessMock)
	t.NotNil(w)

	{
		scanBroken := errors.New("scan broken")
		clientMock.On("Scan").Return(nil, scanBroken).Once()
		aps, err := w.APs()
		t.NotNil(err)
		t.ErrorIs(err, scanBroken)
		t.Nil(aps)
	}
	{
		retAps := []wifi.AP{
			{ID: 0, SSID: "ssid"},
			{ID: 13, SSID: "ssid2"},
		}
		clientMock.On("Scan").Return(retAps, nil).Once()
		aps, err := w.APs()
		t.Nil(err)
		t.Equal(retAps, aps)
	}
}

func (t *WifiSuite) TestConnect() {
	wirelessMock.On("Interfaces", mock.Anything).Return([]string{"first"}).Once()
	wirelessMock.On("Client", "first").Return(clientMock, nil).Once()
	clientMock.On("SetScanTimeout", mock.Anything).Return()
	w, _ := wifi.NewWithInterface(wirelessMock)
	t.NotNil(w)
	{
		clientMock.On("Status").Return(wifi.Status{Connected: false}, nil).Once()
		n := wifi.Network{
			AP: wifi.AP{
				ID:   1,
				SSID: "blah",
			},
			Password: "123",
		}
		// to speed up test
		w.SetCommandTimeout(0)
		clientMock.On("ConnectWithEvents", n, mock.Anything).Return(nil, nil).Once()
		err := w.Connect(n)
		t.NotNil(err)
		t.ErrorIs(err, wifi.ErrConnectionTimeout)
	}

	{
		clientMock.On("Status").Return(wifi.Status{Connected: false}, nil).Once()
		n := wifi.Network{
			AP: wifi.AP{
				ID:   1,
				SSID: "blah",
			},
			Password: "123",
		}
		// shouldn't play a role
		w.SetCommandTimeout(1 * time.Hour)
		// as Connect should return immediately
		ev := make(chan wifi.Event, 1)
		clientMock.On("ConnectWithEvents", n, mock.Anything).Return(ev, nil).Once()

		ev <- wifi.Event{
			ID:      wifi.AuthReject,
			Message: "reject",
		}

		var err error
		ch := make(chan struct{})
		go func() {
			err = w.Connect(n)
			close(ch)
		}()

		select {
		case <-ch:
		case <-time.After(100 * time.Millisecond):
			t.Fail("shouldn't be here")
		}

		t.NotNil(err)
		t.ErrorIs(err, wifi.ErrConnectionError)
	}
	{
		clientMock.On("Status").Return(wifi.Status{Connected: false}, nil).Once()
		n := wifi.Network{
			AP: wifi.AP{
				ID:   1,
				SSID: "blah",
			},
			Password: "123",
		}
		// shouldn't play a role
		w.SetCommandTimeout(1 * time.Hour)
		// as Connect should return immediately
		ev := make(chan wifi.Event, 1)
		clientMock.On("ConnectWithEvents", n, mock.Anything).Return(ev, nil).Once()

		ev <- wifi.Event{
			ID:      wifi.Connected,
			Message: "reject",
		}

		var err error
		ch := make(chan struct{})
		go func() {
			err = w.Connect(n)
			close(ch)
		}()

		select {
		case <-ch:
		case <-time.After(100 * time.Millisecond):
			t.Fail("shouldn't be here")
		}

		t.Nil(err)
	}

}

func (w *WirelessMock) Interfaces(basePath ...string) []string {
	args := w.Called(basePath)
	return args.Get(0).([]string)

}
func (w *WirelessMock) Client(iface string) (wifi.Client, error) {
	args := w.Called(iface)
	client, _ := args.Get(0).(wifi.Client)
	return client, args.Error(1)
}

func (c *ClientMock) Close() error {
	args := c.Called()
	return args.Error(0)
}

func (c *ClientMock) SetScanTimeout(duration time.Duration) {
	c.Called(duration)
}

func (c *ClientMock) Status() (wifi.Status, error) {
	args := c.Called()
	return args.Get(0).(wifi.Status), args.Error(1)
}

func (c *ClientMock) Scan() ([]wifi.AP, error) {
	args := c.Called()
	aps, _ := args.Get(0).([]wifi.AP)
	return aps, args.Error(1)
}

func (c *ClientMock) ConnectWithEvents(n wifi.Network, events ...wifi.EventID) (<-chan wifi.Event, error) {
	args := c.Called(n, events)
	ch, _ := args.Get(0).(chan wifi.Event)
	return ch, args.Error(1)
}

func (c *ClientMock) Disconnect() error {
	return c.Called().Error(0)
}
