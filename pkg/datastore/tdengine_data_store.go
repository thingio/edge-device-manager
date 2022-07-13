package datastore

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/taosdata/driver-go/v2/taosSql"
	"github.com/thingio/edge-device-manager/pkg/datastore/query"
	"github.com/thingio/edge-device-std/config"
	"github.com/thingio/edge-device-std/errors"
	"github.com/thingio/edge-device-std/models"
	"io"
	"strings"
)

func NewTDengineDataStore(cfg *config.TDengineOptions) (*TDengineDataStore, error) {
	dsn := fmt.Sprintf("%s:%s@/%s(%s:%d)/",
		cfg.Username, cfg.Password, cfg.Schema, cfg.Host, cfg.Port)
	native, err := sql.Open("taosSql", dsn)
	if err != nil {
		return nil, errors.DataStore.Error("failed to connect TDengine, got %s", err)
	}

	return &TDengineDataStore{
		cfg: cfg,
		c:   native,
	}, nil
}

type TDengineDataStore struct {
	dataStore
	cfg *config.TDengineOptions
	c   *sql.DB
}

func (ds *TDengineDataStore) Connect() error {
	if err := ds.c.Ping(); err != nil {
		return err
	}
	return nil
}
func (ds *TDengineDataStore) Close() {
	_ = ds.c.Close()
}

func (ds *TDengineDataStore) CreateDB() error {
	query := fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s`, ds.database)
	if ds.cfg.Keep != 0 {
		query += fmt.Sprintf(` KEEP %d`, ds.cfg.Keep)
	}
	if ds.cfg.Days != 0 {
		query += fmt.Sprintf(` DAYS %d`, ds.cfg.Days)
	}
	if ds.cfg.Blocks != 0 {
		query += fmt.Sprintf(` BLOCKS %d`, ds.cfg.Blocks)
	}
	if ds.cfg.Precision != "" {
		query += fmt.Sprintf(` PRECISION '%s'`, ds.cfg.Precision)
	}
	if ds.cfg.Update != 0 {
		query += fmt.Sprintf(` UPDATE %d`, ds.cfg.Update)
	}
	query += `;`

	stmt, err := ds.c.Prepare(query)
	if err != nil {
		return err
	}

	// FIXME The problem of 'Unable to establish connection' occurs frequently.
	_, _ = stmt.Exec()
	//_, err = stmt.Exec()
	//if err != nil {
	//	return err
	//}
	return nil
}
func (ds *TDengineDataStore) UseDB() error {
	query := fmt.Sprintf(`USE %s`, ds.database)
	stmt, err := ds.c.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return nil
}
func (ds *TDengineDataStore) CreateTable(schema *DeviceDataSchema) error {
	if err := ds.createSTable(schema); err != nil {
		return err
	}
	return nil
}
func (ds *TDengineDataStore) createSTable(schema *DeviceDataSchema) error {
	query := fmt.Sprintf(`CREATE STABLE IF NOT EXISTS %s`, schema.AggregatedTableName())
	// COLUMNS
	fields := `ts TIMESTAMP`
	for _, field := range schema.Fields() {
		fields += fmt.Sprintf(`, %s %s`, field.Name, typeOf(field.Type))
	}
	query += fmt.Sprintf(` (%s)`, fields)
	// TAG COLUMNS
	if len(schema.Tags()) != 0 {
		tags := ``
		for i, tag := range schema.Tags() {
			if i > 0 {
				tags += `, `
			}
			tags += fmt.Sprintf(`%s %s`, tag.Name, typeOf(tag.Type))
		}
		query += fmt.Sprintf(" TAGS(%s)", tags)
	}
	query += `;`

	stmt, err := ds.c.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return nil
}
func (ds *TDengineDataStore) Write(records ...*DeviceDataRecord) error {
	query := fmt.Sprintf(`INSERT INTO`)
	for _, record := range records {
		// TAG COLUMNS
		query += fmt.Sprintf(` %s USING %s`, record.TableName(), record.AggregatedTableName())
		var tagKeys []string
		var tagValues []string
		for tk, tv := range record.Tags() {
			tagKeys = append(tagKeys, tk)
			tagValues = append(tagValues, fmt.Sprintf(`'%s'`, tv))
		}
		query += fmt.Sprintf(` (%s) TAGS (%s)`,
			strings.Join(tagKeys, ", "), strings.Join(tagValues, ", "))

		// COLUMNS
		fieldKeys := []string{"ts"}
		fieldValues := []string{fmt.Sprintf("%d", record.Time().UnixNano())}
		for fk, fv := range record.Fields() {
			fieldKeys = append(fieldKeys, fk)
			if _, ok := fv.(string); ok {
				fieldValues = append(fieldValues, fmt.Sprintf(`'%s'`, fv))
			} else {
				fieldValues = append(fieldValues, fmt.Sprintf(`%v`, fv))
			}
		}
		query += fmt.Sprintf(` (%s) VALUES (%s)`,
			strings.Join(fieldKeys, ", "), strings.Join(fieldValues, ", "))
	}
	query += `;`

	stmt, err := ds.c.Prepare(query)
	if err != nil {
		return err
	}
	res, err := stmt.Exec()
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	ds.lg.Debugf("TDengine: write %d rows", affected)
	return nil
}

func (ds *TDengineDataStore) Query(ctx context.Context, req *query.Request) (*query.Response, error) {
	return nil, fmt.Errorf("implement me")
}

func (ds *TDengineDataStore) Export(ctx context.Context, req *query.Request, writer io.Writer, format string) error {
	return fmt.Errorf("implement me")
}

func typeOf(fieldType string) string {
	switch fieldType {
	case models.PropertyValueTypeBool:
		return "bool"
	case models.PropertyValueTypeInt, models.PropertyValueTypeUint:
		return "int"
	case models.PropertyValueTypeFloat:
		return "float"
	case models.PropertyValueTypeString:
		// FIXME support string more than 256 runes
		return fmt.Sprintf("NCHAR(%d)", 256)
	}

	return fieldType
}
