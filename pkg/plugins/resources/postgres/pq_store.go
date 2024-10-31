package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/maps"

	config "github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	common_postgres "github.com/kumahq/kuma/pkg/plugins/common/postgres"
)

type postgresResourceStore struct {
	db                   *sql.DB
	maxListQueryElements uint32
}

var _ store.ResourceStore = &postgresResourceStore{}

func NewPqStore(metrics core_metrics.Metrics, config config.PostgresStoreConfig) (store.ResourceStore, error) {
	db, err := common_postgres.ConnectToDb(config)
	if err != nil {
		return nil, err
	}

	if err := registerPqMetrics(metrics, db); err != nil {
		return nil, errors.Wrapf(err, "could not register DB metrics")
	}

	return &postgresResourceStore{
		db:                   db,
		maxListQueryElements: config.MaxListQueryElements,
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
	statement := `INSERT INTO resources (name, mesh, type, version, spec, creation_time, modification_time, owner_name, owner_mesh, owner_type, labels) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`

	labels, err := prepareLabels(opts.Labels)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(statement, opts.Name, opts.Mesh, resource.Descriptor().Name, version, string(bytes),
		opts.CreationTime.UTC(), opts.CreationTime.UTC(), ownerName, ownerMesh, ownerType, labels)
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
		Labels:           maps.Clone(opts.Labels),
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

	updateLabels := resource.GetMeta().GetLabels()
	if opts.ModifyLabels {
		updateLabels = opts.Labels
	}
	labels, err := prepareLabels(updateLabels)
	if err != nil {
		return err
	}

	statement := `UPDATE resources SET spec=$1, version=$2, modification_time=$3, labels=$4 WHERE name=$5 AND mesh=$6 AND type=$7 AND version=$8;`
	result, err := r.db.Exec(
		statement,
		string(bytes),
		newVersion,
		opts.ModificationTime.UTC(),
		labels,
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
		Labels:           maps.Clone(opts.Labels),
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

	statement := `SELECT spec, version, creation_time, modification_time, labels FROM resources WHERE name=$1 AND mesh=$2 AND type=$3;`
	row := r.db.QueryRow(statement, opts.Name, opts.Mesh, resource.Descriptor().Name)

	var spec string
	var version int
	var creationTime, modificationTime time.Time
	var labels string
	err := row.Scan(&spec, &version, &creationTime, &modificationTime, &labels)
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
		Labels:           map[string]string{},
	}
	if err := json.Unmarshal([]byte(labels), &meta.Labels); err != nil {
		return errors.Wrap(err, "failed to convert json to labels")
	}
	resource.SetMeta(meta)

	if opts.Version != "" && resource.GetMeta().GetVersion() != opts.Version {
		return store.ErrorResourceConflict(resource.Descriptor().Name, opts.Name, opts.Mesh)
	}
	return nil
}

func (r *postgresResourceStore) List(_ context.Context, resources core_model.ResourceList, args ...store.ListOptionsFunc) error {
	opts := store.NewListOptions(args...)

	statement := `SELECT name, mesh, spec, version, creation_time, modification_time, labels FROM resources WHERE type=$1`
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

	rows, err := r.db.Query(statement, statementArgs...)
	if err != nil {
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		item, err := pqRowToItem(resources, rows)
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

func pqRowToItem(resources core_model.ResourceList, rows *sql.Rows) (core_model.Resource, error) {
	var name, mesh, spec string
	var version int
	var creationTime, modificationTime time.Time
	var labels string
	if err := rows.Scan(&name, &mesh, &spec, &version, &creationTime, &modificationTime, &labels); err != nil {
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
		Labels:           map[string]string{},
	}
	if err := json.Unmarshal([]byte(labels), &meta.Labels); err != nil {
		return nil, errors.Wrap(err, "failed to convert json to labels")
	}
	item.SetMeta(meta)

	return item, nil
}

func (r *postgresResourceStore) Close() error {
	return r.db.Close()
}

func registerPqMetrics(metrics core_metrics.Metrics, db *sql.DB) error {
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
