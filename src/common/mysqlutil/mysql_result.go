package mysqlutil

import (
	"strconv"

	"golang.org/x/exp/slices"
)

// 栏位值
type FieldValue string

func (fv *FieldValue) String() string {
	return string(*fv)
}

func (fv *FieldValue) Int() int {
	return int(fv.Int64())
}

func (fv *FieldValue) Uint() uint {
	return uint(fv.Uint64())
}

func (fv *FieldValue) Int32() int32 {
	return int32(fv.Int64())
}

func (fv *FieldValue) Uint32() uint32 {
	return uint32(fv.Uint64())
}

func (fv *FieldValue) Int64() int64 {
	i, _ := strconv.ParseInt(string(*fv), 10, 64)
	return i
}

func (fv *FieldValue) Uint64() uint64 {
	i, _ := strconv.ParseUint(string(*fv), 10, 64)
	return i
}

// 单行数据
type ResultRow struct {
	columns     []string //pointer equals to query result's columns pointer
	fieldValues []FieldValue
}

func (row *ResultRow) Columns() []string {
	return row.columns
}

func (row *ResultRow) Field(column string) *FieldValue {
	i := slices.Index(row.columns, column)
	if i < 0 {
		return nil
	}
	return &row.fieldValues[i]
}

func (row *ResultRow) FieldString(column string) string {
	f := row.Field(column)
	if f == nil {
		return ""
	}
	return f.String()
}

func (row *ResultRow) FieldInt(column string) int {
	f := row.Field(column)
	if f == nil {
		return 0
	}
	return f.Int()
}

func (row *ResultRow) FieldUint(column string) uint {
	f := row.Field(column)
	if f == nil {
		return 0
	}
	return f.Uint()
}

func (row *ResultRow) FieldInt32(column string) int32 {
	f := row.Field(column)
	if f == nil {
		return 0
	}
	return f.Int32()
}

func (row *ResultRow) FieldUint32(column string) uint32 {
	f := row.Field(column)
	if f == nil {
		return 0
	}
	return f.Uint32()
}

func (row *ResultRow) FieldInt64(column string) int64 {
	f := row.Field(column)
	if f == nil {
		return 0
	}
	return f.Int64()
}

func (row *ResultRow) FieldUint64(column string) uint64 {
	f := row.Field(column)
	if f == nil {
		return 0
	}
	return f.Uint64()
}

func (row *ResultRow) FieldByIndex(index int) *FieldValue {
	if index >= len(row.fieldValues) {
		return nil
	}
	return &row.fieldValues[index]
}

func (row *ResultRow) FieldByIndexString(index int) string {
	f := row.FieldByIndex(index)
	if f == nil {
		return ""
	}
	return f.String()
}

func (row *ResultRow) FieldByIndexInt(index int) int {
	f := row.FieldByIndex(index)
	if f == nil {
		return 0
	}
	return f.Int()
}

func (row *ResultRow) FieldByIndexUint(index int) uint {
	f := row.FieldByIndex(index)
	if f == nil {
		return 0
	}
	return f.Uint()
}

func (row *ResultRow) FieldByIndexInt32(index int) int32 {
	f := row.FieldByIndex(index)
	if f == nil {
		return 0
	}
	return f.Int32()
}

func (row *ResultRow) FieldByIndexUint32(index int) uint32 {
	f := row.FieldByIndex(index)
	if f == nil {
		return 0
	}
	return f.Uint32()
}

func (row *ResultRow) FieldByIndexInt64(index int) int64 {
	f := row.FieldByIndex(index)
	if f == nil {
		return 0
	}
	return f.Int64()
}

func (row *ResultRow) FieldByIndexUint64(index int) uint64 {
	f := row.FieldByIndex(index)
	if f == nil {
		return 0
	}
	return f.Uint64()
}

// 多行数据
type QueryResult struct {
	columns []string
	rows    []*ResultRow
}

func (result *QueryResult) RowCount() int {
	return len(result.rows)
}

func (result *QueryResult) Row(index int) *ResultRow {
	if index >= len(result.rows) {
		return nil
	}
	return result.rows[index]
}

func (result *QueryResult) Rows() []*ResultRow {
	return result.rows
}
