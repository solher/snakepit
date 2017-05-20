package snakepit

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/solher/arangolite"
	"github.com/solher/arangolite/requests"
)

const (
	colTypeDoc  = 2
	colTypeEdge = 3
)

type Seed interface{}

type ArangoDBManager struct {
	db                     *arangolite.Database
	localSeed, distantSeed Seed
	URL, Name              string
	User, UserPassword     string
}

func NewArangoDBManager(localSeed, distantSeed Seed) *ArangoDBManager {
	return &ArangoDBManager{
		db:          arangolite.NewDatabase(),
		localSeed:   localSeed,
		distantSeed: distantSeed,
	}
}

func (d *ArangoDBManager) Connect(ctx context.Context, url, name, user, userPassword string) (*ArangoDBManager, error) {
	d.URL = url
	d.Name = name
	d.User = user
	d.UserPassword = userPassword

	d.db.Options(
		arangolite.OptBasicAuth(user, userPassword),
		arangolite.OptDatabaseName(name),
		arangolite.OptEndpoint(url),
	)

	if err := d.db.Connect(ctx); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *ArangoDBManager) Options(opts ...arangolite.Option) {
	d.db.Options(opts...)
}

func (d *ArangoDBManager) Run(ctx context.Context, v interface{}, q arangolite.Runnable) error {
	return d.db.Run(ctx, v, q)
}

func (d *ArangoDBManager) Create(ctx context.Context, rootUser, rootPassword string) error {
	d.db.Options(
		arangolite.OptBasicAuth(rootUser, rootPassword),
		arangolite.OptDatabaseName("_system"),
	)
	defer func() {
		d.db.Options(
			arangolite.OptBasicAuth(d.User, d.UserPassword),
			arangolite.OptDatabaseName(d.Name),
		)
	}()

	err := d.db.Run(ctx, nil, &requests.CreateDatabase{
		Name: d.Name,
		Users: []map[string]interface{}{
			{"username": rootUser, "passwd": rootPassword},
			{"username": d.User, "passwd": d.UserPassword},
		},
	})

	if err != nil && !strings.Contains(err.Error(), "duplicate name") {
		return err
	}

	return nil
}

func (d *ArangoDBManager) Migrate(ctx context.Context) error {
	local := reflect.ValueOf(d.localSeed)

	if local.Kind() != reflect.Ptr {
		return errors.New("invalid seed type: not a pointer")
	}

	local = local.Elem()

	if local.Kind() != reflect.Struct {
		return errors.New("invalid seed type: not a struct")
	}

	for i := 0; i < local.NumField(); i++ {
		field := local.Field(i)

		if field.Kind() != reflect.Slice {
			continue
		}

		arrayElem := field.Type().Elem()

		if arrayElem.Kind() != reflect.Struct {
			return errors.New("invalid seed field type: not a struct")
		}

		colName := local.Type().Field(i).Name
		colName = strings.ToLower(colName[0:1]) + colName[1:]
		colType := colTypeDoc

		// if _, ok := local.Field(i).Type().Elem().FieldByName("Edge"); ok {
		if _, ok := arrayElem.FieldByName("From"); ok {
			colType = colTypeEdge
		}

		err := d.db.Run(ctx, nil, &requests.CreateCollection{
			Name: colName,
			Type: colType,
		})
		if err != nil && !strings.Contains(err.Error(), "duplicate name") {
			return err
		}
	}

	return nil
}

func (d *ArangoDBManager) Drop(ctx context.Context, rootUser, rootPassword string) error {
	d.db.Options(
		arangolite.OptBasicAuth(rootUser, rootPassword),
		arangolite.OptDatabaseName("_system"),
	)
	defer func() {
		d.db.Options(
			arangolite.OptBasicAuth(d.User, d.UserPassword),
			arangolite.OptDatabaseName(d.Name),
		)
	}()

	if err := d.db.Run(ctx, nil, &requests.DropDatabase{Name: d.Name}); err != nil {
		return err
	}

	return nil
}

