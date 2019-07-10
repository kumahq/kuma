package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
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

type Config struct {
	Host     string `envconfig:"store_postgres_host" default:"localhost"`
	Port     int    `envconfig:"store_postgres_port" default:"5432"`
	User     string `envconfig:"store_postgres_user" default:"konvoy"`
	Password string `envconfig:"store_postgres_password" default:"konvoy"`
	DbName   string `envconfig:"store_postgres_db_name" default:"konvoy"`
}

func NewStore(config Config) (store.ResourceStore, error) {
	s, err := newPostgresStore(config)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func newPostgresStore(config Config) (*postgresResourceStore, error) {
	db, err := connectToDb(config)
	if err != nil {
		return nil, err
	}

	return &postgresResourceStore{
		db: db,
	}, nil
}

func connectToDb(config Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DbName)
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

func (r *postgresResourceStore) Create(_ context.Context, resource model.Resource, fs ...store.CreateOptionsFunc) error {
	opts := store.NewCreateOptions(fs...)

	bytes, err := json.Marshal(resource.GetSpec())
	if err != nil {
		return errors.Wrap(err, "failed to convert spec to json")
	}

	version := 0
	statement := `INSERT INTO resources VALUES ($1, $2, $3, $4, $5, $6);`
	_, err = r.db.Exec(statement, opts.Name, opts.Namespace, opts.Mesh, resource.GetType(), version, string(bytes))
	if err != nil {
		if strings.Contains(err.Error(), duplicateKeyErrorMsg) {
			return store.ErrorResourceAlreadyExists(resource.GetType(), opts.Namespace, opts.Name, opts.Mesh)
		}
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}

	resource.SetMeta(&resourceMetaObject{
		Name:      opts.Name,
		Namespace: opts.Namespace,
		Mesh:      opts.Mesh,
		Version:   strconv.Itoa(version),
	})
	return nil
}

func (r *postgresResourceStore) Update(_ context.Context, resource model.Resource, fs ...store.UpdateOptionsFunc) error {
	bytes, err := json.Marshal(resource.GetSpec())
	if err != nil {
		return err
	}

	version, err := strconv.Atoi(resource.GetMeta().GetVersion())
	if err != nil {
		return errors.Wrap(err, "failed to convert meta version to int")
	}
	statement := `UPDATE resources SET spec=$1, version=$2 WHERE name=$3 AND namespace=$4 AND mesh=$5 AND type=$6 AND version=$7;`
	result, err := r.db.Exec(
		statement,
		string(bytes),
		version+1,
		resource.GetMeta().GetName(),
		resource.GetMeta().GetNamespace(),
		resource.GetMeta().GetMesh(),
		resource.GetType(),
		version,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to execute query %s", statement)
	}
	if rows, _ := result.RowsAffected(); rows != 1 { // error ignored, postgres supports RowsAffected()
		// todo(jakubdyszkiewicz) throw ErrorResourceConflict when resource is found, but the version does not match
		return store.ErrorResourceNotFound(resource.GetType(), resource.GetMeta().GetNamespace(), resource.GetMeta().GetName(), resource.GetMeta().GetMesh())
	}

	// update resource's meta with new version
	resource.SetMeta(&resourceMetaObject{
		Name:      resource.GetMeta().GetName(),
		Namespace: resource.GetMeta().GetNamespace(),
		Mesh:      resource.GetMeta().GetMesh(),
		Version:   strconv.Itoa(version),
	})

	return nil
}

func (r *postgresResourceStore) Delete(_ context.Context, resource model.Resource, fs ...store.DeleteOptionsFunc) error {
	opts := store.NewDeleteOptions(fs...)

	statement := `DELETE FROM resources WHERE name=$1 AND namespace=$2 AND type=$3`
	var statementArgs []interface{}
	statementArgs = append(statementArgs, opts.Name, opts.Namespace, resource.GetType())
	if opts.Mesh != "" {
		statement += " AND mesh=$4"
		statementArgs = append(statementArgs, opts.Name)
	}
	_, err := r.db.Exec(statement, statementArgs...)
	if err != nil {
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}

	return nil
}

func (r *postgresResourceStore) Get(_ context.Context, resource model.Resource, fs ...store.GetOptionsFunc) error {
	opts := store.NewGetOptions(fs...)

	statement := `SELECT spec, version FROM resources WHERE name=$1 AND namespace=$2 AND mesh=$3 AND type=$4;`
	row := r.db.QueryRow(statement, opts.Name, opts.Namespace, opts.Mesh, resource.GetType())

	var spec string
	var version int
	err := row.Scan(&spec, &version)
	if err == sql.ErrNoRows {
		return store.ErrorResourceNotFound(resource.GetType(), opts.Namespace, opts.Name, opts.Mesh)
	}
	if err != nil {
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}

	err = json.Unmarshal([]byte(spec), resource.GetSpec())
	if err != nil {
		return errors.Wrap(err, "failed to convert json to spec")
	}

	meta := &resourceMetaObject{
		Name:      opts.Name,
		Namespace: opts.Namespace,
		Mesh:      opts.Mesh,
		Version:   strconv.Itoa(version),
	}
	resource.SetMeta(meta)
	return nil
}

func (r *postgresResourceStore) List(_ context.Context, resources model.ResourceList, args ...store.ListOptionsFunc) error {
	opts := store.NewListOptions(args...)

	statement := `SELECT name, spec, version FROM resources WHERE type=?`
	var statementArgs []interface{}
	statementArgs = append(statementArgs, resources.GetItemType())
	if opts.Namespace != "" {
		statement += " AND namespace=?"
		statementArgs = append(statementArgs, opts.Namespace)
	}
	if opts.Mesh != "" {
		statement += " AND mesh=?"
		statementArgs = append(statementArgs, opts.Mesh)
	}
	rows, err := r.db.Query(statement, statementArgs...)
	if err != nil {
		return errors.Wrapf(err, "failed to execute query: %s", statement)
	}
	defer rows.Close()
	for rows.Next() {
		item, err := rowToItem(resources, opts, rows)
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

func rowToItem(resources model.ResourceList, opts *store.ListOptions, rows *sql.Rows) (model.Resource, error) {
	var name, spec string
	var version int
	err := rows.Scan(&name, &spec, &version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve elements from query")
	}

	item := resources.NewItem()
	err = json.Unmarshal([]byte(spec), item.GetSpec())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert json to spec")
	}

	meta := &resourceMetaObject{
		Name:      name,
		Namespace: opts.Namespace,
		Mesh:      opts.Mesh,
		Version:   strconv.Itoa(version),
	}
	item.SetMeta(meta)

	return item, nil
}

func (r *postgresResourceStore) Close() error {
	return r.db.Close()
}

type resourceMetaObject struct {
	Name      string
	Namespace string
	Version   string
	Mesh      string
}

var _ model.ResourceMeta = &resourceMetaObject{}

func (r *resourceMetaObject) GetName() string {
	return r.Name
}

func (r *resourceMetaObject) GetNamespace() string {
	return r.Namespace
}

func (r *resourceMetaObject) GetVersion() string {
	return r.Version
}

func (r *resourceMetaObject) GetMesh() string {
	return r.Mesh
}
