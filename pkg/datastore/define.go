package datastore

import (
	"context"
	"fmt"
	"github.com/apache/arrow/go/v8/arrow"
	"github.com/apache/arrow/go/v8/arrow/array"
	"github.com/apache/arrow/go/v8/arrow/csv"
	"github.com/apache/arrow/go/v8/parquet"
	"github.com/apache/arrow/go/v8/parquet/pqarrow"
	"github.com/pkg/errors"
	"github.com/thingio/edge-device-manager/pkg/datastore/query"
	"github.com/thingio/edge-device-std/config"
	"github.com/thingio/edge-device-std/logger"
	"github.com/thingio/edge-device-std/models"
	"io"
	"time"
)

type (
	ExportFormat = string
)

const (
	ExportFormatArrowParquet ExportFormat = "parquet"
	ExportFormatCSV          ExportFormat = "csv"
)

func NewDataStore(ctx context.Context, lg *logger.Logger, opts *config.DataStoreOptions) (DataStore, error) {
	var ds DataStore

	ctx, cancel := context.WithCancel(ctx)
	super := dataStore{
		lg:     lg,
		ctx:    ctx,
		cancel: cancel,

		database:  opts.Database,
		batchSize: opts.BatchSize,
		ch:        make(chan *DeviceDataRecord, opts.BatchSize),
	}

	switch opts.Type {
	case config.DataStoreTypeInfluxDB:
		ids, err := NewInfluxDBDataStore(opts.InfluxDB)
		if err != nil {
			return nil, err
		}
		super.executor = ids
		ids.dataStore = super

		ds = ids
	case config.DataStoreTypeTDengine:
		tds, err := NewTDengineDataStore(opts.TDengine)
		if err != nil {
			return nil, err
		}
		super.executor = tds
		tds.dataStore = super

		ds = tds
	default:
		return nil, fmt.Errorf("unsupported datastore type: %s", opts.Type)
	}

	if err := ds.Connect(); err != nil {
		return nil, errors.Wrapf(err, "failed to connect to datastore: %+v", opts)
	}
	if err := ds.CreateDB(); err != nil {
		return nil, errors.Wrapf(err, "failed to create database: %s", opts.Database)
	}
	if err := ds.UseDB(); err != nil {
		return nil, errors.Wrapf(err, "failed to use database: %s", opts.Database)
	}
	return ds, nil
}

type Connector interface {
	Connect() error
	Close()
}
type Writer interface {
	CreateDB() error
	UseDB() error
	CreateTable(schema *DeviceDataSchema) error
	Write(records ...*DeviceDataRecord) error
}
type Reader interface {
	Query(ctx context.Context, query *query.Request) (*query.Response, error)

	Export(ctx context.Context, query *query.Request, writer io.Writer, format string) error
}

type DataStore interface {
	Connector
	Writer
	Reader
}

type dataStore struct {
	lg     *logger.Logger
	ctx    context.Context
	cancel context.CancelFunc

	database  string
	batchSize int
	ch        chan *DeviceDataRecord

	executor DataStore
}

func (ds *dataStore) Connect() error {
	if err := ds.executor.Connect(); err != nil {
		return err
	}

	ds.listen(ds.ch)
	return nil
}
func (ds *dataStore) Close() {
	ds.cancel()
	ds.executor.Close()
}

func (ds *dataStore) CreateDB() error {
	return ds.executor.CreateDB()
}
func (ds *dataStore) UseDB() error {
	return ds.executor.UseDB()
}
func (ds *dataStore) CreateTable(schema *DeviceDataSchema) error {
	return ds.executor.CreateTable(schema)
}
func (ds *dataStore) Write(records ...*DeviceDataRecord) error {
	for _, record := range records {
		ds.ch <- record
	}
	return nil
}
func (ds *dataStore) listen(ch <-chan *DeviceDataRecord) {
	for {
		select {
		case record, ok := <-ch:
			if !ok {
				ds.lg.Warnf("the channel receiving points has been closed")
				return
			}

			size := len(ch) + 1
			records := make([]*DeviceDataRecord, size)
			records[0] = record
			for i := 1; i < size; i++ {
				records[i] = <-ch
			}
			if err := ds.executor.Write(records...); err != nil {
				ds.lg.WithError(err).Errorf("fail to write batchly")
			} else {
				ds.lg.Debugf("success to write batchly %d records", size)
			}
		case <-ds.ctx.Done():
			size := len(ch)
			records := make([]*DeviceDataRecord, size)
			for i := 0; i < size; i++ {
				records[i] = <-ch
			}
			if err := ds.executor.Write(records...); err != nil {
				ds.lg.WithError(err).Errorf("fail to write batchly")
			} else {
				ds.lg.Debugf("success to write batchly %d records", size)
			}
			return
		}
	}
}

func (ds *dataStore) Query(ctx context.Context, req *query.Request) (*query.Response, error) {
	return ds.executor.Query(ctx, req)
}

