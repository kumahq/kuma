package resources_postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/client"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/model"
	_ "github.com/lib/pq"
)

type ResourceClient struct {
	db *sql.DB
}

var _ client.ResourceClient = &ResourceClient{}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DbName   string
}

func NewResourceClient(config PostgresConfig) (client.ResourceClient, error) {
	db, err := connectToDb(config)
	if err != nil {
		return nil, err
	}

	c := &ResourceClient{
		db: db,
	}
	return client.NewStrictResourceClient(c), nil
}

func connectToDb(config PostgresConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Check connection to DB, Open does not check it.
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

//TODO(jakubdyszkiewicz) close

func (r *ResourceClient) Create(_ context.Context, resource model.Resource, args ...client.CreateOptionsFunc) error {
	opts := client.NewCreateOptions(args...)

	exists, err := r.exists(opts.Name, opts.Namespace, resource.GetType())
	if err != nil {
		return err
	}
	if exists {
		return client.ErrorResourceAlreadyExists(resource.GetType(), opts.Namespace, opts.Name)
	}

	bytes, err := json.Marshal(resource.GetSpec())
	if err != nil {
		return err
	}

	statement := `INSERT INTO resources VALUES ($1, $2, $3, $4);`
	_, err = r.db.Exec(statement, opts.Name, opts.Namespace, resource.GetType(), string(bytes))
	if err != nil {
		return err
	}

	resource.SetMeta(&resourceMetaObject{
		Name:      opts.Name,
		Namespace: opts.Namespace,
	})

	return nil
}

func (r *ResourceClient) Update(context context.Context, resource model.Resource, args ...client.UpdateOptionsFunc) error {
	_ = client.NewUpdateOptions(args...)

	exists, err := r.exists(resource.GetMeta().GetName(), resource.GetMeta().GetNamespace(), resource.GetType())
	if err != nil {
		return err
	}
	if !exists {
		return client.ErrorResourceNotFound(resource.GetType(), resource.GetMeta().GetNamespace(), resource.GetMeta().GetName())
	}

	bytes, err := json.Marshal(resource.GetSpec())
	if err != nil {
		return err
	}

	statement := `UPDATE resources SET spec=$1 WHERE name=$2 AND namespace=$3 AND type=$4;`
	_, err = r.db.Exec(
		statement,
		string(bytes),
		resource.GetMeta().GetName(),
		resource.GetMeta().GetNamespace(),
		resource.GetType(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ResourceClient) Delete(context context.Context, resource model.Resource, args ...client.DeleteOptionsFunc) error {
	opts := client.NewDeleteOptions(args...)

	statement := `DELETE FROM resources WHERE name=$1 AND namespace=$2 AND type=$3;`
	_, err := r.db.Exec(statement, opts.Name, opts.Namespace, resource.GetType())
	if err != nil {
		return err
	}

	return nil
}

func (r *ResourceClient) Get(context context.Context, resource model.Resource, args ...client.GetOptionsFunc) error {
	opts := client.NewGetOptions(args...)

	statement := `SELECT spec FROM resources WHERE name=$1 AND namespace=$2 AND type=$3;`
	row := r.db.QueryRow(statement, opts.Name, opts.Namespace, resource.GetType())

	var spec string
	err := row.Scan(&spec)
	if err == sql.ErrNoRows {
		return client.ErrorResourceNotFound(resource.GetType(), opts.Namespace, opts.Name)
	}
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(spec), resource.GetSpec())
	if err != nil {
		return err
	}

	meta := &resourceMetaObject{
		Name:      opts.Name,
		Namespace: opts.Namespace,
		Version:   "v1",
	}
	resource.SetMeta(meta)
	return nil
}

func (r *ResourceClient) List(context context.Context, resources model.ResourceList, args ...client.ListOptionsFunc) error {
	opts := client.NewListOptions(args...)

	statement := `SELECT name, spec FROM resources WHERE namespace=$1 AND type=$2;`
	rows, err := r.db.Query(statement, opts.Namespace, resources.GetItemType())
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var spec string
		err = rows.Scan(&name, &spec)
		if err != nil {
			return err
		}

		item := resources.NewItem()
		err = json.Unmarshal([]byte(spec), item.GetSpec())
		if err != nil {
			return err
		}

		meta := &resourceMetaObject{
			Name:      name,
			Namespace: opts.Namespace,
			Version:   "v1",
		}
		item.SetMeta(meta)

		err = resources.AddItem(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ResourceClient) exists(name string, namespace string, resourceType model.ResourceType) (bool, error) {
	statement := `SELECT count(*) FROM resources WHERE name=$1 AND namespace=$2 AND type=$3`
	row := r.db.QueryRow(statement, name, namespace, resourceType)

	var elements int
	err := row.Scan(&elements)
	if err != nil {
		return false, err
	}

	if elements > 0 {
		return true, nil
	} else {
		return false, nil
	}
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
