package packetcapture

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

const (
	captureDevice = "eth0"
)

// PacketReceiver implementations can be registered with a PacketCapture to be
// recieve all the packets for a specified set of LayerTypes.
type PacketReceiver interface {
	LayerTypes() []gopacket.LayerType
	Receive(gopacket.Layer, gopacket.Packet)
}

// Handler is passed to AddHandler() to capture a specific gopacket.LayerType.
type Handler func(gopacket.Layer, gopacket.Packet)

type PacketCapture struct {
	inactiveHandle  *pcap.InactiveHandle
	handle          *pcap.Handle
	packetSource    *gopacket.PacketSource
	packetReceivers map[gopacket.LayerType][]PacketReceiver
}

// New returns a new Trace instance.
//
// Close() must be called on the Trace instance.
func New() (*PacketCapture, error) {
	inactiveHandle, err := pcap.NewInactiveHandle(captureDevice)
	if err != nil {
		return nil, err
	}

	return &PacketCapture{
		inactiveHandle:  inactiveHandle,
		handle:          nil,
		packetSource:    nil,
		packetReceivers: make(map[gopacket.LayerType][]PacketReceiver),
	}, nil
}

// RegisterReceiver registers a receiver for a given set of gopacket LayerTypes.
//
// Each PacketReceiver will be called each time a packet in the set of
// LayerTypes is received.
//
// Calls to PacketReceiver.Receiver happen in a separate goroutine.
func (pc *PacketCapture) RegisterReceiver(receiver PacketReceiver) {
	// TODO add a check to ensure the capture hasn't started
	for _, lt := range receiver.LayerTypes() {
		if _, exists := pc.packetReceivers[lt]; !exists {
			pc.packetReceivers[lt] = make([]PacketReceiver, 0)
		}
		pc.packetReceivers[lt] = append(pc.packetReceivers[lt], receiver)
	}
}

func (pc *PacketCapture) Start() error {
	handle, err := pc.inactiveHandle.Activate()
	if err != nil {
		return err
	}
	pc.handle = handle
	pc.packetSource = gopacket.NewPacketSource(pc.handle, pc.handle.LinkType())
	go func() {
		for packet := range pc.packetSource.Packets() {
			pc.handlePacket(packet)
		}
	}()
	return nil
}

func (pc *PacketCapture) Close() {
	if pc.inactiveHandle != nil {
		pc.inactiveHandle.CleanUp()
	}
	if pc.handle != nil {
		pc.handle.Close()
	}
	pc.inactiveHandle = nil
	pc.handle = nil
	pc.packetSource = nil
}

func (pc *PacketCapture) handlePacket(packet gopacket.Packet) {
	for t, receivers := range pc.packetReceivers {
		l := packet.Layer(t)
		if l == nil {
			continue
		}
		for _, r := range receivers {
			r.Receive(l, packet)
		}
	}
}
