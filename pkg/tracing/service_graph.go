package tracing

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/uptrace/uptrace/pkg/attrkey"
	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/idgen"
	"github.com/zyedidia/generic/list"
	"go.uber.org/zap"
)

const batchSize = 10000

type ServiceGraphEdge struct {
	ch.CHModel `ch:"service_graph_edges,insert:service_graph_edges_buffer,alias:e"`

	ProjectID uint32
	Type      EdgeType
	Time      time.Time `ch:"type:DateTime"`

	ClientAttr            string `ch:",lc"`
	ClientName            string `ch:",lc"`
	ServerAttr            string `ch:",lc"`
	ServerName            string `ch:",lc"`
	DeploymentEnvironment string `ch:",lc"`
	ServiceNamespace      string `ch:",lc"`

	ClientDurationMin float32
	ClientDurationMax float32
	ClientDurationSum float32

	ServerDurationMin float32
	ServerDurationMax float32
	ServerDurationSum float32

	Count      uint32
	ErrorCount uint32
}

type EdgeType string

const (
	EdgeTypeUnset     EdgeType = "unset"
	EdgeTypeHTTP      EdgeType = "http"
	EdgeTypeDB        EdgeType = "db"
	EdgeTypeMessaging EdgeType = "messaging"
)

func (e *ServiceGraphEdge) FillFrom(span *BaseIndex) {
	e.ProjectID = span.ProjectID
	clientAttr, clientName, serverAttr, serverName := serviceGraphNode(span)

	switch span.Kind {
	case SpanKindProducer, SpanKindClient:
		if clientName != "" {
			e.ClientAttr = clientAttr
			e.ClientName = clientName
		}
		if e.ServerName == "" {
			e.ServerAttr = serverAttr
			e.ServerName = serverName
		}
		e.Time = span.Time.Truncate(time.Minute)
		e.setClientDuration(span)
	case SpanKindConsumer, SpanKindServer:
		if e.ClientName == "" {
			e.ClientAttr = clientAttr
			e.ClientName = clientName
		}
		if serverName != "" {
			e.ServerAttr = serverAttr
			e.ServerName = serverName
		}
		if e.Time.IsZero() {
			e.Time = span.Time.Truncate(time.Minute)
		}
		e.setServerDuration(span)
	}

	if e.DeploymentEnvironment == "" {
		e.DeploymentEnvironment = span.DeploymentEnvironment
	}
	if e.ServiceNamespace == "" {
		e.ServiceNamespace = span.ServiceNamespace
	}
}

func (e *ServiceGraphEdge) setClientDuration(span *BaseIndex) {
	dur := float32(span.Duration)
	e.ClientDurationMin = dur
	e.ClientDurationMax = dur
	e.ClientDurationSum = dur

	e.Count = uint32(span.Count)
	if span.StatusCode == StatusCodeError {
		e.ErrorCount = e.Count
	}
}

func (e *ServiceGraphEdge) setServerDuration(span *BaseIndex) {
	dur := float32(span.Duration)
	e.ServerDurationMin = dur
	e.ServerDurationMax = dur
	e.ServerDurationSum = dur

	if e.Count == 0 {
		e.Count = uint32(span.Count)
		if span.StatusCode == StatusCodeError {
			e.ErrorCount = e.Count
		}
	}
}

type ServiceGraphProcessor struct {
	app *bunapp.App

	storeShards []*ServiceGraphStore
	edgeCh      chan *ServiceGraphEdge

}

func NewServiceGraphProcessor(app *bunapp.App) *ServiceGraphProcessor {
	conf := app.Config().ServiceGraph
	p := &ServiceGraphProcessor{
		app:    app,
		edgeCh: make(chan *ServiceGraphEdge, batchSize),
	}

	n := runtime.GOMAXPROCS(0)
	if n < 1 {
		n = 1
	}
	p.storeShards = make([]*ServiceGraphStore, n)
	for i := range p.storeShards {
		p.storeShards[i] = NewServiceGraphStore(
			conf.Store.Size/n,
			conf.Store.TTL,
			p.onCompleteEdge,
			p.onExpiredEdge,
		)
	}

	go p.insertEdgesLoop(app.Context())

	return p
}

