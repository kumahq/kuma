package postgres

import (
	"context"
	"database/sql"
	"fmt"
	config "github.com/Kong/kuma/pkg/config/plugins/resources/postgres"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/util/proto"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

const duplicateKeyErrorMsg = "duplicate key value violates unique constraint"

type postgresResourceStore struct {
	db *sql.DB
}

var _ store.ResourceStore = &postgresResourceStore{}

func NewStore(config config.PostgresStoreConfig) (store.ResourceStore, error) {
	db, err := connectToDb(config)
	if err != nil {
		return nil, err
	}

	return &postgresResourceStore{
		db: db,
	}, nil
}

func connectToDb(cfg config.PostgresStoreConfig) (*sql.DB, error) {
	mode, err := postgresMode(cfg.TLS.Mode)
	if err != nil {
		return nil, err
	}
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s connect_timeout=%d sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DbName, cfg.ConnectionTimeout, mode, cfg.TLS.CertPath, cfg.TLS.KeyPath, cfg.TLS.CAPath)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create connection to DB")
	}

	// check connection to DB, Open() does not check it.
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "cannot connect to DB")
	}

	return db, nil
}

func postgresMode(mode config.TLSMode) (string, error) {
	switch mode {
	case config.Disable:
		return "disable", nil
	case config.VerifyNone:
		return "require", nil
	case config.VerifyCa:
		return "verify-ca", nil
	case config.VerifyFull:
		return "verify-full", nil
	default:
		return "", errors.Errorf("could not translate mode %q to postgres mode", mode)
	}
}

func (r *postgresResourceStore) Create(_ context.Context, resource model.Resource, fs ...store.CreateOptionsFunc) error {
	opts := store.NewCreateOptions(fs...)

	bytes, err := proto.ToJSON(resource.GetSpec())
	if err != nil {
		return errors.Wrap(err, "failed to convert spec to json")
	}

	version := 0
	statement := `INSERT INTO resources VALUES ($1, $2, $3, $4, $5, $6);`
	_, err = r.db.Exec(statement, opts.Name, "", opts.Mesh, resource.GetType(), version, string(bytes)) // todo(jakubdyszkiewicz) solve db migration
	if err != nil {
		if strings.Contains(err.Error(), duplicateKeyErrorMsg) {
			return store.ErrorResourceAlreadyExists(resource.GetType(), opts.Name, opts.Mesh)
		}
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}

	resource.SetMeta(&resourceMetaObject{
		Name:    opts.Name,
		Mesh:    opts.Mesh,
		Version: strconv.Itoa(version),
	})
	return nil
}

func (r *postgresResourceStore) Update(_ context.Context, resource model.Resource, fs ...store.UpdateOptionsFunc) error {
	bytes, err := proto.ToJSON(resource.GetSpec())
	if err != nil {
		return err
	}

	version, err := strconv.Atoi(resource.GetMeta().GetVersion())
	if err != nil {
		return errors.Wrap(err, "failed to convert meta version to int")
	}
	statement := `UPDATE resources SET spec=$1, version=$2 WHERE name=$3 AND mesh=$4 AND type=$5 AND version=$6;`
	result, err := r.db.Exec(
		statement,
		string(bytes),
		version+1,
		resource.GetMeta().GetName(),
		resource.GetMeta().GetMesh(),
		resource.GetType(),
		version,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to execute query %s", statement)
	}
	if rows, _ := result.RowsAffected(); rows != 1 { // error ignored, postgres supports RowsAffected()
		return store.ErrorResourceConflict(resource.GetType(), resource.GetMeta().GetName(), resource.GetMeta().GetMesh())
	}

	// update resource's meta with new version
	resource.SetMeta(&resourceMetaObject{
		Name:    resource.GetMeta().GetName(),
		Mesh:    resource.GetMeta().GetMesh(),
		Version: strconv.Itoa(version),
	})

	return nil
}

