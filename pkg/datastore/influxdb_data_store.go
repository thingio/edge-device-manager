package datastore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/arrow/go/v8/arrow"
	"github.com/apache/arrow/go/v8/arrow/array"
	"github.com/apache/arrow/go/v8/arrow/memory"
	"github.com/influxdata/influxdb/client"
	"github.com/influxdata/influxdb/models"
	"github.com/thingio/edge-device-manager/pkg/datastore/query"
	"github.com/thingio/edge-device-std/config"
	"io"
	"net/url"
	"reflect"
)

type (
	Point       = client.Point
	BatchPoints = client.BatchPoints
)

func NewInfluxDBDataStore(cfg *config.InfluxDBOptions) (*InfluxDBDataStore, error) {
	addr, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}
	c, err := client.NewClient(client.Config{
		URL:              *addr,
		Username:         cfg.Username,
		Password:         cfg.Password,
		UserAgent:        cfg.UserAgent,
		Timeout:          cfg.Timeout,
		Precision:        cfg.Precision,
		WriteConsistency: cfg.WriteConsistency,
	})
	if err != nil {
		return nil, err
	}
	return &InfluxDBDataStore{
		cfg: cfg,
		c:   c,
	}, nil
}

type InfluxDBDataStore struct {
	dataStore
	cfg *config.InfluxDBOptions
	c   *client.Client
}

func (ds *InfluxDBDataStore) Connect() error {
	if _, _, err := ds.c.Ping(); err != nil {
		return err
	}

	return nil
}
func (ds *InfluxDBDataStore) Close() {
	// nothing to do
}

func (ds *InfluxDBDataStore) CreateDB() error {
	rsp, err := ds.c.Query(client.Query{
		Command: fmt.Sprintf(`CREATE DATABASE "%s"`, ds.database),
	})
	if err != nil {
		return err
	}
	if rsp != nil && rsp.Err != nil {
		return rsp.Err
	}
	return nil
}
func (ds *InfluxDBDataStore) UseDB() error {
	return nil
}
func (ds *InfluxDBDataStore) CreateTable(schema *DeviceDataSchema) error {
	return nil
}
func (ds *InfluxDBDataStore) Write(records ...*DeviceDataRecord) error {
	points := make([]Point, len(records))
	for i, record := range records {
		points[i] = Point{
			Measurement: record.AggregatedTableName(),
			Tags:        record.Tags(),
			Time:        record.Time(),
			Fields:      record.Fields(),
			Precision:   ds.cfg.Precision,
		}
	}

	batchPoints := BatchPoints{
		Points:   points,
		Database: ds.database,
	}
	rsp, err := ds.c.Write(batchPoints)
	if err != nil {
		return err
	}
	if rsp != nil && rsp.Err != nil {
		return rsp.Err
	}
	return nil
}

func (ds *InfluxDBDataStore) Query(ctx context.Context, req *query.Request) (*query.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("query cannot be empty")
	}

	command := req.BuildStatement(config.DataStoreTypeInfluxDB)
	rsp, err := ds.c.QueryContext(ctx, client.Query{
		Command:  command,
		Database: ds.database,
	})
	if err != nil {
		return nil, err
	}

	qr := &query.Response{}
	if len(rsp.Results) == 0 || len(rsp.Results[0].Series) == 0 {
		qr.Rows = make([]map[string]interface{}, 0)
	} else {
		series := rsp.Results[0].Series[0]
		rows := make([]map[string]interface{}, len(series.Values))
		for rIdx, values := range series.Values {
			rows[rIdx] = make(map[string]interface{})
			for cIdx, column := range series.Columns {
				rows[rIdx][column] = values[cIdx]
			}
		}
		qr.Rows = rows
	}
	return qr, nil
}

