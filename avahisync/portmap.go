package avahisync

import (
	"log"
	"net"
)

type PortMapper interface {
	MapPort(ip net.IP, port uint16) uint16
}

type StaticPortMap map[uint16]uint16

type StaticPortMapper struct {
	PortMap StaticPortMap
}

func (m *StaticPortMapper) MapPort(ip net.IP, port uint16) uint16 {
	outPort, found := m.PortMap[port]

	if found {
		log.Printf("Mapping %s:%d -> %d", ip.String(), port, outPort)
		return outPort
	} else {
		return port
	}
}