func (ds *dataStore) Export(ctx context.Context, query *query.Request, writer io.Writer, format string) error {
	return ds.executor.Export(ctx, query, writer, format)
}
func (ds *dataStore) export(record arrow.Record, writer io.Writer, format string) error {
	switch format {
	case ExportFormatArrowParquet:
		return ds.exportAsArrowParquet(record, writer)
	case ExportFormatCSV:
		return ds.exportAsCSV(record, writer)
	default:
		return fmt.Errorf("unsupport export format: %s", format)
	}
}

func (ds *dataStore) exportAsArrowParquet(record arrow.Record, writer io.Writer) error {
	tbl := array.NewTableFromRecords(record.Schema(), []arrow.Record{record})
	defer tbl.Release()

	return pqarrow.WriteTable(tbl, writer, tbl.NumRows(),
		parquet.NewWriterProperties(
			parquet.WithDictionaryDefault(false),
		),
		pqarrow.DefaultWriterProps(),
	)
}
func (ds *dataStore) exportAsCSV(record arrow.Record, writer io.Writer) error {
	w := csv.NewWriter(writer, record.Schema(),
		csv.WithComma(';'),
		csv.WithCRLF(false),
		csv.WithHeader(true),
		csv.WithNullWriter("null"),
	)
	if err := w.Write(record); err != nil {
		return err
	}
	return w.Flush()
}

type (
	DeviceDataSchemaColumnName = string
)

const (
	DeviceDataSchemaColumnNameProtocol DeviceDataSchemaColumnName = "protocol"
	DeviceDataSchemaColumnNameProduct  DeviceDataSchemaColumnName = "product"
	DeviceDataSchemaColumnNameDevice   DeviceDataSchemaColumnName = "device"
)

type DeviceDataSchemaColumn struct {
	Name string
	Type string
}

type DeviceDataSchema struct {
	ProtocolID string
	ProductID  string
	DeviceID   string
	Properties []*models.ProductProperty
}

// AggregatedTableName is used to get the name of table for aggregating logically data.
// For InfluxDB, it is the measurement.
// For TDengine, it is the STable.
func (w DeviceDataSchema) AggregatedTableName() string {
	return fmt.Sprintf("%s_%s", w.ProtocolID, w.ProductID)
}

// TableName is used to get the name of table for storing physically data.
// For InfluxDB, there is no term corresponding to it.
// For TDengine, it is the Table.
func (w DeviceDataSchema) TableName() string {
	return fmt.Sprintf("%s_%s_%s", w.ProtocolID, w.ProductID, w.DeviceID)
}
func (w DeviceDataSchema) Tags() []*DeviceDataSchemaColumn {
	return []*DeviceDataSchemaColumn{
		{
			Name: DeviceDataSchemaColumnNameProtocol,
			Type: models.PropertyValueTypeString,
		},
		{
			Name: DeviceDataSchemaColumnNameProduct,
			Type: models.PropertyValueTypeString,
		},
		{
			Name: DeviceDataSchemaColumnNameDevice,
			Type: models.PropertyValueTypeString,
		},
	}
}
func (w DeviceDataSchema) Fields() []*DeviceDataSchemaColumn {
	fields := make([]*DeviceDataSchemaColumn, len(w.Properties))
	for i, property := range w.Properties {
		fields[i] = &DeviceDataSchemaColumn{
			property.Name,
			property.FieldType,
		}
	}
	return fields
}

type DeviceDataRecord struct {
	ProtocolID string
	ProductID  string
	DeviceID   string
	Properties map[models.ProductPropertyID]*models.DeviceData
}

// AggregatedTableName is used to get the name of table for aggregating logically data.
// For InfluxDB, it is the measurement.
// For TDengine, it is the STable.
func (w DeviceDataRecord) AggregatedTableName() string {
	return fmt.Sprintf("%s_%s", w.ProtocolID, w.ProductID)
}

// TableName is used to get the name of table for storing physically data.
// For InfluxDB, there is no term corresponding to it.
// For TDengine, it is the Table.
func (w DeviceDataRecord) TableName() string {
	return fmt.Sprintf("%s_%s_%s", w.ProtocolID, w.ProductID, w.DeviceID)
}
func (w DeviceDataRecord) Tags() map[string]string {
	return map[string]string{
		DeviceDataSchemaColumnNameProtocol: w.ProtocolID,
		DeviceDataSchemaColumnNameProduct:  w.ProductID,
		DeviceDataSchemaColumnNameDevice:   w.DeviceID,
	}
}
func (w DeviceDataRecord) Fields() map[string]interface{} {
	fields := map[string]interface{}{}
	for _, property := range w.Properties {
		fields[property.Name] = property.Value
	}
	return fields
}

func (w DeviceDataRecord) Time() time.Time {
	for _, property := range w.Properties {
		return property.Ts
	}
	return time.Now()
}
