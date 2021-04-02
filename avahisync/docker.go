package avahisync

import (
	"context"
	"log"
	"net"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type DockerPortMapper struct {
	client *client.Client
}

func NewDockerPortMapper() (PortMapper, error) {
	cl, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err == nil {
		return &DockerPortMapper{
			client: cl,
		}, nil
	}
	return nil, err
}

func (m *DockerPortMapper) MapPort(ip net.IP, port uint16) uint16 {
	ctx := context.Background()
	containers, err := m.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		log.Printf("Cannot list containers: %v", err)
		return port
	}

	for _, c := range containers {
		for name, network := range c.NetworkSettings.Networks {
			log.Printf("Found %s with ip %s in network %s",
				c.Names[0], network.IPAddress, name)

			if network.IPAddress == ip.String() {
				for _, cPort := range c.Ports {
					if uint16(cPort.PrivatePort) == port {
						log.Printf("Found %s with private port %d and public port %d",
							c.Names[0], cPort.PrivatePort, cPort.PublicPort)
						return uint16(cPort.PublicPort)
					}
				}

				break
			}
		}
	}

	return port
}
