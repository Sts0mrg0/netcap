/*
 * NETCAP - Traffic Analysis Framework
 * Copyright (c) 2017-2020 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package encoder

import (
	"github.com/dreadl0ck/gopacket/layers"
	"github.com/dreadl0ck/ja3"
	"github.com/dreadl0ck/netcap/resolvers"
	"github.com/dreadl0ck/netcap/utils"
	"github.com/evilsocket/islazy/tui"
	"log"
	"strconv"
	"strings"
	"sync/atomic"

	"sync"

	"github.com/dreadl0ck/gopacket"
	"github.com/dreadl0ck/netcap/dpi"
	"github.com/dreadl0ck/netcap/types"
	"github.com/golang/protobuf/proto"
)

var products = []string{
	"Windows NT",
	"Win64",
	"Trident",
	"Firefox",
	"Chrome",
	"Safari",
	"Apache",
	"nginx",
	"AmazonS3",
	"PHP",
	"Java",
	"Microsoft-IIS",
	"Netscape-Enterprise",
	"Syntactic",
	"Squid",
	"Python-urllib",
	"Edge",
	"Opera",
	"Firebird",
	"Iceweasel",
	"lighttpd",
	"Apache-Coyote",
	"Sun-ONE-Web-Server",
	"OracleAS-Web-Cache-10g",
	"Sun-Java-System-Web-Server",
	// Mozilla ? e.g: Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; WOW64; Trident/6.0)
}

var vendors = []string{
	"Apple",
	"Microsoft",
	"Cisco",
	"Mozilla",
	"BlackBerry",
}

var operatingSystems = []string{
	"ubuntu",
	"macOS",
	"linux",
	"windows",
	"android",
	"ios",
}

type Software struct {
	*types.Software
	sync.Mutex
}

// AtomicDeviceProfileMap contains all connections and provides synchronized access
type AtomicSoftwareMap struct {
	// map Product Name + "Version" to Software?
	Items map[string]*Software
	sync.Mutex
}

// Size returns the number of elements in the Items map
func (a *AtomicSoftwareMap) Size() int {
	a.Lock()
	defer a.Unlock()
	return len(a.Items)
}

var (
	// SoftwareStore hold all connections
	SoftwareStore = &AtomicSoftwareMap{
		Items: make(map[string]*Software),
	}

	// TODO: create a central stats source for this and other metrics: encoderStats
	numSoftware int64

	//parser, errInitUAParser = uaparser.New("./regexes.yaml")
	//pMu sync.Mutex
)

func findVendor(in string) string {
	for _, v := range vendors {
		if strings.Contains(in, v) {
			return v
		}
	}
	return ""
}

// e.g: XXX Firefox/12.0 YYY -> [ "XXX Firefox" "/12.0 YYY" ] -> 12.0
// e.g: XXX Windows NT 6.1 YYY -> [ "XXX Windows NT" " 6.1 YYY" ] -> 6.1
func findVersion(in string, product string) string {
	parts := strings.Split(in, product)
	if len(parts) > 1 {
		if strings.HasPrefix(parts[1], "/") {
			return strings.TrimSuffix(
				strings.TrimSuffix(
					strings.Fields(
						strings.TrimPrefix(parts[1], "/"),
					)[0],
					";"),
				"|")
		}
		if strings.HasPrefix(parts[1], " ") {
			return strings.TrimSuffix(
				strings.TrimSuffix(
					strings.Fields(
						strings.TrimPrefix(parts[1], " "),
					)[0],
					";"),
				"|")
		}
	}
	return ""
}

func whatSoftware(dp *DeviceProfile, i *packetInfo, f, serviceNameSrc, serviceNameDst, ja3Result, userAgents, serverNames string, protos []string) (software []*Software) {

	//fmt.Println(serviceNameSrc, serviceNameDst, manufacturer, ja3Result, userAgents, serverNames, protos)

	var service string
	if serviceNameSrc != "" {
		service = serviceNameSrc
	}
	if serviceNameDst != "" {
		service = serviceNameDst
	}

	// test ua parser pkg
	// TODO: implement a caching layer, to ensure each UA is only parsed once + add regexes to db directory
	//for _, ua := range strings.Split(userAgents, "| ") {
	//	pMu.Lock()
	//	client := parser.Parse(ua)
	//	pMu.Unlock()
	//	fmt.Println("UserAgent.Family:", client.UserAgent.Family)  // "Amazon Silk"
	//	fmt.Println("UserAgent.Major:", client.UserAgent.Major)   // "1"
	//	fmt.Println("UserAgent.Minor:", client.UserAgent.Minor)   // "1"
	//	fmt.Println("UserAgent.Patch:", client.UserAgent.Patch)   // "0-80"
	//	fmt.Println("Os.Family:", client.Os.Family)         // "Android"
	//	fmt.Println("Os.Major:", client.Os.Major)          // ""
	//	fmt.Println("Os.Minor:", client.Os.Minor)          // ""
	//	fmt.Println("Os.Patch:", client.Os.Patch)          // ""
	//	fmt.Println("Os.PatchMinor:", client.Os.PatchMinor)     // ""
	//	fmt.Println("Device.Family:", client.Device.Family)     // "Kindle Fire"
	//}

	var s []*Software
	for _, p := range products {
		if strings.Contains(userAgents, p) {
			s = append(s, &Software{
				Software: &types.Software{
					Timestamp:      i.timestamp,
					Product:        p,
					Vendor:         findVendor(userAgents),
					Version:        findVersion(userAgents, p),
					DeviceProfiles: []string{dp.MacAddr + "-" + dp.DeviceManufacturer},
					Source:         "userAgents: " + userAgents,
					Service:        service,
					DPIResults:     protos,
					Flow:           f,
				},
			})
		}
		if strings.Contains(serverNames, p) {
			s = append(s, &Software{
				Software: &types.Software{
					Timestamp:      i.timestamp,
					Product:        p,
					Vendor:         findVendor(serverNames),
					Version:        findVersion(serverNames, p),
					Source:         "serverNames: " + serverNames,
					DeviceProfiles: []string{dp.MacAddr + "-" + dp.DeviceManufacturer},
					Service:        service,
					DPIResults:     protos,
					Flow:           f,
				},
			})
		}
		if ja3Result != "" {
			if strings.Contains(ja3Result, p) {
				s = append(s, &Software{
					Software: &types.Software{
						Timestamp:      i.timestamp,
						Product:        p,
						Vendor:         findVendor(ja3Result),
						Version:        findVersion(ja3Result, p),
						DeviceProfiles: []string{dp.MacAddr + "-" + dp.DeviceManufacturer},
						Source:         "ja3Result: " + ja3Result,
						Service:        service,
						DPIResults:     protos,
						Flow:           f,
					},
				})
			}
		}
	}

	return s
}

// AnalyzeSoftware tries to identify software based on observations from the data
func AnalyzeSoftware(i *packetInfo) {

	var (
		serviceNameSrc, serviceNameDst string
		ja3Hash                        = ja3.DigestHexPacket(i.p)
		ja3Result                      string
		protos                         []string
		userAgents, serverNames        string
		f                              string
	)
	if ja3Hash == "" {
		ja3Hash = ja3.DigestHexPacketJa3s(i.p)
	}

	// Lookup Service For Port Numbers
	if tl := i.p.TransportLayer(); tl != nil {

		// set flow ident
		f = i.srcIP + ":" + tl.TransportFlow().Src().String() + "->" + i.dstIP + ":" + tl.TransportFlow().Dst().String()

		// get source port and convert to integer
		src, err := strconv.Atoi(tl.TransportFlow().Src().String())
		if err == nil {
			switch tl.LayerType() {
			case layers.LayerTypeTCP:
				serviceNameSrc = resolvers.LookupServiceByPort(src, "tcp")
			case layers.LayerTypeUDP:
				serviceNameSrc = resolvers.LookupServiceByPort(src, "udp")
			default:
			}
		}
		dst, err := strconv.Atoi(tl.TransportFlow().Dst().String())
		if err == nil {
			switch tl.LayerType() {
			case layers.LayerTypeTCP:
				serviceNameDst = resolvers.LookupServiceByPort(dst, "tcp")
			case layers.LayerTypeUDP:
				serviceNameDst = resolvers.LookupServiceByPort(dst, "udp")
			default:
			}
		}
	} else {

		// no transport layer
		f = i.srcIP + "->" + i.dstIP
	}

	// Deep Packet Inspection
	results := dpi.GetProtocols(i.p)
	for p := range results {
		protos = append(protos, p)
	}

	// Check available HTTP meta infos
	httpStore.Lock()
	if uas, ok := httpStore.UserAgents[i.srcIP]; ok {
		userAgents = uas
	}
	if sns, ok := httpStore.ServerNames[i.dstIP]; ok {
		serverNames = sns
	}
	httpStore.Unlock()

	// TLS fingerprinting
	if ja3Hash != "" {
		ja3Result = resolvers.LookupJa3(ja3Hash)
	}

	dp := getDeviceProfile(i.srcMAC, i)
	software := whatSoftware(dp, i, f, serviceNameSrc, serviceNameDst, ja3Result, userAgents, serverNames, protos)
	if len(software) == 0 {
		return
	}

	// lookup profile
	SoftwareStore.Lock()
	for _, s := range software {
		if p, ok := SoftwareStore.Items[s.Product+"/"+s.Version]; ok {
			updateSoftwareAuditRecord(dp, p, i)
		} else {
			SoftwareStore.Items[s.Product+"/"+s.Version] = s
			numSoftware++
		}
	}
	SoftwareStore.Unlock()
}

// NewDeviceProfile creates a new device specific profile
func NewSoftware(i *packetInfo) *Software {
	return &Software{
		Software: &types.Software{
			Timestamp: i.timestamp,
		},
	}
}

func updateSoftwareAuditRecord(dp *DeviceProfile, p *Software, i *packetInfo) {

	ident := dp.MacAddr + "-" + dp.DeviceManufacturer

	p.Lock()
	for _, pr := range p.DeviceProfiles {
		if pr == ident {
			p.Unlock()
			return
		}
	}
	p.DeviceProfiles = append(p.DeviceProfiles, ident)
	p.Unlock()
}

var softwareEncoder = CreateCustomEncoder(types.Type_NC_SOFTWARE, "Software", func(d *CustomEncoder) error {
	return nil
}, func(p gopacket.Packet) proto.Message {

	// handle packet
	AnalyzeSoftware(newPacketInfo(p))

	return nil
}, func(e *CustomEncoder) error {

	httpStore.Lock()
	var rows [][]string
	for ip, ua := range httpStore.UserAgents {
		rows = append(rows, []string{ip, ua})
	}
	tui.Table(utils.DebugLogFileHandle, []string{"IP", "UserAgents"}, rows)
	rows = [][]string{}
	for ip, sn := range httpStore.ServerNames {
		rows = append(rows, []string{ip, sn})
	}
	tui.Table(utils.DebugLogFileHandle, []string{"IP", "ServerNames"}, rows)
	httpStore.Unlock()

	// teardown DPI C libs
	dpi.Destroy()

	// flush writer
	if !e.writer.IsChanWriter {
		for _, c := range SoftwareStore.Items {
			c.Lock()
			e.write(c.Software)
			c.Unlock()
		}
	}
	return nil
})

// TODO: move into CustomEncoder and use in other places
// remove unnecessary global state
// writeProfile writes the profile
func (e *CustomEncoder) write(c types.AuditRecord) {

	if e.export {
		c.Inc()
	}

	atomic.AddInt64(&e.numRecords, 1)
	err := e.writer.Write(c.(proto.Message))
	if err != nil {
		log.Fatal("failed to write proto: ", err)
	}
}
