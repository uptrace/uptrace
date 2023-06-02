package metrics

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/segmentio/encoding/json"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bununit"
	"github.com/uptrace/uptrace/pkg/httputil"
	"github.com/uptrace/uptrace/pkg/org"
)

type KinesisHandler struct {
	*bunapp.App

	mp *MeasureProcessor
}

func NewKinesisHandler(app *bunapp.App, mp *MeasureProcessor) *KinesisHandler {
	return &KinesisHandler{
		App: app,
		mp:  mp,
	}
}

type KinesisEvent struct {
	RequestID string               `json:"requestId"`
	Records   []KinesisEventRecord `json:"records"`
}

type KinesisEventRecord struct {
	Data []byte `json:"data"`
}

type CloudwatchMeasure struct {
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

	project, err := org.SelectProjectByDSN(ctx, h.App, dsn)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	event := new(KinesisEvent)

	body, err = json.Parse(body, event, json.ZeroCopy)
	if err != nil {
		return err
	}

	p := otlpProcessor{
		App: h.App,

		mp: h.mp,

		ctx:     ctx,
		project: project,
	}

	var src CloudwatchMeasure
	for _, record := range event.Records {
		data := record.Data
		for len(data) > 2 {
			src = CloudwatchMeasure{}

			var err error
			data, err = json.Parse(data, &src, json.ZeroCopy)
			if err != nil {
				if err == io.ErrUnexpectedEOF {
					break
				}
				return err
			}

			dest := new(Measure)

			if err := h.initMeasureFromAWS(project, dest, &src); err != nil {
				return err
			}

			p.enqueue(dest)
		}
	}

	return httputil.JSON(w, bunrouter.H{
		"requestId": event.RequestID,
		"timestamp": time.Now().UnixMilli(),
	})
}

func (h *KinesisHandler) initMeasureFromAWS(
	project *org.Project,
	dest *Measure,
	src *CloudwatchMeasure,
) error {
	attrs := make(AttrMap, len(src.Dimensions)+4)
	for key, value := range src.Dimensions {
		key = attrkey.Clean(key)
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
		attrs["aws.metric_stream_name"] = src.MetricStreamName
	}

	dest.ProjectID = project.ID
	dest.Metric = attrkey.AWSMetricName(src.Namespace, src.MetricName)
	dest.Unit = bununit.FromString(src.Unit)
	dest.Attrs = attrs

	dest.Time = time.Unix(0, src.Timestamp*int64(time.Millisecond))
	dest.Instrument = InstrumentSummary
	dest.Sum = src.Value.Sum
	dest.Count = uint64(src.Value.Count)
	dest.Min = src.Value.Min
	dest.Max = src.Value.Max

	return nil
}
