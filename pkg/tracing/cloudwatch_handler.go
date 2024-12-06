package tracing

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/bun"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
)

type KinesisHandler struct {
	logger   *otelzap.Logger
	pg       *bun.DB
	consumer *SpanConsumer
}

func NewKinesisHandler(logger *otelzap.Logger, pg *bun.DB, consumer *SpanConsumer) *KinesisHandler {
	return &KinesisHandler{
		logger:   logger,
		pg:       pg,
		consumer: consumer,
	}
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

	fakeApp := &bunapp.App{PG: h.pg}
	project, err := org.SelectProjectByDSN(ctx, fakeApp, dsn)
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
			h.consumer.AddSpan(ctx, span)
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
