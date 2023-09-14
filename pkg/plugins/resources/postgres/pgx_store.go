package postgres

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	config "github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/common/postgres"
	pgx_config "github.com/kumahq/kuma/pkg/plugins/resources/postgres/config"
)

type pgxResourceStore struct {
	pool                 *pgxpool.Pool
	roPool               *pgxpool.Pool
	roRatio              uint
	maxListQueryElements uint32
}

type ResourceNamesByMesh map[string][]string

var _ store.ResourceStore = &pgxResourceStore{}

func NewPgxStore(metrics core_metrics.Metrics, config config.PostgresStoreConfig, customizer pgx_config.PgxConfigCustomization) (store.ResourceStore, error) {
	pool, err := postgres.ConnectToDbPgx(config, customizer)
	if err != nil {
		return nil, err
	}
	var roPool *pgxpool.Pool
	if config.ReadReplica.Host != "" {
		roConfig := config
		roConfig.Host = config.ReadReplica.Host
		roConfig.Port = int(config.ReadReplica.Port)
		roPool, err = postgres.ConnectToDbPgx(roConfig)
		if err != nil {
			return nil, err
		}
		if err := registerMetrics(metrics, roPool, "ro"); err != nil {
			return nil, errors.Wrapf(err, "could not register DB metrics")
		}
	}

	if err := registerMetrics(metrics, pool, "rw"); err != nil {
		return nil, errors.Wrapf(err, "could not register DB metrics")
	}

	return &pgxResourceStore{
		pool:                 pool,
		roPool:               roPool,
		maxListQueryElements: config.MaxListQueryElements,
		roRatio:              config.ReadReplica.Ratio,
	}, nil
}

func (r *pgxResourceStore) Create(ctx context.Context, resource core_model.Resource, fs ...store.CreateOptionsFunc) error {
	opts := store.NewCreateOptions(fs...)

	bytes, err := core_model.ToJSON(resource.GetSpec())
	if err != nil {
		return errors.Wrap(err, "failed to convert spec to json")
	}

	var ownerName *string
	var ownerMesh *string
	var ownerType *string

	if opts.Owner != nil {
		ptr := func(s string) *string { return &s }
		ownerName = ptr(opts.Owner.GetMeta().GetName())
		ownerMesh = ptr(opts.Owner.GetMeta().GetMesh())
		ownerType = ptr(string(opts.Owner.Descriptor().Name))
	}

	version := 0
	statement := `INSERT INTO resources VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`
	_, err = r.pool.Exec(ctx, statement, opts.Name, opts.Mesh, resource.Descriptor().Name, version, string(bytes),
		opts.CreationTime.UTC(), opts.CreationTime.UTC(), ownerName, ownerMesh, ownerType)
	if err != nil {
		if strings.Contains(err.Error(), duplicateKeyErrorMsg) {
			return store.ErrorResourceAlreadyExists(resource.Descriptor().Name, opts.Name, opts.Mesh)
		}
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}

	resource.SetMeta(&resourceMetaObject{
		Name:             opts.Name,
		Mesh:             opts.Mesh,
		Version:          strconv.Itoa(version),
		CreationTime:     opts.CreationTime,
		ModificationTime: opts.CreationTime,
	})
	return nil
}

func (r *pgxResourceStore) Update(ctx context.Context, resource core_model.Resource, fs ...store.UpdateOptionsFunc) error {
	bytes, err := core_model.ToJSON(resource.GetSpec())
	if err != nil {
		return err
	}

	opts := store.NewUpdateOptions(fs...)

	version, err := strconv.Atoi(resource.GetMeta().GetVersion())
	newVersion := version + 1
	if err != nil {
		return errors.Wrap(err, "failed to convert meta version to int")
	}
	statement := `UPDATE resources SET spec=$1, version=$2, modification_time=$3 WHERE name=$4 AND mesh=$5 AND type=$6 AND version=$7;`
	result, err := r.pool.Exec(
		ctx,
		statement,
		string(bytes),
		newVersion,
		opts.ModificationTime.UTC(),
		resource.GetMeta().GetName(),
		resource.GetMeta().GetMesh(),
		resource.Descriptor().Name,
		version,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to execute query %s", statement)
	}
	if rows := result.RowsAffected(); rows != 1 {
		return store.ErrorResourceConflict(resource.Descriptor().Name, resource.GetMeta().GetName(), resource.GetMeta().GetMesh())
	}

	// update resource's meta with new version
	resource.SetMeta(&resourceMetaObject{
		Name:             resource.GetMeta().GetName(),
		Mesh:             resource.GetMeta().GetMesh(),
		Version:          strconv.Itoa(newVersion),
		ModificationTime: opts.ModificationTime,
	})

	return nil
}

