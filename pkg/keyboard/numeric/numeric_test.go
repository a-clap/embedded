package numeric

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type KeyboardSuite struct {
	suite.Suite
}

type ValuerMock struct {
	mock.Mock
}

var (
	_      Value = &ValuerMock{}
	mocker *ValuerMock
	app    fyne.App
)

func TestKeyboardSuite(t *testing.T) {
	suite.Run(t, new(KeyboardSuite))
}

func (ks *KeyboardSuite) SetupTest() {
	mocker = new(ValuerMock)
	app = test.NewApp()
}

func (ks *KeyboardSuite) TearDownSuite() {
	app.Quit()
}

func (ks *KeyboardSuite) TestInput() {
	tests := []struct {
		name      string
		getReturn string
		setArg    string
		sequence  []button
	}{
		{
			name:      "basic",
			getReturn: "12",
			setArg:    "12123",
			sequence:  []button{'1', '2', '3', enter},
		},
		{
			name:      "clear, then input",
			getReturn: "12",
			setArg:    "456",
			sequence:  []button{clr, '4', '5', '6', enter},
		},
		{
			name:      "float, then input",
			getReturn: "1.2",
			setArg:    "1.23",
			sequence:  []button{'3', enter},
		},
		{
			name:      "float: clear, then input with dot",
			getReturn: "7.6",
			setArg:    "567.1",
			sequence:  []button{clr, '5', '6', '7', dot, '1', enter},
		},
		{
			name:      "float: clear, then input without dot",
			getReturn: "7.6",
			setArg:    "567.0",
			sequence:  []button{clr, '5', '6', '7', enter},
		},
		{
			name:      "float: multiple dots",
			getReturn: "13.6",
			setArg:    "981.3",
			sequence:  []button{clr, '9', '8', '1', dot, '3', dot, enter},
		},
		{
			name:      "int: minus toggles impl",
			getReturn: "631",
			setArg:    "-631",
			sequence:  []button{minus, enter},
		},
		{
			name:      "int: minus toggles impl twice",
			getReturn: "631",
			setArg:    "631",
			sequence:  []button{minus, minus, enter},
		},
		{
			name:      "float: minus toggles impl",
			getReturn: "631.0",
			setArg:    "-631.0",
			sequence:  []button{minus, enter},
		},
		{
			name:      "float: minus toggles impl twice",
			getReturn: "631.0",
			setArg:    "631.0",
			sequence:  []button{minus, minus, enter},
		},
		{
			name:      "int clr",
			getReturn: "12",
			setArg:    "0",
			sequence:  []button{clr, enter},
		},
		{
			name:      "float clr",
			getReturn: "12.0",
			setArg:    "0.0",
			sequence:  []button{clr, enter},
		},
		{
			name:      "int single bs",
			getReturn: "12",
			setArg:    "1",
			sequence:  []button{bs, enter},
		},
		{
			name:      "float single bs",
			getReturn: "123.1",
			setArg:    "123.0",
			sequence:  []button{bs, enter},
		},
		{
			name:      "int multiple bs",
			getReturn: "91",
			setArg:    "0",
			sequence:  []button{bs, bs, enter},
		},
		{
			name:      "float multiple bs",
			getReturn: "9.0",
			setArg:    "0.0",
			sequence:  []button{bs, bs, bs, enter},
		},
		{
			name:      "float: allow to write 0.0x",
			getReturn: "3.5",
			setArg:    "0.01",
			sequence:  []button{clr, '0', dot, '0', '1', enter},
		},
	}
	for _, args := range tests {
		mocker = new(ValuerMock)
		mocker.On("Get").Return(args.getReturn)
		mocker.On("Set", args.setArg)
		w, _ := Show(mocker)
		ks.NotNil(w)
		for _, button := range args.sequence {
			test.Tap(numericKeyboard.buttons[button])
		}
		if mocker.AssertExpectations(ks.T()) == false {
			ks.Failf("expectations not fulfilled", "test name: %s\n", args.name)
		}
	}
}

func (ks *KeyboardSuite) TestEsc() {
	mocker = new(ValuerMock)
	mocker.On("Get").Return("123")
	w, _ := Show(mocker)
	called := false
	w.SetOnClosed(func() {
		called = true
	})
	ks.NotNil(w)

	test.Tap(numericKeyboard.buttons['1'])
	test.Tap(numericKeyboard.buttons['6'])
	test.Tap(numericKeyboard.buttons['7'])
	test.Tap(numericKeyboard.buttons[esc])

	if mocker.AssertExpectations(ks.T()) == false {
		ks.Failf("expectations not fulfilled", "TestEsc")
	}
	ks.True(called, "should be closed on esc")
}

func (v *ValuerMock) Set(val string) {
	_ = v.Called(val)
}

func (v *ValuerMock) Get() string {
	args := v.Called()
	return args.Get(0).(string)
}
