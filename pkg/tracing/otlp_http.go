package tracing

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/uptrace/bunrouter"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	pbContentType   = "application/x-protobuf"
	jsonContentType = "application/json"
)

func (s *TraceServiceServer) httpTraces(w http.ResponseWriter, req bunrouter.Request) error {
	switch contentType := req.Header.Get("content-type"); contentType {
	case jsonContentType:
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}

		td := new(tracepb.TracesData)
		if err := protojson.Unmarshal(body, td); err != nil {
			return err
		}

		s.process(td.ResourceSpans)

		resp := new(collectortrace.ExportTraceServiceResponse)
		b, err := protojson.Marshal(resp)
		if err != nil {
			return err
		}

		if _, err := w.Write(b); err != nil {
			return err
		}

		return nil
	case pbContentType:
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return err
		}

		td := new(collectortrace.ExportTraceServiceRequest)
		if err := proto.Unmarshal(body, td); err != nil {
			return err
		}

		s.process(td.ResourceSpans)

		resp := new(collectortrace.ExportTraceServiceResponse)
		b, err := proto.Marshal(resp)
		if err != nil {
			return err
		}

		if _, err := w.Write(b); err != nil {
			return err
		}

		return nil
	default:
		return fmt.Errorf("unsupported content type: %q", contentType)
	}
}
