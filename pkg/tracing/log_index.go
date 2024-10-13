package tracing

import (
	"time"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/uptrace/uptrace/pkg/utf8util"
	"golang.org/x/exp/slices"
)

type LogIndex struct {
	ch.CHModel `ch:"table:logs_index,insert:logs_index_buffer,alias:s"`

	ID       idgen.SpanID  `json:"id" msgpack:"-" ch:"id"`
	ParentID idgen.SpanID  `json:"parentId,omitempty" msgpack:"-"`
	TraceID  idgen.TraceID `json:"traceId" msgpack:"-" ch:"type:UUID"`

	ProjectID uint32 `json:"projectId" msgpack:"-"`
	Type      string `json:"-" msgpack:"-" ch:",lc"`
	System    string `json:"system" ch:",lc"`
	GroupID   uint64 `json:"groupId,string"`
	Kind      string `json:"kind" ch:",lc"`

	Name       string    `json:"name" ch:",lc"`
	EventName  string    `json:"eventName,omitempty" ch:",lc"`
	Time       time.Time `json:"time" msgpack:"-"`
	StatusCode string    `json:"statusCode" ch:",lc"`

	DisplayName string `json:"displayName"`

	TelemetrySDKName     string `ch:",lc"`
	TelemetrySDKLanguage string `ch:",lc"`
	TelemetrySDKVersion  string `ch:",lc"`
	TelemetryAutoVersion string `ch:",lc"`

	OtelLibraryName    string `ch:",lc"`
	OtelLibraryVersion string `ch:",lc"`

	DeploymentEnvironment string `ch:",lc"`

	ServiceName      string `ch:",lc"`
	ServiceVersion   string `ch:",lc"`
	ServiceNamespace string `ch:",lc"`
	HostName         string `ch:",lc"`

	ClientAddress       string `ch:",lc"`
	ClientSocketAddress string `ch:",lc"`
	ClientSocketPort    int32

	URLScheme string `attr:"url.scheme" ch:",lc"`
	URLFull   string `attr:"url.full"`
	URLPath   string `attr:"url.path" ch:",lc"`

	HTTPRequestMethod      string `ch:",lc"`
	HTTPResponseStatusCode uint16
	HTTPRoute              string `ch:",lc"`

	RPCMethod  string `ch:",lc"`
	RPCService string `ch:",lc"`

	DBSystem    string `ch:",lc"`
	DBName      string `ch:",lc"`
	DBStatement string
	DBOperation string `ch:",lc"`
	DBSqlTable  string `ch:",lc"`

	LogSeverity   string `ch:",lc"`
	ExceptionType string `ch:",lc"`

	AllKeys      []string `ch:"type:Array(LowCardinality(String))"`
	StringKeys   []string `ch:"type:Array(LowCardinality(String))"`
	StringValues []string
}

func (index *LogIndex) init(span *Span) {

	index.ID = span.ID
	index.TraceID = span.TraceID
	index.ParentID = span.ParentID

	index.ProjectID = span.ProjectID
	index.Type = span.Type
	index.System = span.System
	index.GroupID = span.GroupID

	index.Kind = span.Kind
	index.Name = span.Name
	index.EventName = span.EventName
	index.DisplayName = utf8util.TruncLarge(span.DisplayName)

	index.Time = span.Time

	index.StatusCode = span.StatusCode

	index.TelemetrySDKName = span.Attrs.Text(attrkey.TelemetrySDKName)
	index.TelemetrySDKLanguage = span.Attrs.Text(attrkey.TelemetrySDKLanguage)
	index.TelemetrySDKVersion = span.Attrs.Text(attrkey.TelemetrySDKVersion)
	index.TelemetryAutoVersion = span.Attrs.Text(attrkey.TelemetryAutoVersion)

	index.OtelLibraryName = span.Attrs.Text(attrkey.OtelLibraryName)
	index.OtelLibraryVersion = span.Attrs.Text(attrkey.OtelLibraryVersion)

	index.DeploymentEnvironment, _ = span.Attrs[attrkey.DeploymentEnvironment].(string)

	index.ServiceName = span.Attrs.ServiceName()
	index.ServiceVersion = span.Attrs.Text(attrkey.ServiceVersion)
	index.ServiceNamespace = span.Attrs.Text(attrkey.ServiceNamespace)
	index.HostName = span.Attrs.HostName()

	index.ClientAddress = span.Attrs.Text(attrkey.ClientAddress)
	index.ClientSocketAddress = span.Attrs.Text(attrkey.ClientSocketAddress)
	index.ClientSocketPort = int32(span.Attrs.Int64(attrkey.ClientSocketPort))

	index.URLScheme = span.Attrs.Text(attrkey.URLScheme)
	index.URLFull = span.Attrs.Text(attrkey.URLFull)
	index.URLPath = span.Attrs.Text(attrkey.URLPath)

	index.HTTPRequestMethod = span.Attrs.Text(attrkey.HTTPRequestMethod)
	index.HTTPResponseStatusCode = uint16(span.Attrs.Uint64(attrkey.HTTPResponseStatusCode))
	index.HTTPRoute = span.Attrs.Text(attrkey.HTTPRoute)

	index.RPCMethod = span.Attrs.Text(attrkey.RPCMethod)
	index.RPCService = span.Attrs.Text(attrkey.RPCService)

	index.DBSystem = span.Attrs.Text(attrkey.DBSystem)
	index.DBName = span.Attrs.Text(attrkey.DBName)
	index.DBStatement = span.Attrs.Text(attrkey.DBStatement)
	index.DBOperation = span.Attrs.Text(attrkey.DBOperation)
	index.DBSqlTable = span.Attrs.Text(attrkey.DBSqlTable)

	index.LogSeverity = span.Attrs.Text(attrkey.LogSeverity)
	index.ExceptionType = span.Attrs.Text(attrkey.ExceptionType)

	index.AllKeys = mapKeys(span.Attrs)
	slices.Sort(index.AllKeys)

	index.StringKeys, index.StringValues = attrKeysAndValues(span.Attrs, index.AllKeys)
}
