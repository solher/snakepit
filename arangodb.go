package snakepit

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/solher/arangolite"
)

const (
	colTypeDoc  = 2
	colTypeEdge = 3
)

type Seed interface{}

type ArangoDB struct {
	db                     *arangolite.DB
	localSeed, distantSeed Seed
	URL, Name              string
	User, UserPassword     string
}

func NewArangoDB(localSeed, distantSeed Seed) *ArangoDB {
	return &ArangoDB{
		db:          arangolite.New(),
		localSeed:   localSeed,
		distantSeed: distantSeed,
	}
}

func (d *ArangoDB) Connect(url, name, user, userPassword string) *ArangoDB {
	d.URL = url
	d.Name = name
	d.User = user
	d.UserPassword = userPassword

	d.db.Connect(url, name, user, userPassword)

	return d
}

func (d *ArangoDB) LoggerOptions(enabled, printQuery, printResult bool) *ArangoDB {
	d.db.LoggerOptions(enabled, printQuery, printResult)
	return d
}

func (d *ArangoDB) Run(q arangolite.Runnable) ([]byte, error) {
	return d.db.Run(q)
}

func (d *ArangoDB) RunAsync(q arangolite.Runnable) (*arangolite.Result, error) {
	return d.db.RunAsync(q)
}

func (d *ArangoDB) Create(rootUser, rootPassword string) error {
	d.db.SwitchDatabase("_system").SwitchUser(rootUser, rootPassword)
	defer func() { d.db.SwitchDatabase(d.Name).SwitchUser(d.User, d.UserPassword) }()

	_, err := d.Run(&arangolite.CreateDatabase{
		Name: d.Name,
		Users: []map[string]interface{}{
			{"username": rootUser, "passwd": rootPassword},
			{"username": d.User, "passwd": d.UserPassword},
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func (d *ArangoDB) Migrate() error {
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

		_, err := d.Run(&arangolite.CreateCollection{
			Name: colName,
			Type: colType,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *ArangoDB) Drop(rootUser, rootPassword string) error {
	d.db.SwitchDatabase("_system").SwitchUser(rootUser, rootPassword)
	defer func() { d.db.SwitchDatabase(d.Name).SwitchUser(d.User, d.UserPassword) }()

	_, err := d.Run(&arangolite.DropDatabase{Name: d.Name})
	if err != nil {
		return err
	}

	return nil
}

func (d *ArangoDB) LoadDistantSeed() error {
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

		if localField.Len() == 0 {
			continue
		}

		arrayElem := localField.Type().Elem()

		if arrayElem.Kind() != reflect.Struct {
			return errors.New("invalid seed field type: not a struct")
		}

		colName := local.Type().Field(i).Name
		colName = strings.ToLower(colName[0:1]) + colName[1:]

		// util.Dump(string(m))

		// q := arangolite.NewQuery(`
		// 			FOR x IN %s
		// 				FOR y IN %s
		// 				LET merged = MERGE(x, y)
		// 				FILTER MATCHES(x, merged)
		// 				RETURN x
		// 		`, colName, m)
		var q *arangolite.Query

		if local.Type().Field(i).Tag.Get("check") == "keyOnly" {
			q = arangolite.NewQuery(`
				FOR x IN @@colName
					FOR y IN @seed
					FILTER x._key == y._key
					RETURN DISTINCT x
			`).Bind("seed", localField.Interface()).Bind("@colName", colName)
		} else {
			q = arangolite.NewQuery(`
				FOR x IN @@colName
				LET i = UNSET(x,"_id","_rev")
					FOR y IN @seed
					LET j = UNSET(y,"_id","_rev")
					LET merged = MERGE(i,j)
					FILTER MATCHES(i, merged)
					RETURN DISTINCT x
			`).Bind("seed", localField.Interface()).Bind("@colName", colName)
		}

		r, err := d.Run(q)
		if err != nil {
			return err
		}

		json.Unmarshal(r, distantField.Addr().Interface())

		if distantField.Len() < localField.Len() {
			return fmt.Errorf("seeds not synchronized: %s", colName)
		}
	}

	return nil
}

func (d *ArangoDB) SyncSeeds() error {
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

		var q *arangolite.Query

		if local.Type().Field(i).Tag.Get("seed") == "forceUpdate" {
			q = arangolite.NewQuery(`
				FOR x IN @seed
				FILTER x._key != "" && x._key != NULL
				UPSERT { '_key': x._key }
                INSERT x
				REPLACE x IN @@colName
			`).Bind("seed", field.Interface()).Bind("@colName", colName)
		} else {
			q = arangolite.NewQuery(`
				FOR x IN @seed
				FILTER x._key != "" && x._key != NULL
                INSERT x IN @@colName OPTIONS { ignoreErrors: true }
			`).Bind("seed", field.Interface()).Bind("@colName", colName)
		}

		if _, err := d.Run(q); err != nil {
			return err
		}
	}

	if err := d.LoadDistantSeed(); err != nil {
		return err
	}

	return nil
}
