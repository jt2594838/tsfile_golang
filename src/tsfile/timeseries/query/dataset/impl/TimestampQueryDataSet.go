package impl

import (
	"tsfile/timeseries/read/datatype"
	"tsfile/timeseries/query/timegen"
	"tsfile/timeseries/read/reader/impl/seek"
	"tsfile/timeseries/read/reader"
	"tsfile/timeseries/filter"
	"tsfile/timeseries/query/timegen/impl"
	"tsfile/common/constant"
	"tsfile/timeseries/read/reader/impl/basic"
)

type TimestampQueryDataSet struct {
	tGen timegen.ITimestampGenerator
	rGen *basic.FilteredRowReader
	r reader.ISeekableRowReader

	currTime int64
	current  *datatype.RowRecord
}

func NewTimestampQueryDataSet(selectPaths []string, conditionPaths []string,
	selectReaderMap map[string]reader.ISeekableTimeValuePairReader, conditionReaderMap map[string]reader.TimeValuePairReader, filter filter.Filter) *TimestampQueryDataSet {
	tGen := impl.NewRowRecordTimestampGenerator(conditionPaths, conditionReaderMap, filter)
	rGen := basic.NewFilteredRowReader(conditionPaths, conditionReaderMap, filter)
	r := seek.NewSeekableRowReader(selectPaths, selectReaderMap)
	return &TimestampQueryDataSet{tGen:tGen, rGen:rGen, r:r, currTime:constant.INVALID_TIMESTAMP}
}

func (set *TimestampQueryDataSet) fetch() {
	//if set.tGen.HasNext() {
	//	currTime := set.tGen.Next()
	//	if set.r.Seek(currTime) {
	//		set.current = set.r.Current()
	//	}
	//}
	if set.rGen.HasNext() {
		currTime := set.rGen.Next().Timestamp()
		if set.r.Seek(currTime) {
			set.current = set.r.Current()
		}
	}
}

func (set *TimestampQueryDataSet) HasNext() bool {
	if set.current != nil {
		return true
	}
	set.fetch()
	return set.current != nil
}

func (set *TimestampQueryDataSet) Next() *datatype.RowRecord {
	ret := set.current
	set.current = nil
	set.fetch()
	return ret
}

func (set *TimestampQueryDataSet) Close() {
	set.r.Close()
}

