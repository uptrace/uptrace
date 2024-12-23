package ch

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/uptrace/pkg/clickhouse/ch/internal"
	"go4.org/syncutil"
	"log/slog"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

type Supercluster struct {
	*Clusters
	weights []int
	healthy atomic.Pointer[Clusters]
}
type Clusters struct {
	clusters    []*Cluster
	weighted    []*Cluster
	nextCluster atomic.Uint32
}

func (s *Clusters) Len() int { return len(s.clusters) }
func (s *Clusters) All(yield func(cl *Cluster) bool) {
	for _, cl := range s.clusters {
		if !yield(cl) {
			break
		}
	}
}
func (s *Clusters) Shards(yield func(cl *Cluster, shard *DB) bool) {
	s.All(func(cl *Cluster) bool {
		var stop bool
		cl.Shards(func(shard *DB) bool { stop = yield(cl, shard); return stop })
		return stop
	})
}
func (s *Clusters) Replicas(yield func(cl *Cluster, replica *DB) bool) {
	s.All(func(cl *Cluster) bool {
		for _, replica := range cl.Replicas() {
			if !yield(cl, replica) {
				return false
			}
		}
		return true
	})
}
func (s *Clusters) ByID(index int) *Cluster { return s.clusters[index] }
func (s *Clusters) Cluster(clientID uint64) *Cluster {
	idx := clientID % uint64(len(s.weighted))
	return s.weighted[idx]
}
func (s *Clusters) Rand() *Cluster { i := int(s.nextCluster.Add(1)); return s.ByID(i % s.Len()) }
func (s *Clusters) CheckHealth() error {
	var err error
	s.All(func(cl *Cluster) bool {
		if err = cl.CheckHealth(); err != nil {
			return false
		}
		return true
	})
	return err
}
func NewSupercluster(clusters []*Cluster, weights []int) *Supercluster {
	if len(clusters) != len(weights) {
		panic("clusters and weights have different length")
	}
	if slices.Equal(weights, make([]int, len(weights))) {
		slog.Warn("all cluster weights are zero")
	}
	s := new(Supercluster)
	s.weights = weights
	s.Clusters = s.newClusters(clusters, false)
	s.healthy.Store(s.newHealthyClusters())
	return s
}
func (s *Supercluster) Len() int           { return len(s.clusters) }
func (s *Supercluster) Healthy() *Clusters { return s.healthy.Load() }
func (s *Supercluster) ForEachCluster(fn func(cl *Cluster) error) error {
	var group syncutil.Group
	for _, cl := range s.clusters {
		group.Go(func() error { return fn(cl) })
	}
	return group.Err()
}
func (s *Supercluster) ForEachShard(fn func(cl *Cluster, shard *DB) error) error {
	var group syncutil.Group
	for _, cl := range s.clusters {
		group.Go(func() error { return cl.ForEachShard(func(shard *DB) error { return fn(cl, shard) }) })
	}
	return group.Err()
}
func (s *Supercluster) ForEachHealthyShard(fn func(cl *Cluster, shard *DB) error) error {
	var group syncutil.Group
	s.Healthy().All(func(cl *Cluster) bool {
		group.Go(func() error { return cl.ForEachHealthyShard(func(shard *DB) error { return fn(cl, shard) }) })
		return true
	})
	return group.Err()
}
func (s *Supercluster) ForEachReplica(fn func(cl *Cluster, replica *DB) error) error {
	var group syncutil.Group
	for _, cl := range s.clusters {
		group.Go(func() error { return cl.ForEachReplica(func(replica *DB) error { return fn(cl, replica) }) })
	}
	return group.Err()
}
func (s *Supercluster) ForEachHealthyReplica(fn func(cl *Cluster, replica *DB) error) error {
	var group syncutil.Group
	s.Healthy().All(func(cl *Cluster) bool {
		group.Go(func() error { return cl.ForEachHealthyReplica(func(replica *DB) error { return fn(cl, replica) }) })
		return true
	})
	return group.Err()
}
func (s *Supercluster) ReplicaByAddr(replicaAddr string) (*Cluster, *DB, error) {
	var foundCluster *Cluster
	var foundReplica *DB
	s.Replicas(func(cluster *Cluster, replica *DB) bool {
		if replica.Config().Addr == replicaAddr {
			foundCluster = cluster
			foundReplica = replica
			return false
		}
		return true
	})
	if foundCluster == nil || foundReplica == nil {
		return nil, nil, fmt.Errorf("can't find replica with addr=%q", replicaAddr)
	}
	return foundCluster, foundReplica, nil
}
func (s *Supercluster) Monitor(ctx context.Context) {
	for {
		if err := sleep(ctx, 3*time.Second); err != nil {
			return
		}
		s.healthy.Store(s.newHealthyClusters())
	}
}
func (s *Supercluster) newHealthyClusters() *Clusters {
	if res := s.newClusters(s.clusters, true); len(res.clusters) > 0 {
		return res
	}
	return s.newClusters(s.clusters, false)
}
func (s *Supercluster) newClusters(allClusters []*Cluster, checkHealth bool) *Clusters {
	clusters := make([]*Cluster, 0, len(allClusters))
	weighted := make([]*Cluster, 0, len(allClusters))
	for i, cluster := range allClusters {
		if checkHealth {
			if len(cluster.HealthyReplicas()) == 0 {
				continue
			}
		}
		weight := s.weights[i]
		if weight == 0 {
			continue
		}
		clusters = append(clusters, cluster)
		for i := 0; i < weight; i++ {
			weighted = append(weighted, cluster)
		}
	}
	return &Clusters{clusters: clusters, weighted: weighted}
}