func (p *ServiceGraphProcessor) ProcessSpan(
	ctx context.Context, span *BaseIndex,
) error {
	edgeType := EdgeTypeUnset
	switch span.Kind {
	case SpanKindProducer:
		edgeType = EdgeTypeMessaging
		fallthrough
	case SpanKindClient:
		if edgeType == EdgeTypeUnset {
			edgeType = edgeTypeFromSpanType(span.Type)
		}

		key := ServiceGraphEdgeKey{
			TraceID: span.TraceID,
			SpanID:  span.ID,
		}
		p.store(span.TraceID).WithEdge(ctx, key, func(edge *ServiceGraphEdge) {
			if edgeType != EdgeTypeUnset {
				edge.Type = edgeType
			}
			edge.FillFrom(span)
		})
		return nil
	case SpanKindConsumer:
		edgeType = EdgeTypeMessaging
		fallthrough
	case SpanKindServer:
		if edgeType == EdgeTypeUnset {
			edgeType = edgeTypeFromSpanType(span.Type)
		}

		key := ServiceGraphEdgeKey{
			TraceID: span.TraceID,
			SpanID:  span.ParentID,
		}
		p.store(span.TraceID).WithEdge(ctx, key, func(edge *ServiceGraphEdge) {
			if edgeType != EdgeTypeUnset {
				edge.Type = edgeType
			}
			edge.FillFrom(span)
		})
		return nil
	default:
		return nil
	}
}

func serviceGraphNode(
	span *BaseIndex,
) (clientAttr, clientName, serverAttr, serverName string) {
	switch span.Type {
	case TypeSpanRPC:
		switch span.Kind {
		case SpanKindClient:
			clientAttr = attrkey.ServiceName
			clientName = span.ServiceName
			return clientAttr, clientName, "", ""
		case SpanKindServer:
			if span.ParentID == 0 {
				clientName = "<rpc-client>"
			}
			if span.RPCService != "" {
				serverAttr = attrkey.RPCService
				serverName = span.RPCService
				return clientAttr, clientName, serverAttr, serverName
			}
			if span.ServiceName != "" {
				serverAttr = attrkey.ServiceName
				serverName = span.ServiceName
				return clientAttr, clientName, serverAttr, serverName
			}
		}

	case TypeSpanDB:
		switch span.Kind {
		case SpanKindClient:
			clientAttr = attrkey.ServiceName
			clientName = span.ServiceName
		case SpanKindServer:
			if span.ParentID == 0 {
				clientName = "<db-client>"
			}
		}

		if span.DBName != "" {
			serverAttr = attrkey.DBName
			serverName = span.DBName
			return clientAttr, clientName, serverAttr, serverName
		}

		serverAttr = attrkey.SpanSystem
		serverName = span.System
		return clientAttr, clientName, serverAttr, serverName

	case TypeSpanHTTPServer:
		if span.ParentID == 0 {
			clientName = "<http-client>"
		}
		if domain := span.Attrs.GetAsLCString(attrkey.ServerSocketDomain); domain != "" {
			serverAttr = attrkey.ServerSocketDomain
			serverName = domain
			return clientAttr, clientName, serverAttr, serverName
		}
		if addr := span.Attrs.GetAsLCString(attrkey.ServerSocketAddress); addr != "" {
			serverAttr = attrkey.ServerSocketAddress
			serverName = addr
			return clientAttr, clientName, serverAttr, serverName
		}
		if span.ServiceName != "" {
			serverAttr = attrkey.ServiceName
			serverName = span.ServiceName
			return clientAttr, clientName, serverAttr, serverName
		}

	case TypeSpanMessaging:
		switch span.Kind {
		case SpanKindProducer:
			clientAttr = attrkey.ServiceName
			clientName = span.ServiceName
			if clientName == "" {
				clientAttr = attrkey.MessagingClientID
				clientName = span.Attrs.GetAsLCString(attrkey.MessagingClientID)
			}
			serverAttr = attrkey.MessagingDestinationName
			serverName = span.Attrs.GetAsLCString(serverAttr)
			if serverName != "" {
				return clientAttr, clientName, serverAttr, serverName
			}

		case SpanKindConsumer:
			if dest := span.Attrs.GetAsLCString(attrkey.MessagingDestinationName); dest != "" {
				clientAttr = attrkey.MessagingDestinationName
				clientName = dest
			}
			serverAttr = attrkey.MessagingKafkaConsumerGroup
			serverName = span.Attrs.GetAsLCString(serverAttr)
			if serverName != "" {
				return clientAttr, clientName, serverAttr, serverName
			}
		}
	}

	return "", "", "", ""
}

func edgeTypeFromSpanType(spanType string) EdgeType {
	switch spanType {
	case TypeSpanHTTPClient, TypeSpanHTTPServer:
		return EdgeTypeHTTP
	case TypeSpanDB:
		return EdgeTypeDB
	default:
		return EdgeTypeUnset
	}
}

