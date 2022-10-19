package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	config "github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	common_postgres "github.com/kumahq/kuma/pkg/plugins/common/postgres"
)

const duplicateKeyErrorMsg = "duplicate key value violates unique constraint"

type postgresResourceStore struct {
	db *sql.DB
}

var _ store.ResourceStore = &postgresResourceStore{}

func NewStore(metrics core_metrics.Metrics, config config.PostgresStoreConfig) (store.ResourceStore, error) {
	db, err := common_postgres.ConnectToDb(config)
	if err != nil {
		return nil, err
	}

	if err := registerMetrics(metrics, db); err != nil {
		return nil, errors.Wrapf(err, "could not register DB metrics")
	}

	return &postgresResourceStore{
		db: db,
	}, nil
}

func (r *postgresResourceStore) Create(_ context.Context, resource core_model.Resource, fs ...store.CreateOptionsFunc) error {
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
	_, err = r.db.Exec(statement, opts.Name, opts.Mesh, resource.Descriptor().Name, version, string(bytes),
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

func (r *postgresResourceStore) Update(_ context.Context, resource core_model.Resource, fs ...store.UpdateOptionsFunc) error {
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
	result, err := r.db.Exec(
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
	if rows, _ := result.RowsAffected(); rows != 1 { // error ignored, postgres supports RowsAffected()
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

func (r *postgresResourceStore) Delete(_ context.Context, resource core_model.Resource, fs ...store.DeleteOptionsFunc) error {
	opts := store.NewDeleteOptions(fs...)

	statement := `DELETE FROM resources WHERE name=$1 AND type=$2 AND mesh=$3`
	result, err := r.db.Exec(statement, opts.Name, resource.Descriptor().Name, opts.Mesh)
	if err != nil {
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}
	if rows, _ := result.RowsAffected(); rows == 0 { // error ignored, postgres supports RowsAffected()
		return store.ErrorResourceNotFound(resource.Descriptor().Name, opts.Name, opts.Mesh)
	}

	return nil
}

func (r *postgresResourceStore) Get(_ context.Context, resource core_model.Resource, fs ...store.GetOptionsFunc) error {
	opts := store.NewGetOptions(fs...)

	statement := `SELECT spec, version, creation_time, modification_time FROM resources WHERE name=$1 AND mesh=$2 AND type=$3;`
	row := r.db.QueryRow(statement, opts.Name, opts.Mesh, resource.Descriptor().Name)

	var spec string
	var version int
	var creationTime, modificationTime time.Time
	err := row.Scan(&spec, &version, &creationTime, &modificationTime)
	if err == sql.ErrNoRows {
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

func (r *postgresResourceStore) List(_ context.Context, resources core_model.ResourceList, args ...store.ListOptionsFunc) error {
	opts := store.NewListOptions(args...)

	statement := `SELECT name, mesh, spec, version, creation_time, modification_time FROM resources WHERE type=$1`
	var statementArgs []interface{}
	statementArgs = append(statementArgs, resources.GetItemType())
	argsIndex := 1
	if opts.Mesh != "" {
		argsIndex++
		statement += fmt.Sprintf(" AND mesh=$%d", argsIndex)
		statementArgs = append(statementArgs, opts.Mesh)
	}
	statement += " ORDER BY name, mesh"

	rows, err := r.db.Query(statement, statementArgs...)
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

func rowToItem(resources core_model.ResourceList, rows *sql.Rows) (core_model.Resource, error) {
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

func (r *postgresResourceStore) Close() error {
	return r.db.Close()
}

type resourceMetaObject struct {
	Name             string
	Version          string
	Mesh             string
	CreationTime     time.Time
	ModificationTime time.Time
}

var _ core_model.ResourceMeta = &resourceMetaObject{}

func (r *resourceMetaObject) GetName() string {
	return r.Name
}

func (r *resourceMetaObject) GetNameExtensions() core_model.ResourceNameExtensions {
	return core_model.ResourceNameExtensionsUnsupported
}

func (r *resourceMetaObject) GetVersion() string {
	return r.Version
}

func (r *resourceMetaObject) GetMesh() string {
	return r.Mesh
}

func (r *resourceMetaObject) GetCreationTime() time.Time {
	return r.CreationTime
}

func (r *resourceMetaObject) GetModificationTime() time.Time {
	return r.ModificationTime
}

func registerMetrics(metrics core_metrics.Metrics, db *sql.DB) error {
	postgresCurrentConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connections",
		Help: "Current number of postgres store connections",
		ConstLabels: map[string]string{
			"type": "open_connections",
		},
	}, func() float64 {
		return float64(db.Stats().OpenConnections)
	})

	postgresInUseConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connections",
		Help: "Current number of postgres store connections",
		ConstLabels: map[string]string{
			"type": "in_use",
		},
	}, func() float64 {
		return float64(db.Stats().InUse)
	})

	postgresIdleConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connections",
		Help: "Current number of postgres store connections",
		ConstLabels: map[string]string{
			"type": "idle",
		},
	}, func() float64 {
		return float64(db.Stats().Idle)
	})

	postgresMaxOpenConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connections_max",
		Help: "Max postgres store open connections",
	}, func() float64 {
		return float64(db.Stats().MaxOpenConnections)
	})

	postgresWaitConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connection_wait_count",
		Help: "Current waiting postgres store connections",
	}, func() float64 {
		return float64(db.Stats().WaitCount)
	})

	postgresWaitConnectionDurationMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connection_wait_duration",
		Help: "Time Blocked waiting for new connection in seconds",
	}, func() float64 {
		return db.Stats().WaitDuration.Seconds()
	})

	postgresMaxIdleClosedConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connection_closed",
		Help: "Current number of closed postgres store connections",
		ConstLabels: map[string]string{
			"type": "max_idle_conns",
		},
	}, func() float64 {
		return float64(db.Stats().MaxIdleClosed)
	})

	postgresMaxIdleTimeClosedConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connection_closed",
		Help: "Current number of closed postgres store connections",
		ConstLabels: map[string]string{
			"type": "conn_max_idle_time",
		},
	}, func() float64 {
		return float64(db.Stats().MaxIdleTimeClosed)
	})

	postgresMaxLifeTimeClosedConnectionMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "store_postgres_connection_closed",
		Help: "Current number of closed postgres store connections",
		ConstLabels: map[string]string{
			"type": "conn_max_life_time",
		},
	}, func() float64 {
		return float64(db.Stats().MaxLifetimeClosed)
	})

	if err := metrics.
		BulkRegister(postgresCurrentConnectionMetric, postgresInUseConnectionMetric, postgresIdleConnectionMetric,
			postgresMaxOpenConnectionMetric, postgresWaitConnectionMetric, postgresWaitConnectionDurationMetric,
			postgresMaxIdleClosedConnectionMetric, postgresMaxIdleTimeClosedConnectionMetric, postgresMaxLifeTimeClosedConnectionMetric); err != nil {
		return err
	}
	return nil
}