type Cluster struct {
	id                int
	shards            [][]*DB
	healthyShards     atomic.Pointer[[][]*DB]
	replicas          []*DB
	healthyReplicas   atomic.Pointer[[]*DB]
	allReplicas       [][]*clusterReplica
	checkHealthResult atomic.Pointer[errorHolder]
	nextReplica       atomic.Uint32
}

func NewCluster(id int, shards [][]*DB) *Cluster {
	var replicas []*DB
	for _, shard := range shards {
		replicas = append(replicas, shard...)
	}
	if len(replicas) == 0 {
		panic("ClickHouse Cluster requires at least one replica")
	}
	allReplicas := make([][]*clusterReplica, len(shards))
	for i, shardReplicas := range shards {
		tmp := make([]*clusterReplica, len(shardReplicas))
		for i, replica := range shardReplicas {
			tmp[i] = &clusterReplica{db: replica, isBackup: replica.Config().IsBackup, errors: internal.NewErrorCounter(60, 5)}
		}
		slices.SortFunc(tmp, func(a, b *clusterReplica) int { return bool2int(a.isBackup) - bool2int(b.isBackup) })
		allReplicas[i] = tmp
	}
	cl := &Cluster{id: id, shards: shards, replicas: replicas, allReplicas: allReplicas}
	cl.healthyShards.Store(&shards)
	cl.healthyReplicas.Store(&replicas)
	return cl
}
func (cl *Cluster) ID() int                { return cl.id }
func (cl *Cluster) Replicas() []*DB        { return cl.replicas }
func (cl *Cluster) HealthyReplicas() []*DB { return *cl.healthyReplicas.Load() }
func (cl *Cluster) Replica(clientID uint64) *DB {
	healthyReplicas := cl.HealthyReplicas()
	idx := clientID % uint64(len(healthyReplicas))
	return healthyReplicas[idx]
}
func (cl *Cluster) RandReplica() *DB {
	healthyReplicas := cl.HealthyReplicas()
	if len(healthyReplicas) == 1 {
		return healthyReplicas[0]
	}
	i := int(cl.nextReplica.Add(1))
	return healthyReplicas[i%len(healthyReplicas)]
}
func (cl *Cluster) ForEachShard(fn func(shard *DB) error) error {
	return cl.forEachShard(cl.shards, fn)
}
func (cl *Cluster) ForEachHealthyShard(fn func(shard *DB) error) error {
	return cl.forEachShard(*cl.healthyShards.Load(), fn)
}
func (cl *Cluster) forEachShard(shards [][]*DB, fn func(shard *DB) error) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	cl.iterShards(shards, func(shard *DB) bool {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(shard); err != nil {
				select {
				case errCh <- err:
				default:
				}
			}
		}()
		return true
	})
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
func (cl *Cluster) Shards(yield func(shard *DB) bool) { cl.iterShards(cl.shards, yield) }
func (cl *Cluster) HealthyShards(yield func(shard *DB) bool) {
	cl.iterShards(*cl.healthyShards.Load(), yield)
}
func (cl *Cluster) iterShards(shards [][]*DB, yield func(shard *DB) bool) {
	for _, replicas := range shards {
		replica := replicas[0]
		if !yield(replica) {
			return
		}
	}
}
func (cl *Cluster) ForEachReplica(fn func(replica *DB) error) error {
	return cl.forEachReplica(cl.replicas, fn)
}
func (cl *Cluster) ForEachHealthyReplica(fn func(replica *DB) error) error {
	return cl.forEachReplica(cl.HealthyReplicas(), fn)
}
func (cl *Cluster) forEachReplica(replicas []*DB, fn func(replica *DB) error) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	wg.Add(len(replicas))
	for _, replica := range replicas {
		go func(replica *DB) {
			defer wg.Done()
			if err := fn(replica); err != nil {
				select {
				case errCh <- err:
				default:
				}
			}
		}(replica)
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

type clusterReplica struct {
	db       *DB
	isBackup bool
	errors   *internal.ErrorCounter
}

func (r *clusterReplica) String() string { return r.db.String() }
func (r *clusterReplica) ping(ctx context.Context) error {
	var err error
	for i := 0; i < 3; i++ {
		err = r.db.Ping(ctx)
		if err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return err
}
func (cl *Cluster) CheckHealth() error {
	if h := cl.checkHealthResult.Load(); h != nil {
		return h.error
	}
	return nil
}
func (cl *Cluster) Monitor(ctx context.Context) {
	for {
		if err := sleep(ctx, 3*time.Second); err != nil {
			return
		}
		cl.monitorTick(ctx)
	}
}
func (cl *Cluster) monitorTick(ctx context.Context) {
	healthyShards := make([][]*DB, len(cl.shards))
	var lastErr error
	for i, shardReplicas := range cl.allReplicas {
		healthyReplicas := make([]*DB, 0, len(shardReplicas))
		for _, replica := range shardReplicas {
			replicaErr := cl.checkReplica(ctx, replica)
			if err := replica.errors.Add(replicaErr); err != nil {
				lastErr = err
			}
			if replicaErr != nil {
				if replicaErr != ErrClosed {
					slog.Error("replica is unheathy", slog.String("replica", replica.String()), slog.Any("err", replicaErr))
				}
				continue
			}
			if len(healthyReplicas) > 0 && replica.isBackup {
				continue
			}
			healthyReplicas = append(healthyReplicas, replica.db)
		}
		if len(healthyReplicas) == 0 {
			healthyReplicas = cl.shards[i]
		}
		healthyShards[i] = healthyReplicas
	}
	healthyReplicas := make([]*DB, 0, len(cl.replicas))
	for _, shardReplicas := range healthyShards {
		healthyReplicas = append(healthyReplicas, shardReplicas...)
	}
	cl.healthyShards.Store(&healthyShards)
	cl.healthyReplicas.Store(&healthyReplicas)
	cl.checkHealthResult.Store(&errorHolder{lastErr})
}
func (cl *Cluster) checkReplica(ctx context.Context, replica *clusterReplica) error {
	if err := replica.ping(ctx); err != nil {
		return err
	}
	if err := cl.checkNumPartition(ctx, replica); err != nil {
		return err
	}
	if err := cl.checkReplication(ctx, replica); err != nil {
		return err
	}
	return nil
}
func (cl *Cluster) checkNumPartition(ctx context.Context, replica *clusterReplica) error {
	var table, partition string
	var numPart int
	if err := replica.db.NewSelect().ColumnExpr("table, partition, count()").TableExpr("system.parts").GroupExpr("table, partition").Where("active").Having("count() >= 200").Limit(1).Scan(ctx, &table, &partition, &numPart); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return fmt.Errorf("table=%q on replica=%q has too many parts (%d) in partition=%q", table, replica, numPart, partition)
}
func (cl *Cluster) checkReplication(ctx context.Context, replica *clusterReplica) error {
	var num int
	if err := replica.db.NewSelect().ColumnExpr("count()").TableExpr("system.replication_queue").Scan(ctx, &num); err != nil {
		return err
	}
	if num >= 1000 {
		return fmt.Errorf("replica=%q has large replication queue: %d", replica, num)
	}
	return nil
}

type errorHolder struct{ error }

func sleep(ctx context.Context, d time.Duration) error {
	done := ctx.Done()
	if done == nil {
		time.Sleep(d)
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-done:
		return ctx.Err()
	}
}
func bool2int(v bool) int {
	if v {
		return 1
	}
	return 0
}