func (r *pgxResourceStore) Delete(ctx context.Context, resource core_model.Resource, fs ...store.DeleteOptionsFunc) error {
	opts := store.NewDeleteOptions(fs...)

	statement := `DELETE FROM resources WHERE name=$1 AND type=$2 AND mesh=$3`
	result, err := r.pool.Exec(ctx, statement, opts.Name, resource.Descriptor().Name, opts.Mesh)
	if err != nil {
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}
	if rows := result.RowsAffected(); rows == 0 {
		return store.ErrorResourceNotFound(resource.Descriptor().Name, opts.Name, opts.Mesh)
	}

	return nil
}

func (r *pgxResourceStore) Get(ctx context.Context, resource core_model.Resource, fs ...store.GetOptionsFunc) error {
	opts := store.NewGetOptions(fs...)

	statement := `SELECT spec, version, creation_time, modification_time FROM resources WHERE name=$1 AND mesh=$2 AND type=$3;`
	row := r.pickRoPool().QueryRow(ctx, statement, opts.Name, opts.Mesh, resource.Descriptor().Name)

	var spec string
	var version int
	var creationTime, modificationTime time.Time
	err := row.Scan(&spec, &version, &creationTime, &modificationTime)
	if err == pgx.ErrNoRows {
		return store.ErrorResourceNotFound(resource.Descriptor().Name, opts.Name, opts.Mesh)
	}
	if err != nil {
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}

	if err := core_model.FromJSON([]byte(spec), resource.GetSpec()); err != nil {
		return errors.Wrap(err, "failed to convert json to spec")
	}

	meta := &resourceMetaObject{
		Name:             opts.Name,
		Mesh:             opts.Mesh,
		Version:          strconv.Itoa(version),
		CreationTime:     creationTime.Local(),
		ModificationTime: modificationTime.Local(),
	}
	resource.SetMeta(meta)

	if opts.Version != "" && resource.GetMeta().GetVersion() != opts.Version {
		return store.ErrorResourcePreconditionFailed(resource.Descriptor().Name, opts.Name, opts.Mesh)
	}
	return nil
}

func (r *pgxResourceStore) pickRoPool() *pgxpool.Pool {
	if r.roPool == nil {
		return r.pool
	}
	// #nosec G404 - math rand is enough
	if rand.Int31n(101) <= int32(r.roRatio) {
		return r.roPool
	}
	return r.pool
}

func (r *pgxResourceStore) List(ctx context.Context, resources core_model.ResourceList, args ...store.ListOptionsFunc) error {
	opts := store.NewListOptions(args...)

	statement := `SELECT name, mesh, spec, version, creation_time, modification_time FROM resources WHERE type=$1`
	var statementArgs []interface{}
	statementArgs = append(statementArgs, resources.GetItemType())
	argsIndex := 1
	rkSize := len(opts.ResourceKeys)
	if rkSize > 0 && rkSize < int(r.maxListQueryElements) {
		statement += " AND ("
		res := resourceNamesByMesh(opts.ResourceKeys)
		iter := 0
		for mesh, names := range res {
			if iter > 0 {
				statement += " OR "
			}
			argsIndex++
			statement += fmt.Sprintf("(mesh=$%d AND", argsIndex)
			statementArgs = append(statementArgs, mesh)
			for idx, name := range names {
				argsIndex++
				if idx == 0 {
					statement += fmt.Sprintf(" name IN ($%d", argsIndex)
				} else {
					statement += fmt.Sprintf(",$%d", argsIndex)
				}
				statementArgs = append(statementArgs, name)
			}
			statement += "))"
			iter++
		}
		statement += ")"
	}
	if opts.Mesh != "" {
		argsIndex++
		statement += fmt.Sprintf(" AND mesh=$%d", argsIndex)
		statementArgs = append(statementArgs, opts.Mesh)
	}
	if opts.NameContains != "" {
		argsIndex++
		statement += fmt.Sprintf(" AND name LIKE $%d", argsIndex)
		statementArgs = append(statementArgs, "%"+opts.NameContains+"%")
	}
	statement += " ORDER BY name, mesh"

	rows, err := r.pickRoPool().Query(ctx, statement, statementArgs...)
	if err != nil {
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		item, err := rowToItem(resources, rows)
		if err != nil {
			return err
		}
		if err := resources.AddItem(item); err != nil {
			return err
		}
		total++
	}

	resources.GetPagination().SetTotal(uint32(total))
	return nil
}

func resourceNamesByMesh(resourceKeys map[core_model.ResourceKey]struct{}) ResourceNamesByMesh {
	resourceNamesByMesh := ResourceNamesByMesh{}
	for key := range resourceKeys {
		if val, exists := resourceNamesByMesh[key.Mesh]; exists {
			resourceNamesByMesh[key.Mesh] = append(val, key.Name)
		} else {
			resourceNamesByMesh[key.Mesh] = []string{key.Name}
		}
	}
	return resourceNamesByMesh
}

