package dest_test

import (
	"errors"
	"github.com/a-clap/iot/internal/dest"
	"github.com/a-clap/iot/internal/embedded"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type HeatersMock struct {
	m []embedded.HeaterConfig
	mock.Mock
}

func (h *HeatersMock) Discover() ([]embedded.HeaterConfig, error) {
	args := h.Called()
	h.m, _ = args.Get(0).([]embedded.HeaterConfig)
	return h.m, args.Error(1)
}

func (h *HeatersMock) Set(hardwareID string, heater embedded.HeaterConfig) error {
	args := h.Called(hardwareID, heater)
	for i, hr := range h.m {
		if hr.HardwareID == hardwareID {
			h.m[i] = heater
			break
		}
	}

	return args.Error(0)
}

type HeatersTestSuite struct {
	suite.Suite
	heatersMock *HeatersMock
}

func TestDestHeaters(t *testing.T) {
	suite.Run(t, new(HeatersTestSuite))
}

func (t *HeatersTestSuite) SetupTest() {
	t.heatersMock = new(HeatersMock)
}

func (t *HeatersTestSuite) TestNewHeaters_GlobalConfigSetHeater() {
	heaters := []embedded.HeaterConfig{
		{
			HardwareID: "1",
			Enabled:    true,
			Power:      15,
		},
		{
			HardwareID: "13",
			Enabled:    false,
			Power:      0,
		},
	}

	for _, h := range heaters {
		if !h.Enabled {
			continue
		}
		h.Enabled = false
		t.heatersMock.On("Set", h.HardwareID, h).Return(nil).Once()

	}
	t.heatersMock.On("Discover").Return(heaters, nil)
	h, _ := dest.New(dest.WithHeaters(t.heatersMock))
	t.NotNil(h)

	heaterHandler := h.Config().Heaters()
	ids, _ := heaterHandler.HardwareIDs()
	for i, id := range ids {
		heater, err := heaterHandler.Heater(id)
		t.Nil(err)
		info := heater.Config()

		info.Power = uint(i + 10)
		t.heatersMock.On("Set", info.HardwareID, info).Return(nil).Once()
		t.Nil(heater.Power(uint(i + 10)))

		info.Enabled = i%2 == 0
		t.heatersMock.On("Set", info.HardwareID, info).Return(nil).Once()
		t.Nil(heater.Enable(i%2 == 0))
	}

	// Check, if Set was reflected in models
	heaterHandler2 := h.Config().Heaters()
	ids, _ = heaterHandler2.HardwareIDs()
	for i, id := range ids {
		heater, _ := heaterHandler2.Heater(id)
		info := heater.Config()

		expectedInfo := embedded.HeaterConfig{
			HardwareID: id,
			Enabled:    i%2 == 0,
			Power:      uint(i + 10),
		}

		t.EqualValues(expectedInfo, info)

	}

}

func (t *HeatersTestSuite) TestNewHeaters_GlobalConfigInheritsHeaters() {
	heaters := []embedded.HeaterConfig{
		{

			HardwareID: "1",
			Enabled:    false,
			Power:      0,
		},
		{

			HardwareID: "2",
			Enabled:    true,
			Power:      15,
		},
	}

	t.heatersMock.On("Discover").Return(heaters, nil)
	for _, h := range heaters {
		h.Enabled = false
		t.heatersMock.On("Set", h.HardwareID, h).Return(nil)
	}
	h, _ := dest.New(dest.WithHeaters(t.heatersMock))
	t.NotNil(h)

	configHeaters := h.Config().Heaters()

	var names []embedded.HeaterConfig
	ids, _ := configHeaters.HardwareIDs()
	for _, id := range ids {
		heater, _ := configHeaters.Heater(id)
		names = append(names, heater.Config())
	}
	t.EqualValues(heaters, names)
}
func (t *HeatersTestSuite) TestNewHeaters_DisableEnabledHeaters() {
	heaters := []embedded.HeaterConfig{
		{

			HardwareID: "1",
			Enabled:    true,
			Power:      0,
		},
		{
			HardwareID: "2",
			Enabled:    false,
			Power:      15,
		},
	}

	t.heatersMock.On("Discover").Return(heaters, nil)
	for _, h := range heaters {
		if !h.Enabled {
			continue
		}
		h.Enabled = false
		t.heatersMock.On("Set", h.HardwareID, h).Return(nil)
	}
	h, _ := dest.New(dest.WithHeaters(t.heatersMock))
	t.NotNil(h)
}

func (t *HeatersTestSuite) TestNewHeaters_NoError() {
	t.heatersMock.On("Discover").Return(nil, nil)
	h, err := dest.New(dest.WithHeaters(t.heatersMock))
	t.Nil(err)
	t.NotNil(h)
}

func (t *HeatersTestSuite) TestNewHeaters_InitialError() {
	newErr := errors.New("newErr")
	t.heatersMock.On("Discover").Return(nil, newErr)
	h, err := dest.New(dest.WithHeaters(t.heatersMock))
	t.ErrorIs(err, newErr)
	t.Nil(h)
}
