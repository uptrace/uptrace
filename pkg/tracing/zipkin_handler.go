package tracing

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/segmentio/encoding/json"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/tracing/attrkey"
	"github.com/uptrace/uptrace/pkg/uuid"
)

type ZipkinHandler struct {
	*bunapp.App

	sp *SpanProcessor
}

func NewZipkinHandler(app *bunapp.App, sp *SpanProcessor) *ZipkinHandler {
	return &ZipkinHandler{
		App: app,
		sp:  sp,
	}
}

type ZipkinSpan struct {
	ID             string             `json:"id"`
	ParentID       string             `json:"parentId"`
	TraceID        string             `json:"traceId"`
	Name           string             `json:"name"`
	Timestamp      Int64OrString      `json:"timestamp"`
	Duration       Int64OrString      `json:"duration"`
	Kind           string             `json:"kind"`
	LocalEndpoint  ZipkinEndpoint     `json:"localEndpoint"`
	RemoteEndpoint ZipkinEndpoint     `json:"remoteEndpoint"`
	Tags           AttrMap            `json:"tags"`
	Annotations    []ZipkinAnnotation `json:"annotations"`
}

type ZipkinEndpoint struct {
	ServiceName string `json:"serviceName"`
	IPV4        string `json:"string"`
	IPV6        string `json:"string"`
	Port        int    `json:"port"`
}

type ZipkinAnnotation struct {
	Timestamp int64  `json:"timestamp"`
	Value     string `json:"value"`
}

func (h *ZipkinHandler) PostSpans(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()
	dec := json.NewDecoder(req.Body)

	var zipkinSpans []ZipkinSpan

	if err := dec.Decode(&zipkinSpans); err != nil {
		return err
	}

	spans := make([]Span, len(zipkinSpans))
	for i := range zipkinSpans {
		span := &spans[i]
		// TODO: accept project id
		span.ProjectID = 2

		zipkinSpan := &zipkinSpans[i]
		if err := initSpanFromZipkin(span, zipkinSpan); err != nil {
			return err
		}

		h.sp.AddSpan(ctx, span)
	}

	w.WriteHeader(http.StatusAccepted)
	return nil
}

// https://opentelemetry.io/docs/reference/specification/trace/sdk_exporters/zipkin/
func initSpanFromZipkin(dest *Span, src *ZipkinSpan) error {
	var err error

	dest.ID, err = parseZipkinID(src.ID)
	if err != nil {
		return err
	}

	if src.ParentID != "" {
		dest.ParentID, err = parseZipkinID(src.ParentID)
		if err != nil {
			return err
		}
	}

	dest.TraceID, err = uuid.Parse(src.TraceID)
	if err != nil {
		return err
	}

	dest.Name = src.Name
	dest.Kind = parseZipkinKind(src.Kind)
	dest.Time = time.Unix(0, int64(src.Timestamp)*1000)
	dest.Duration = time.Duration(src.Duration * 1000)
	dest.Attrs = src.Tags

	if dest.Attrs == nil {
		dest.Attrs = make(AttrMap)
	}
	if src.LocalEndpoint.ServiceName != "" {
		dest.Attrs.SetDefault(attrkey.ServiceName, src.LocalEndpoint.ServiceName)
	}
	if src.LocalEndpoint.IPV4 != "" {
		dest.Attrs.SetDefault(attrkey.NetHostIP, src.LocalEndpoint.IPV4)
	} else if src.LocalEndpoint.IPV6 != "" {
		dest.Attrs.SetDefault(attrkey.NetHostIP, src.LocalEndpoint.IPV6)
	}
	if src.LocalEndpoint.Port != 0 {
		dest.Attrs.SetDefault(attrkey.NetHostPort, src.LocalEndpoint.Port)
	}
	if src.RemoteEndpoint.ServiceName != "" {
		dest.Attrs.SetDefault(attrkey.PeerService, src.RemoteEndpoint.ServiceName)
	}
	if src.RemoteEndpoint.IPV4 != "" {
		dest.Attrs.SetDefault(attrkey.NetPeerIP, src.RemoteEndpoint.IPV4)
	} else if src.RemoteEndpoint.IPV6 != "" {
		dest.Attrs.SetDefault(attrkey.NetHostIP, src.RemoteEndpoint.IPV6)
	}
	if src.RemoteEndpoint.Port != 0 {
		dest.Attrs.SetDefault(attrkey.NetPeerPort, src.RemoteEndpoint.Port)
	}

	return nil
}

func parseZipkinID(s string) (uint64, error) {
	var buf [8]byte
	_, err := hex.Decode(buf[:], []byte(s))
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf[:]), nil
}

func parseZipkinKind(s string) string {
	switch s {
	case "SERVER":
		return ServerSpanKind
	case "CLIENT":
		return ClientSpanKind
	case "PRODUCER":
		return ProducerSpanKind
	case "CONSUMER":
		return ConsumerSpanKind
	default:
		return InternalSpanKind
	}
}

type Int64OrString int64

func (n *Int64OrString) UnmarshalJSON(b []byte) error {
	if n == nil {
		return errors.New("Int64OrString: UnmarshalJSON on nil pointer")
	}

	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	if len(b) == 0 {
		return fmt.Errorf("can't parse int64: %q", b)
	}

	num, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}

	*n = Int64OrString(num)

	return nil
}
