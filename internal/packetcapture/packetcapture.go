package packetcapture

import (
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

const (
	captureDevice = "eth0"
)

// Handler is passed to AddHandler() to capture a specific gopacket.LayerType.
type Handler func(gopacket.Layer, gopacket.Packet)

type PacketCapture struct {
	inactiveHandle *pcap.InactiveHandle
	handle         *pcap.Handle
	packetSource   *gopacket.PacketSource
	packetHandlers map[gopacket.LayerType][]Handler
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
		inactiveHandle: inactiveHandle,
		handle:         nil,
		packetSource:   nil,
		packetHandlers: make(map[gopacket.LayerType][]Handler),
	}, nil
}

// AddHandler registers a specific Handler for a given gopacket LayerType.
//
// Each Handler will be called each time a packet with the given LayerType is
// received.
//
// Calls to Handler happen in a separate goroutine.
func (pc *PacketCapture) AddHandler(layerType gopacket.LayerType, handler Handler) {
	// TODO add a check to ensure the capture hasn't started
	if _, exists := pc.packetHandlers[layerType]; !exists {
		pc.packetHandlers[layerType] = make([]Handler, 0)
	}
	pc.packetHandlers[layerType] = append(pc.packetHandlers[layerType], handler)
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
	for t, handlers := range pc.packetHandlers {
		l := packet.Layer(t)
		if l == nil {
			continue
		}
		for _, h := range handlers {
			log.Println("Calling Handler")
			h(l, packet)
		}
	}
}
