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
	"go.uber.org/fx"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/org"
)

type ZipkinHandlerParams struct {
	fx.In

	Logger   *otelzap.Logger
	PG       *bun.DB
	Projects *org.ProjectGateway
	Consumer *SpanConsumer
}

type ZipkinHandler struct {
	*ZipkinHandlerParams
}

func NewZipkinHandler(p ZipkinHandlerParams) *ZipkinHandler {
	return &ZipkinHandler{&p}
}

func registerZipkinHandler(h *ZipkinHandler, p bunapp.RouterParams) {
	// https://zipkin.io/zipkin-api/#/default/post_spans
	p.Router.WithGroup("/api/v2", func(g *bunrouter.Group) {
		g.POST("/spans", h.PostSpans)
	})
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
	IPV4        string `json:"ipv4"`
	IPV6        string `json:"ipv6"`
	Port        int    `json:"port"`
}

type ZipkinAnnotation struct {
	Timestamp int64  `json:"timestamp"`
	Value     string `json:"value"`
}

func (h *ZipkinHandler) PostSpans(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	dsn, err := org.DSNFromRequest(req)
	if err != nil {
		return err
	}

	project, err := h.Projects.SelectByDSN(ctx, dsn)
	if err != nil {
		return err
	}

	var zipkinSpans []ZipkinSpan

	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&zipkinSpans); err != nil {
		return err
	}

	spans := make([]Span, len(zipkinSpans))
	for i := range zipkinSpans {
		span := &spans[i]
		span.ProjectID = project.ID

		zipkinSpan := &zipkinSpans[i]
		if err := initSpanFromZipkin(span, zipkinSpan); err != nil {
			return err
		}

		h.Consumer.AddSpan(ctx, span)
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

	dest.TraceID, err = idgen.ParseTraceID(src.TraceID)
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
		dest.Attrs.SetDefault(attrkey.ClientAddress, src.LocalEndpoint.IPV4)
	} else if src.LocalEndpoint.IPV6 != "" {
		dest.Attrs.SetDefault(attrkey.ClientAddress, src.LocalEndpoint.IPV6)
	}
	if src.LocalEndpoint.Port != 0 {
		dest.Attrs.SetDefault(attrkey.ClientPort, src.LocalEndpoint.Port)
	}
	if src.RemoteEndpoint.ServiceName != "" {
		dest.Attrs.SetDefault(attrkey.PeerService, src.RemoteEndpoint.ServiceName)
	}
	if src.RemoteEndpoint.IPV4 != "" {
		dest.Attrs.SetDefault(attrkey.ServerAddress, src.RemoteEndpoint.IPV4)
	} else if src.RemoteEndpoint.IPV6 != "" {
		dest.Attrs.SetDefault(attrkey.ServerAddress, src.RemoteEndpoint.IPV6)
	}
	if src.RemoteEndpoint.Port != 0 {
		dest.Attrs.SetDefault(attrkey.ServerPort, src.RemoteEndpoint.Port)
	}

	return nil
}

func parseZipkinID(s string) (idgen.SpanID, error) {
	var buf [8]byte
	_, err := hex.Decode(buf[:], []byte(s))
	if err != nil {
		return 0, err
	}
	return idgen.SpanID(binary.LittleEndian.Uint64(buf[:])), nil
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
