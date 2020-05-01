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
	"github.com/dreadl0ck/gopacket"
	"github.com/dreadl0ck/gopacket/layers"
	"github.com/dreadl0ck/netcap/types"
	"github.com/golang/protobuf/proto"
)

var udpEncoder = CreateLayerEncoder(types.Type_NC_UDP, layers.LayerTypeUDP, func(layer gopacket.Layer, timestamp string) proto.Message {
	if udp, ok := layer.(*layers.UDP); ok {
		var payload []byte
		if c.IncludePayloads {
			payload = layer.LayerPayload()
		}
		var e float64
		if c.CalculateEntropy {
			e = Entropy(udp.Payload)
		}
		return &types.UDP{
			Timestamp:      timestamp,
			SrcPort:        int32(udp.SrcPort),
			DstPort:        int32(udp.DstPort),
			Length:         int32(udp.Length),
			Checksum:       int32(udp.Checksum),
			PayloadEntropy: e,
			PayloadSize:    int32(len(udp.Payload)),
			Payload:        payload,
		}
	}
	return nil
})
