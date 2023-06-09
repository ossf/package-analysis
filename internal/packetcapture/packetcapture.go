package packetcapture

import (
	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcapgo"
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
	netInterface    string
	handle          *pcapgo.EthernetHandle
	done            chan bool
	stop            chan bool
	packetReceivers map[gopacket.LayerType][]PacketReceiver
}

// New returns a new PacketCapture instance for the given netInterface
//
// Close() must be called on the PacketCapture instance.
func New(netInterface string) *PacketCapture {
	return &PacketCapture{
		netInterface:    netInterface,
		handle:          nil,
		done:            make(chan bool),
		stop:            make(chan bool),
		packetReceivers: make(map[gopacket.LayerType][]PacketReceiver),
	}
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
	// Use the pcapgo library for capturing traffic as it is the most reliable.
	// afpacket is the fastest but segfaults: https://github.com/google/gopacket/issues/717
	// pcap will block during Cancel(): https://github.com/google/gopacket/issues/890
	handle, err := pcapgo.NewEthernetHandle(pc.netInterface)
	if err != nil {
		return err
	}
	pc.handle = handle
	packetSource := gopacket.NewPacketSource(pc.handle, layers.LinkTypeEthernet)
	go func(packets chan gopacket.Packet, done chan bool, stop chan bool) {
		defer func() { done <- true }()
		for {
			select {
			case packet, ok := <-packets:
				if ok {
					pc.handlePacket(packet)
				} else {
					return
				}
			case v, ok := <-stop:
				if !ok || v {
					return
				}
			}
		}
	}(packetSource.Packets(), pc.done, pc.stop)
	return nil
}

func (pc *PacketCapture) Close() {
	if pc.handle != nil {
		pc.stop <- true
		pc.handle.Close()
		// Wait for packet processing to be finished.
		<-pc.done
	}
	pc.handle = nil
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
