package tracing

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/segmentio/encoding/json"
	"go.uber.org/fx"

	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
)

type KinesisHandlerParams struct {
	fx.In

	Logger   *otelzap.Logger
	PG       *bun.DB
	PS       *org.ProjectGateway
	Consumer *SpanConsumer
}

type KinesisHandler struct {
	*KinesisHandlerParams
}

func NewKinesisHandler(p KinesisHandlerParams) *KinesisHandler {
	return &KinesisHandler{&p}
}

func registerKinesisHandler(h *KinesisHandler, p bunapp.RouterParams) {
	p.Router.WithGroup("/api/v1/cloudwatch", func(g *bunrouter.Group) {
		g.POST("/logs", h.Logs)
	})
}

type KinesisEvent struct {
	RequestID string               `json:"requestId"`
	Records   []KinesisEventRecord `json:"records"`
}

type KinesisEventRecord struct {
	Data []byte `json:"data"`
}

type CloudwatchLog struct {
	MessageType         string               `json:"messageType"`
	Owner               string               `json:"owner"`
	LogGroup            string               `json:"logGroup"`
	LogStream           string               `json:"logStream"`
	SubscriptionFilters []string             `json:"subscriptionFilters"`
	LogEvents           []CloudwatchLogEvent `json:"logEvents"`
}

type CloudwatchLogEvent struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

func (h *KinesisHandler) Logs(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	dsn := req.Header.Get("X-Amz-Firehose-Access-Key")
	if dsn == "" {
		return errors.New("X-Amz-Firehose-Access-Key header is empty or missing")
	}

	project, err := h.PS.SelectProjectByDSN(ctx, dsn)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	event := new(KinesisEvent)
	if err := json.Unmarshal(body, &event); err != nil {
		return err
	}

	var log CloudwatchLog
	for _, record := range event.Records {
		rd, err := gzip.NewReader(bytes.NewReader(record.Data))
		if err != nil {
			return err
		}

		data, err := io.ReadAll(rd)
		if err != nil {
			return err
		}

		log = CloudwatchLog{}
		if err := json.Unmarshal(data, &log); err != nil {
			return err
		}

		if log.MessageType != "DATA_MESSAGE" {
			continue
		}

		for i := range log.LogEvents {
			event := &log.LogEvents[i]
			span := h.convEvent(event)
			span.ProjectID = project.ID
			h.Consumer.AddSpan(ctx, span)
		}
	}

	return httputil.JSON(w, bunrouter.H{
		"requestId": event.RequestID,
		"timestamp": time.Now().UnixMilli(),
	})
}

func (h *KinesisHandler) convEvent(event *CloudwatchLogEvent) *Span {
	span := new(Span)

	span.EventName = otelEventLog
	span.Kind = SpanKindInternal
	span.StatusCode = StatusCodeUnset
	span.Time = time.Unix(0, event.Timestamp*int64(time.Millisecond))
	span.Attrs = make(AttrMap, 1)
	if event.Message != "" {
		span.Attrs[attrkey.LogMessage] = event.Message
	}

	return span
}