func rowToItem(resources core_model.ResourceList, rows pgx.Rows) (core_model.Resource, error) {
	var name, mesh, spec string
	var version int
	var creationTime, modificationTime time.Time
	if err := rows.Scan(&name, &mesh, &spec, &version, &creationTime, &modificationTime); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve elements from query")
	}

	item := resources.NewItem()
	if err := core_model.FromJSON([]byte(spec), item.GetSpec()); err != nil {
		return nil, errors.Wrap(err, "failed to convert json to spec")
	}

	meta := &resourceMetaObject{
		Name:             name,
		Mesh:             mesh,
		Version:          strconv.Itoa(version),
		CreationTime:     creationTime.Local(),
		ModificationTime: modificationTime.Local(),
	}
	item.SetMeta(meta)

	return item, nil
}

func (r *pgxResourceStore) Close() error {
	r.pool.Close()
	if r.roPool != nil {
		r.roPool.Close()
	}
	return nil
}

func registerMetrics(metrics core_metrics.Metrics, pool *pgxpool.Pool, poolName string) error {
	postgresCurrentConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connections",
		Help: "Current number of postgres store connections",
		ConstLabels: map[string]string{
			"type": "open_connections",
			"pool": poolName,
		},
	}, func() float64 {
		return float64(pool.Stat().TotalConns())
	})

	postgresInUseConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connections",
		Help: "Current number of postgres store connections",
		ConstLabels: map[string]string{
			"type": "in_use",
			"pool": poolName,
		},
	}, func() float64 {
		return float64(pool.Stat().AcquiredConns())
	})

	postgresIdleConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connections",
		Help: "Current number of postgres store connections",
		ConstLabels: map[string]string{
			"type": "idle",
			"pool": poolName,
		},
	}, func() float64 {
		return float64(pool.Stat().IdleConns())
	})

	postgresMaxOpenConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connections_max",
		Help: "Max postgres store open connections",
		ConstLabels: map[string]string{
			"pool": poolName,
		},
	}, func() float64 {
		return float64(pool.Stat().MaxConns())
	})

	postgresWaitConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connection_wait_count",
		Help: "Current waiting postgres store connections",
		ConstLabels: map[string]string{
			"pool": poolName,
		},
	}, func() float64 {
		return float64(pool.Stat().EmptyAcquireCount())
	})

	postgresWaitConnectionDurationMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connection_wait_duration",
		Help: "Time Blocked waiting for new connection in seconds",
		ConstLabels: map[string]string{
			"pool": poolName,
		},
	}, func() float64 {
		return pool.Stat().AcquireDuration().Seconds()
	})

	postgresMaxIdleClosedConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connection_closed",
		Help: "Current number of closed postgres store connections",
		ConstLabels: map[string]string{
			"type": "max_idle_conns",
			"pool": poolName,
		},
	}, func() float64 {
		return float64(pool.Stat().MaxIdleDestroyCount())
	})

	postgresMaxLifeTimeClosedConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connection_closed",
		Help: "Current number of closed postgres store connections",
		ConstLabels: map[string]string{
			"type": "conn_max_life_time",
			"pool": poolName,
		},
	}, func() float64 {
		return float64(pool.Stat().MaxLifetimeDestroyCount())
	})

	postgresSuccessfulAcquireCountMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connection_acquire",
		Help: "Cumulative count of acquires from the pool",
		ConstLabels: map[string]string{
			"type": "successful",
			"pool": poolName,
		},
	}, func() float64 {
		return float64(pool.Stat().AcquireCount())
	})

	postgresCanceledAcquireCountMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connection_acquire",
		Help: "Cumulative count of acquires from the pool",
		ConstLabels: map[string]string{
			"type": "canceled",
			"pool": poolName,
		},
	}, func() float64 {
		return float64(pool.Stat().CanceledAcquireCount())
	})

	postgresConstructingConnectionsCountMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connections",
		Help: "Current number of postgres store connections",
		ConstLabels: map[string]string{
			"type": "constructing",
			"pool": poolName,
		},
	}, func() float64 {
		return float64(pool.Stat().ConstructingConns())
	})

	postgresNewConnectionsCountMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connections",
		Help: "Current number of postgres store connections",
		ConstLabels: map[string]string{
			"type": "new",
			"pool": poolName,
		},
	}, func() float64 {
		return float64(pool.Stat().NewConnsCount())
	})

	if err := metrics.
		BulkRegister(postgresCurrentConnectionMetric, postgresInUseConnectionMetric, postgresIdleConnectionMetric,
			postgresMaxOpenConnectionMetric, postgresWaitConnectionMetric, postgresWaitConnectionDurationMetric,
			postgresMaxIdleClosedConnectionMetric, postgresMaxLifeTimeClosedConnectionMetric, postgresSuccessfulAcquireCountMetric,
			postgresCanceledAcquireCountMetric, postgresConstructingConnectionsCountMetric, postgresNewConnectionsCountMetric,
		); err != nil {
		return err
	}
	return nil
}
