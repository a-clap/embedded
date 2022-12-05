package wifi

import (
	"github.com/theojulienne/go-wireless"
	"time"
)

const (
	WirelessAnyEvent = "another event"
)

var (
	_ Client   = (*wirelessClient)(nil)
	_ Wireless = (*wirelessWireless)(nil)

	eventMap = map[ID]string{
		NetworkNotFound: wireless.EventNetworkNotFound,
		Connected:       wireless.EventConnected,
		Disconnected:    wireless.EventDisconnected,
		AuthReject:      wireless.EventAuthReject,
		Any:             WirelessAnyEvent,
	}
)

type wirelessWireless struct {
}

type wirelessClient struct {
	quitCh chan struct{}
	*wireless.Client
}

func newWireless() *wirelessWireless {
	return &wirelessWireless{}
}

func (w *wirelessWireless) Interfaces(basePath ...string) []string {
	return wireless.Interfaces(basePath...)
}

func (w *wirelessWireless) Client(iface string) (Client, error) {
	c, err := wireless.NewClient(iface)
	if err != nil {
		return nil, err
	}
	return &wirelessClient{quitCh: nil, Client: c}, err
}

func (c *wirelessClient) Close() error {
	return c.Client.Conn().Close()
}

func (c *wirelessClient) SetScanTimeout(timeout time.Duration) {
	c.Client.ScanTimeout = timeout
}

func (c *wirelessClient) Status() (s Status, err error) {
	nets, err := c.Networks()
	if err != nil {
		return
	}

	if current, ok := nets.FindCurrent(); ok {
		s.Connected = true
		s.SSID = current.SSID
	}

	return
}

func (c *wirelessClient) Scan() ([]AP, error) {
	wAPs, err := c.Client.Scan()
	if err != nil {
		return nil, err
	}

	aps := make([]AP, len(wAPs))
	for i, ap := range wAPs {
		aps[i] = AP{
			ID:   ap.ID,
			SSID: ap.SSID,
		}
	}

	return aps, nil
}

func (c *wirelessClient) ConnectWithEvents(n Network, events ...ID) (<-chan Event, error) {
	var wirelessEvents []string
	handlingEvents := make(map[string]ID)
	for _, event := range events {
		if ev, ok := eventMap[event]; ok {
			wirelessEvents = append(wirelessEvents, ev)
			handlingEvents[ev] = event
		}
	}
	wirelessEvents = append(wirelessEvents, wireless.EventAssocReject)
	net := wireless.NewNetwork(n.SSID, n.Password)
	net, err := c.AddOrUpdateNetwork(net)
	if err != nil {
		return nil, err
	}
	sub := c.Conn().Subscribe(wirelessEvents...)
	if err := c.EnableNetwork(net.ID); err != nil {
		return nil, err
	}

	ch := c.handleEvents(sub, handlingEvents)
	return ch, nil
}

func (c *wirelessClient) handleEvents(s *wireless.Subscription, events map[string]ID) <-chan Event {
	// 99, because go-wireless internally creates channel like that
	evCh := make(chan Event, 99)
	passEvent := func(id ID, msg string) {
		evCh <- Event{
			ID:      id,
			Message: msg,
		}
	}

	go func() {
		run := true
		for run {
			select {
			case <-c.quitCh:
				run = false
			case ev := <-s.Next():
				// Check if event is in should-be-handled extEvents
				if id, ok := events[ev.Name]; ok {
					passEvent(id, ev.Name)
					continue
				}
				// Or handled by any event
				if id, ok := events[WirelessAnyEvent]; ok {
					passEvent(id, ev.Name)
					continue
				}
			}
		}
	}()

	return evCh
}

func (c *wirelessClient) Disconnect() error {
	_ = c.Conn().SendCommandBool(wireless.CmdDisconnect)
	return c.Conn().SendCommandBool(wireless.CmdFlush)
}