func (p *ServiceGraphProcessor) store(traceID idgen.TraceID) *ServiceGraphStore {
	hash := xxhash.Sum64(traceID[:])
	return p.storeShards[hash%uint64(len(p.storeShards))]
}

func (p *ServiceGraphProcessor) onCompleteEdge(ctx context.Context, edge *ServiceGraphEdge) {
	select {
	case p.edgeCh <- edge:
	default:
		p.app.Zap(ctx).Error("edge chan is full (edge is dropped)",
			zap.Int("chan_len", len(p.edgeCh)))
	}
}

func (p *ServiceGraphProcessor) onExpiredEdge(ctx context.Context, edge *ServiceGraphEdge) {}

func (p *ServiceGraphProcessor) insertEdgesLoop(ctx context.Context) {
	const timeout = 10 * time.Second
	timer := time.NewTimer(timeout)

	edges := make([]*ServiceGraphEdge, 0, batchSize)
loop:
	for {
		select {
		case edge := <-p.edgeCh:
			edges = append(edges, edge)

			if len(edges) < cap(edges) {
				continue loop
			}

			if !timer.Stop() {
				<-timer.C
			}
		case <-timer.C:
		}

		if _, err := p.app.CH.NewInsert().
			Model(&edges).
			Exec(ctx); err != nil {
			p.app.Zap(ctx).Error("can't insert service graph edges", zap.Error(err))
		}

		edges = edges[:0]
		timer.Reset(timeout)
	}
}

type ServiceGraphStore struct {
	size int
	ttl  time.Duration

	onComplete func(ctx context.Context, edge *ServiceGraphEdge)
	onExpired  func(ctx context.Context, edge *ServiceGraphEdge)

	mu    sync.Mutex
	list  *list.List[*ServiceGraphEdgeNode]
	table map[ServiceGraphEdgeKey]*list.Node[*ServiceGraphEdgeNode]
}

type ServiceGraphEdgeKey struct {
	TraceID idgen.TraceID
	SpanID  idgen.SpanID
}

type ServiceGraphEdgeNode struct {
	ServiceGraphEdge

	key       ServiceGraphEdgeKey
	expiresAt time.Time
}

func (e *ServiceGraphEdgeNode) IsComplete() bool {
	return e.ClientName != "" && e.ServerName != ""
}

func (e *ServiceGraphEdgeNode) Expired() bool {
	return e.expiresAt.After(time.Now())
}

func NewServiceGraphStore(
	size int,
	ttl time.Duration,
	onComplete func(ctx context.Context, edge *ServiceGraphEdge),
	onExpired func(ctx context.Context, edge *ServiceGraphEdge),
) *ServiceGraphStore {
	return &ServiceGraphStore{
		size: size,
		ttl:  ttl,

		onComplete: onComplete,
		onExpired:  onExpired,

		list:  list.New[*ServiceGraphEdgeNode](),
		table: make(map[ServiceGraphEdgeKey]*list.Node[*ServiceGraphEdgeNode], size),
	}
}

var errStoreFull = errors.New("store is full")

func (s *ServiceGraphStore) WithEdge(
	ctx context.Context, key ServiceGraphEdgeKey, update func(edge *ServiceGraphEdge),
) (isNew bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if node, ok := s.table[key]; ok {
		edge := &node.Value.ServiceGraphEdge
		update(edge)

		if node.Value.IsComplete() {
			s.onComplete(ctx, edge)
			delete(s.table, key)
			s.list.Remove(node)
		}

		return false, nil
	}

	nodeValue := new(ServiceGraphEdgeNode)
	edge := &nodeValue.ServiceGraphEdge
	update(edge)

	if nodeValue.IsComplete() {
		s.onComplete(ctx, edge)
		return true, nil
	}

	if len(s.table) >= s.size {
		edge := s.tryRemoveFront()
		if edge == nil {
			return false, errStoreFull
		}
		s.onExpired(ctx, edge)
	}

	if s.ttl > 0 {
		nodeValue.expiresAt = time.Now().Add(s.ttl)
	}
	s.list.PushBack(nodeValue)
	s.table[key] = s.list.Back

	return true, nil
}

func (s *ServiceGraphStore) tryRemoveFront() *ServiceGraphEdge {
	if s.list.Front == nil {
		return nil
	}

	nodeValue := s.list.Front.Value
	if !nodeValue.Expired() {
		return nil
	}

	delete(s.table, nodeValue.key)
	s.list.Remove(s.list.Front)
	return &nodeValue.ServiceGraphEdge
}
