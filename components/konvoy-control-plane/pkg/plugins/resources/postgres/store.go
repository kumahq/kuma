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
)

type postgresResourceStore struct {
	db *sql.DB
}

var _ store.ResourceStore = &postgresResourceStore{}

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DbName   string
}

func NewStore(config Config) (store.ResourceStore, error) {
	s, err := newPostgresStore(config)
	if err != nil {
		return nil, err
	}
	// wrap it to strict resources store so we don't have to duplicate the validations
	return store.NewStrictResourceStore(s), nil
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
	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "cannot connect to DB")
	}

	return db, nil
}

func (r *postgresResourceStore) Create(_ context.Context, resource model.Resource, fs ...store.CreateOptionsFunc) error {
	opts := store.NewCreateOptions(fs...)

	exists, err := r.exists(opts.Name, opts.Namespace, resource.GetType())
	if err != nil {
		return err
	}
	if exists {
		return store.ErrorResourceAlreadyExists(resource.GetType(), opts.Namespace, opts.Name)
	}

	bytes, err := json.Marshal(resource.GetSpec())
	if err != nil {
		return errors.Wrap(err, "failed to convert spec to json")
	}

	version := 0
	statement := `INSERT INTO resources VALUES ($1, $2, $3, $4, $5);`
	_, err = r.db.Exec(statement, opts.Name, opts.Namespace, resource.GetType(), version, string(bytes))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to execute query: %s", statement))
	}

	resource.SetMeta(&resourceMetaObject{
		Name:      opts.Name,
		Namespace: opts.Namespace,
		Version:   strconv.Itoa(version),
	})
	return nil
}

func (r *postgresResourceStore) exists(name string, namespace string, resourceType model.ResourceType) (bool, error) {
	statement := `SELECT count(*) FROM resources WHERE name=$1 AND namespace=$2 AND type=$3`
	row := r.db.QueryRow(statement, name, namespace, resourceType)

	var elements int
	err := row.Scan(&elements)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("failed to execute query: %s", statement))
	}

	return elements > 0, nil
}

func (r *postgresResourceStore) Update(_ context.Context, resource model.Resource, fs ...store.UpdateOptionsFunc) error {
	_ = store.NewUpdateOptions(fs...)

	exists, err := r.exists(resource.GetMeta().GetName(), resource.GetMeta().GetNamespace(), resource.GetType())
	if err != nil {
		return err
	}
	if !exists {
		return store.ErrorResourceNotFound(resource.GetType(), resource.GetMeta().GetNamespace(), resource.GetMeta().GetName())
	}

	bytes, err := json.Marshal(resource.GetSpec())
	if err != nil {
		return err
	}

	version, err := strconv.Atoi(resource.GetMeta().GetVersion())
	if err != nil {
		return errors.Wrap(err, "failed to convert meta version to int")
	}
	statement := `UPDATE resources SET spec=$1, version=$2 WHERE name=$3 AND namespace=$4 AND type=$5 AND version=$6;`
	_, err = r.db.Exec(
		statement,
		string(bytes),
		version + 1,
		resource.GetMeta().GetName(),
		resource.GetMeta().GetNamespace(),
		resource.GetType(),
		version,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to execute query %s", statement))
	}

	// update resource's meta with new version
	resource.SetMeta(&resourceMetaObject{
		Name:      resource.GetMeta().GetName(),
		Namespace: resource.GetMeta().GetNamespace(),
		Version:   strconv.Itoa(version),
	})

	return nil
}

func (r *postgresResourceStore) Delete(_ context.Context, resource model.Resource, fs ...store.DeleteOptionsFunc) error {
	opts := store.NewDeleteOptions(fs...)

	version, err := strconv.Atoi(opts.Version)
	if err != nil {
		return errors.Wrap(err, "failed to convert meta version to int")
	}

	statement := `DELETE FROM resources WHERE name=$1 AND namespace=$2 AND type=$3 AND version=$4;`
	_, err = r.db.Exec(statement, opts.Name, opts.Namespace, resource.GetType(), version)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to execute query: %s", statement))
	}

	return nil
}

func (r *postgresResourceStore) Get(_ context.Context, resource model.Resource, fs ...store.GetOptionsFunc) error {
	opts := store.NewGetOptions(fs...)

	statement := `SELECT spec, version FROM resources WHERE name=$1 AND namespace=$2 AND type=$3;`
	row := r.db.QueryRow(statement, opts.Name, opts.Namespace, resource.GetType())

	var spec string
	var version int
	err := row.Scan(&spec, &version)
	if err == sql.ErrNoRows {
		return store.ErrorResourceNotFound(resource.GetType(), opts.Namespace, opts.Name)
	}
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to execute query: %s", statement))
	}

	err = json.Unmarshal([]byte(spec), resource.GetSpec())
	if err != nil {
		return errors.Wrap(err, "failed to convert json to spec")
	}

	meta := &resourceMetaObject{
		Name:      opts.Name,
		Namespace: opts.Namespace,
		Version:   strconv.Itoa(version),
	}
	resource.SetMeta(meta)
	return nil
}

func (r *postgresResourceStore) List(_ context.Context, resources model.ResourceList, args ...store.ListOptionsFunc) error {
	opts := store.NewListOptions(args...)

	statement := `SELECT name, spec, version FROM resources WHERE namespace=$1 AND type=$2;`
	rows, err := r.db.Query(statement, opts.Namespace, resources.GetItemType())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to execute query: %s", statement))
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
		Version:   strconv.Itoa(version),
	}
	item.SetMeta(meta)

	return item, nil
}

func (r *postgresResourceStore) Close() error {
	return r.db.Close()
}

func (r *postgresResourceStore) deleteAll() error {
	_, err := r.db.Exec("DELETE FROM resources")
	return err
}

type resourceMetaObject struct {
	Name      string
	Namespace string
	Version   string
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

type resourceObject struct {
	Type model.ResourceType
	Meta model.ResourceMeta
	Spec model.ResourceSpec
}

var _ model.Resource = &resourceObject{}

func (r *resourceObject) GetType() model.ResourceType {
	return r.Type
}

func (r *resourceObject) GetMeta() model.ResourceMeta {
	return r.Meta
}

func (r *resourceObject) SetMeta(meta model.ResourceMeta) {
	r.Meta = meta
}

func (r *resourceObject) GetSpec() model.ResourceSpec {
	return r.Spec
}