func (ds *InfluxDBDataStore) Export(ctx context.Context, req *query.Request, writer io.Writer, format string) error {
	if req == nil {
		return fmt.Errorf("query cannot be empty")
	}

	// TODO large data optimization
	command := req.BuildStatement(config.DataStoreTypeInfluxDB)
	rsp, err := ds.c.QueryContext(ctx, client.Query{
		Command:  command,
		Database: ds.database,
	})
	if err != nil {
		return err
	}

	if len(rsp.Results) == 0 || len(rsp.Results[0].Series) == 0 {
		return nil
	}
	record, err := ds.exportASArrowRecord(rsp.Results[0].Series[0])
	if err != nil {
		return err
	}
	defer record.Release()

	return ds.dataStore.export(record, writer, format)
}
func (ds *InfluxDBDataStore) exportASArrowRecord(series models.Row) (arrow.Record, error) {
	fields := make([]arrow.Field, len(series.Columns))
	for cIdx, column := range series.Columns {
		var fieldType arrow.DataType
		for _, values := range series.Values {
			value := values[cIdx]
			if value == nil {
				continue
			}
			switch reflect.TypeOf(value).Kind() {
			case reflect.String:
				if _, ok := value.(json.Number); ok {
					fieldType = arrow.PrimitiveTypes.Float64
				} else {
					fieldType = arrow.BinaryTypes.String
				}
			case reflect.Bool:
				fieldType = arrow.FixedWidthTypes.Boolean
			case reflect.Int32:
				fieldType = arrow.PrimitiveTypes.Int32
			case reflect.Int, reflect.Int64:
				fieldType = arrow.PrimitiveTypes.Int64
			case reflect.Float32:
				fieldType = arrow.PrimitiveTypes.Float32
			case reflect.Float64:
				fieldType = arrow.PrimitiveTypes.Float64
			default:
				return nil, fmt.Errorf("unsupported data type: %s", reflect.TypeOf(series.Values[cIdx][0]).Name())
			}
			break
		}
		fields[cIdx] = arrow.Field{Name: column, Type: fieldType, Nullable: true}
	}
	schema := arrow.NewSchema(fields, nil)

	rb := array.NewRecordBuilder(memory.NewGoAllocator(), schema)
	defer rb.Release()
	for cIdx, field := range fields {
		valid := make([]bool, len(series.Values))
		switch field.Type {
		case arrow.BinaryTypes.String:
			values := make([]string, len(series.Values))
			for rIdx, row := range series.Values {
				if row[cIdx] == nil {
					valid[rIdx] = false
				} else {
					valid[rIdx] = true
					values[rIdx] = row[cIdx].(string)
				}
			}
			rb.Field(cIdx).(*array.StringBuilder).AppendValues(values, valid)
		case arrow.FixedWidthTypes.Boolean:
			values := make([]bool, len(series.Values))
			for rIdx, row := range series.Values {
				if row[cIdx] == nil {
					valid[rIdx] = false
				} else {
					valid[rIdx] = true
					values[rIdx] = row[cIdx].(bool)
				}
			}
			rb.Field(cIdx).(*array.BooleanBuilder).AppendValues(values, valid)
		case arrow.PrimitiveTypes.Int32:
			values := make([]int32, len(series.Values))
			for rIdx, row := range series.Values {
				if row[cIdx] == nil {
					valid[rIdx] = false
				} else {
					valid[rIdx] = true
					values[rIdx] = row[cIdx].(int32)
				}
			}
			rb.Field(cIdx).(*array.Int32Builder).AppendValues(values, valid)
		case arrow.PrimitiveTypes.Int64:
			values := make([]int64, len(series.Values))
			for rIdx, row := range series.Values {
				if row[cIdx] == nil {
					valid[rIdx] = false
				} else {
					valid[rIdx] = true
					values[rIdx] = row[cIdx].(int64)
				}
			}
			rb.Field(cIdx).(*array.Int64Builder).AppendValues(values, valid)
		case arrow.PrimitiveTypes.Float32:
			values := make([]float32, len(series.Values))
			for rIdx, row := range series.Values {
				if row[cIdx] == nil {
					valid[rIdx] = false
				} else {
					valid[rIdx] = true
					values[rIdx] = row[cIdx].(float32)
				}
			}
			rb.Field(cIdx).(*array.Float32Builder).AppendValues(values, valid)
		case arrow.PrimitiveTypes.Float64:
			values := make([]float64, len(series.Values))
			for rIdx, row := range series.Values {
				if row[cIdx] == nil {
					valid[rIdx] = false
				} else {
					valid[rIdx] = true
					if v, ok := row[cIdx].(json.Number); ok {
						fv, _ := v.Float64()
						values[rIdx] = fv
					} else {
						values[rIdx] = row[cIdx].(float64)
					}
				}
			}
			rb.Field(cIdx).(*array.Float64Builder).AppendValues(values, valid)
		}
	}
	return rb.NewRecord(), nil
}
