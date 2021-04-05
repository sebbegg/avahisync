package avahisync

import (
	"bytes"
	"context"
	"encoding/xml"
	"log"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strings"

	"github.com/grandcat/zeroconf"
)

type SyncConfig struct {
	Service      string
	Domain       string
	PortMapper   PortMapper
	OutputFolder string
	FilePrefix   string
	HostName     string
}

func Sync(config *SyncConfig) {
	// Discover all services on the network (e.g. _workstation._tcp)
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	// start listening for incoming entries
	entries := make(chan *zeroconf.ServiceEntry)
	go syncEntries(config, entries)

	// wait for SIGINT
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = resolver.Browse(ctx, config.Service, config.Domain, entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	select {
	case <-signals:
		log.Println("Received signal SIGINT - stopping.")
		cancel()
	case <-ctx.Done():
		log.Println("Received ctx.Done")
	}
	log.Println("Done")
}

type XmlServiceGroup struct {
	XMLName xml.Name        `xml:"service-group"`
	Name    string          `xml:"name"`
	Service XmlServiceEntry `xml:"service"`
}

type XmlServiceEntry struct {
	XMLName    xml.Name `xml:"service"`
	HostName   string   `xml:"host-name"`
	Type       string   `xml:"type"`
	Port       uint16   `xml:"port"`
	TxtRecords []string `xml:"txt-record"`
}

const SyncedMarkerText string = "from-avahi-sync=1"

func IsSyncedEntry(entry *zeroconf.ServiceEntry) bool {
	for _, record := range entry.Text {
		if record == SyncedMarkerText {
			return true
		}
	}
	return false
}

func syncEntries(config *SyncConfig, results <-chan *zeroconf.ServiceEntry) {
	for entry := range results {
		log.Println("######################################")
		log.Println(entry)

		if IsSyncedEntry(entry) {
			log.Printf("Skipping %s - already sync'd\n", entry.Instance)
			continue
		}
		entry.Text = append(entry.Text, SyncedMarkerText)
		xmlEntry, err := serviceEntryToXml(entry, config)

		if err != nil {
			log.Printf("ERROR: %s", err)
			return
		}

		log.Println(string(xmlEntry))

		fName := path.Join(config.OutputFolder, xmlName(entry, config))
		os.MkdirAll(config.OutputFolder, 0o0755)
		log.Println("Writing to: " + fName)

		err = os.WriteFile(fName, xmlEntry, 0o0666)
		if err != nil {
			log.Fatalf("ERROR: %s", err.Error())
		}
	}
	log.Println("No more entries.")
}

func xmlName(entry *zeroconf.ServiceEntry, config *SyncConfig) string {
	re := regexp.MustCompile(`[^\w0-9]+`)

	return config.FilePrefix + re.ReplaceAllString(entry.Instance, "_") + ".service"
}

const xmlHeader = `<?xml version="1.0" standalone='no'?>
<!DOCTYPE service-group SYSTEM "avahi-service.dtd">`

func serviceEntryToXml(entry *zeroconf.ServiceEntry, config *SyncConfig) ([]byte, error) {

	xmlEntry := XmlServiceEntry{
		HostName:   config.HostName,
		Type:       entry.Service,
		Port:       config.PortMapper.MapPort(entry.AddrIPv4[0], uint16(entry.Port)),
		TxtRecords: entry.Text,
	}

	XmlServiceGroup := &XmlServiceGroup{
		// fix escaped whitespace characters
		Name:    strings.ReplaceAll(entry.Instance, `\ `, " "),
		Service: xmlEntry,
	}

	xmlBytes, err := xml.MarshalIndent(XmlServiceGroup, "", "  ")
	return bytes.Join([][]byte{[]byte(xmlHeader), xmlBytes}, []byte("\n")), err
}