func (r *postgresResourceStore) Delete(_ context.Context, resource model.Resource, fs ...store.DeleteOptionsFunc) error {
	opts := store.NewDeleteOptions(fs...)

	statement := `DELETE FROM resources WHERE name=$1 AND type=$2 AND mesh=$3`
	result, err := r.db.Exec(statement, opts.Name, resource.GetType(), opts.Mesh)
	if err != nil {
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}
	if rows, _ := result.RowsAffected(); rows == 0 { // error ignored, postgres supports RowsAffected()
		return store.ErrorResourceNotFound(resource.GetType(), opts.Name, opts.Mesh)
	}

	return nil
}

func (r *postgresResourceStore) Get(_ context.Context, resource model.Resource, fs ...store.GetOptionsFunc) error {
	opts := store.NewGetOptions(fs...)

	statement := `SELECT spec, version FROM resources WHERE name=$1 AND mesh=$2 AND type=$3`
	if opts.Version != "" {
		statement += " AND version=$4"
	}
	var statementArgs []interface{}
	statementArgs = append(statementArgs, opts.Name, opts.Mesh, resource.GetType())
	if opts.Version != "" {
		v, err := strconv.Atoi(opts.Version)
		if err != nil {
			return errors.Wrap(err, "invalid version. Postgres version should be an integer")
		}
		statementArgs = append(statementArgs, v)
	}
	row := r.db.QueryRow(statement, statementArgs...)

	var spec string
	var version int
	err := row.Scan(&spec, &version)
	if err == sql.ErrNoRows {
		return store.ErrorResourceNotFound(resource.GetType(), opts.Name, opts.Mesh)
	}
	if err != nil {
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}

	if err := proto.FromJSON([]byte(spec), resource.GetSpec()); err != nil {
		return errors.Wrap(err, "failed to convert json to spec")
	}

	meta := &resourceMetaObject{
		Name:    opts.Name,
		Mesh:    opts.Mesh,
		Version: strconv.Itoa(version),
	}
	resource.SetMeta(meta)
	return nil
}

func (r *postgresResourceStore) List(_ context.Context, resources model.ResourceList, args ...store.ListOptionsFunc) error {
	opts := store.NewListOptions(args...)

	statement := `SELECT name, mesh, spec, version FROM resources WHERE type=$1`
	var statementArgs []interface{}
	statementArgs = append(statementArgs, resources.GetItemType())
	argsIndex := 1
	if opts.Mesh != "" {
		argsIndex++
		statement += fmt.Sprintf(" AND mesh=$%d", argsIndex)
		statementArgs = append(statementArgs, opts.Mesh)
	}
	rows, err := r.db.Query(statement, statementArgs...)
	if err != nil {
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}
	defer rows.Close()
	for rows.Next() {
		item, err := rowToItem(resources, rows)
		if err != nil {
			return err
		}
		err = resources.AddItem(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func rowToItem(resources model.ResourceList, rows *sql.Rows) (model.Resource, error) {
	var name, mesh, spec string
	var version int
	if err := rows.Scan(&name, &mesh, &spec, &version); err != nil {
		return nil, errors.Wrap(err, "failed to retrieve elements from query")
	}

	item := resources.NewItem()
	if err := proto.FromJSON([]byte(spec), item.GetSpec()); err != nil {
		return nil, errors.Wrap(err, "failed to convert json to spec")
	}

	meta := &resourceMetaObject{
		Name:    name,
		Mesh:    mesh,
		Version: strconv.Itoa(version),
	}
	item.SetMeta(meta)

	return item, nil
}

func (r *postgresResourceStore) Close() error {
	return r.db.Close()
}

type resourceMetaObject struct {
	Name    string
	Version string
	Mesh    string
}

var _ model.ResourceMeta = &resourceMetaObject{}

func (r *resourceMetaObject) GetName() string {
	return r.Name
}

func (r *resourceMetaObject) GetVersion() string {
	return r.Version
}

func (r *resourceMetaObject) GetMesh() string {
	return r.Mesh
}
