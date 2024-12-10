package metrics

import (
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
	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
)

type KinesisHandlerParams struct {
	fx.In

	Logger *otelzap.Logger
	PG     *bun.DB
	MP     *DatapointProcessor
}

type KinesisHandler struct {
	*KinesisHandlerParams
}

func NewKinesisHandler(p KinesisHandlerParams) *KinesisHandler {
	return &KinesisHandler{&p}
}

func registerKinesisHandler(h *KinesisHandler, p bunapp.RouterParams) {
	p.Router.WithGroup("/api/v1/cloudwatch", func(g *bunrouter.Group) {
		g.POST("/metrics", h.Metrics)
	})
}

type KinesisEvent struct {
	RequestID string               `json:"requestId"`
	Records   []KinesisEventRecord `json:"records"`
}

type KinesisEventRecord struct {
	Data []byte `json:"data"`
}

type CloudwatchDatapoint struct {
	MetricStreamName string            `json:"metric_stream_name"`
	AccountID        string            `json:"account_id"`
	Region           string            `json:"region"`
	Namespace        string            `json:"namespace"`
	MetricName       string            `json:"metric_name"`
	Dimensions       map[string]string `json:"dimensions"`
	Timestamp        int64             `json:"timestamp"`
	Value            struct {
		Min   float64 `json:"min"`
		Max   float64 `json:"max"`
		Sum   float64 `json:"sum"`
		Count float64 `json:"count"`
	} `json:"value"`
	Unit string `json:"unit"`
}

func (h *KinesisHandler) Metrics(w http.ResponseWriter, req bunrouter.Request) error {
	ctx := req.Context()

	dsn := req.Header.Get("X-Amz-Firehose-Access-Key")
	if dsn == "" {
		return errors.New("X-Amz-Firehose-Access-Key header is empty or missing")
	}

	project, err := org.SelectProjectByDSN(ctx, h.PG, dsn)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	event := new(KinesisEvent)

	body, err = json.Parse(body, event, json.ZeroCopy)
	if err != nil {
		return err
	}

	p := otlpProcessor{
		logger:  h.Logger,
		pg:      h.PG,
		mp:      h.MP,
		project: project,
	}
	defer p.close(ctx)

	var src CloudwatchDatapoint
	for _, record := range event.Records {
		data := record.Data
		for len(data) > 2 {
			src = CloudwatchDatapoint{}

			var err error
			data, err = json.Parse(data, &src, json.ZeroCopy)
			if err != nil {
				if err == io.ErrUnexpectedEOF {
					break
				}
				return err
			}

			dest := new(Datapoint)

			if err := h.initDatapointFromAWS(project, dest, &src); err != nil {
				return err
			}

			p.enqueue(ctx, dest)
		}
	}

	return httputil.JSON(w, bunrouter.H{
		"requestId": event.RequestID,
		"timestamp": time.Now().UnixMilli(),
	})
}

func (h *KinesisHandler) initDatapointFromAWS(
	project *org.Project,
	dest *Datapoint,
	src *CloudwatchDatapoint,
) error {
	attrs := make(AttrMap, len(src.Dimensions)+4)
	for key, value := range src.Dimensions {
		key = attrkey.Underscore(key)
		if key == "" {
			continue
		}
		attrs[key] = value
	}
	attrs[attrkey.CloudProvider] = "aws"
	if src.AccountID != "" {
		attrs[attrkey.CloudAccountID] = src.AccountID
	}
	if src.Region != "" {
		attrs[attrkey.CloudRegion] = src.Region
	}
	if src.MetricStreamName != "" {
		attrs["metric_stream_name"] = src.MetricStreamName
	}

	dest.ProjectID = project.ID
	dest.Metric = attrkey.AWSMetricName(src.Namespace, src.MetricName)
	dest.Unit = bunconv.NormUnit(src.Unit)
	dest.Attrs = attrs

	dest.Time = time.Unix(0, src.Timestamp*int64(time.Millisecond))
	dest.Instrument = InstrumentSummary
	dest.Sum = src.Value.Sum
	dest.Count = uint64(src.Value.Count)
	dest.Min = src.Value.Min
	dest.Max = src.Value.Max

	return nil
}
