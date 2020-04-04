/*
 * NETCAP - Traffic Analysis Framework
 * Copyright (c) 2017 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package types

import (
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var fieldsPOP3 = []string{
	"Timestamp",
	"Client",    // string
	"Server",    // string
	"AuthToken", // string
	"User",      // string
	"Pass",      // string
	"NumRequests",  // []*POP3Request
	"NumResponses", // []*POP3Response
	"NumMails",     // []*Mail
}

func (a POP3) CSVHeader() []string {
	return filter(fieldsPOP3)
}

func (a POP3) CSVRecord() []string {
	return filter([]string{
		formatTimestamp(a.Timestamp),
		a.Client,    // string
		a.Server,    // string
		a.AuthToken, // string
		a.User,      // string
		a.Pass,      // string
		strconv.Itoa(len(a.Requests)),  // []*POP3Request
		strconv.Itoa(len(a.Responses)), // []*POP3Response
		strconv.Itoa(len(a.Mails)),     // []*Mail
	})
}

func (a POP3) Time() string {
	return a.Timestamp
}

func (a POP3) JSON() (string, error) {
	return jsonMarshaler.MarshalToString(&a)
}

var pop3Metric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: strings.ToLower(Type_NC_POP3.String()),
		Help: Type_NC_POP3.String() + " audit records",
	},
	fieldsPOP3[1:],
)

func init() {
	prometheus.MustRegister(pop3Metric)
}

func (a POP3) Inc() {
	pop3Metric.WithLabelValues(a.CSVRecord()[1:]...).Inc()
}

func (a *POP3) SetPacketContext(ctx *PacketContext) {}

func (a POP3) Src() string {
	return a.Client
}

func (a POP3) Dst() string {
	return a.Server
}