func (d *ArangoDBManager) LoadDistantSeed(ctx context.Context) error {
	local := reflect.ValueOf(d.localSeed)
	distant := reflect.ValueOf(d.distantSeed)

	if local.Kind() != reflect.Ptr || distant.Kind() != reflect.Ptr {
		return errors.New("invalid seed type: not a pointer")
	}

	local = local.Elem()
	distant = distant.Elem()

	if local.Kind() != reflect.Struct || distant.Kind() != reflect.Struct {
		return errors.New("invalid seed type: not a struct")
	}

	for i := 0; i < local.NumField(); i++ {
		localField := local.Field(i)
		distantField := distant.Field(i)

		if localField.Kind() != reflect.Slice {
			continue
		}

		arrayElem := localField.Type().Elem()

		if arrayElem.Kind() != reflect.Struct {
			return errors.New("invalid seed field type: not a struct")
		}

		colName := local.Type().Field(i).Name
		colName = strings.ToLower(colName[0:1]) + colName[1:]

		var q *requests.AQL

		if local.Type().Field(i).Tag.Get("check") == "keyOnly" {
			q = requests.NewAQL(`
				FOR x IN @@colName
					FOR y IN @seed != null ? @seed : []
					FILTER x._key == y._key
					RETURN DISTINCT x
			`).Bind("seed", localField.Interface()).Bind("@colName", colName)
		} else {
			q = requests.NewAQL(`
				FOR x IN @@colName
				LET i = UNSET(x,"_id","_rev")
					FOR y IN @seed != null ? @seed : []
					LET j = UNSET(y,"_id","_rev")
					LET merged = MERGE(i,j)
					FILTER MATCHES(i, merged)
					RETURN DISTINCT x
			`).Bind("seed", localField.Interface()).Bind("@colName", colName)
		}

		if err := d.db.Run(ctx, distantField.Addr().Interface(), q); err != nil {
			return err
		}

		if distantField.Len() < localField.Len() {
			return fmt.Errorf("seeds not synchronized: %s", colName)
		}
	}

	return nil
}

func (d *ArangoDBManager) SyncSeeds(ctx context.Context) error {
	local := reflect.ValueOf(d.localSeed)

	if local.Kind() != reflect.Ptr {
		return errors.New("invalid seed type: not a pointer")
	}

	local = local.Elem()

	if local.Kind() != reflect.Struct {
		return errors.New("invalid seed type: not a struct")
	}

	for i := 0; i < local.NumField(); i++ {
		field := local.Field(i)

		if field.Kind() != reflect.Slice {
			continue
		}

		if field.Len() == 0 {
			continue
		}

		arrayElem := field.Type().Elem()

		if arrayElem.Kind() != reflect.Struct {
			return errors.New("invalid seed field type: not a struct")
		}

		colName := local.Type().Field(i).Name
		colName = strings.ToLower(colName[0:1]) + colName[1:]

		var q *requests.AQL

		if local.Type().Field(i).Tag.Get("seed") == "forceUpdate" {
			q = requests.NewAQL(`
				FOR x IN @seed
				FILTER x._key != "" && x._key != NULL
				UPSERT { '_key': x._key }
                INSERT x
				REPLACE x IN @@colName
			`).Bind("seed", field.Interface()).Bind("@colName", colName)
		} else {
			q = requests.NewAQL(`
				FOR x IN @seed
				FILTER x._key != "" && x._key != NULL
                INSERT x IN @@colName OPTIONS { ignoreErrors: true }
			`).Bind("seed", field.Interface()).Bind("@colName", colName)
		}

		if err := d.db.Run(ctx, nil, q); err != nil {
			return err
		}
	}

	if err := d.LoadDistantSeed(ctx); err != nil {
		return err
	}

	return nil
}
