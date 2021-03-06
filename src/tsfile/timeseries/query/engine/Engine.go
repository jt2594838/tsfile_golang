package engine

import (
	"tsfile/file/metadata"
	"tsfile/timeseries/read"
	"tsfile/timeseries/query"
	"tsfile/timeseries/query/dataset"
	"log"
	"fmt"
	"tsfile/common/constant"
	"tsfile/timeseries/read/reader"
	impl2 "tsfile/timeseries/query/dataset/impl"
	"strings"
	"tsfile/file/header"
	"tsfile/timeseries/read/reader/impl/seek"
	"tsfile/timeseries/read/reader/impl/basic"
)

type Engine struct {
	reader *read.TsFileSequenceReader
	fileMeta *metadata.FileMetaData
}

func (e *Engine) Open(reader *read.TsFileSequenceReader) {
	e.reader = reader
	e.fileMeta = reader.ReadFileMetadata()
}

func (e *Engine) Close() {
	e.reader.Close()
	e.reader = nil
	e.fileMeta = nil
}

func (e *Engine) Query(exp *query.QueryExpression) dataset.IQueryDataSet{
	dataSet := e.decideQuerySet(exp)
	return dataSet
}

func (e *Engine) decideQuerySet(exp *query.QueryExpression) dataset.IQueryDataSet {
	if len(exp.ConditionPaths()) == 0 {
		exp.SetConditionPaths(exp.SelectPaths())
	}
	selectReaderMap := e.constructSeekableReaderMap(exp)
	conditionReaderMap := e.consturctReaderMapFromPaths(exp.ConditionPaths())
	return impl2.NewTimestampQueryDataSet(exp.SelectPaths(), exp.ConditionPaths(), selectReaderMap, conditionReaderMap, exp.Filter())
}

func (e *Engine) consturctReaderMapFromPaths(paths []string) map[string]reader.TimeValuePairReader{
	readerMap := make(map[string]reader.TimeValuePairReader)
	for _, path := range paths{
		readerMap[path] = e.constructReader(path)
	}
	return readerMap
}

func (e *Engine) constructReaderMap(exp *query.QueryExpression) map[string]reader.TimeValuePairReader{
	readerMap := make(map[string]reader.TimeValuePairReader)
	for _, path := range exp.SelectPaths(){
		readerMap[path] = e.constructReader(path)
	}
	for _, path := range exp.ConditionPaths(){
		if _, ok := readerMap[path]; !ok {
			readerMap[path] = e.constructReader(path)
		}
	}
	return readerMap
}

func (e *Engine) constructSeekableReaderMap(exp *query.QueryExpression) map[string]reader.ISeekableTimeValuePairReader{
	readerMap := make(map[string]reader.ISeekableTimeValuePairReader)
	for _, path := range exp.SelectPaths(){
		readerMap[path] = e.constructSeekableReader(path)
	}
	for _, path := range exp.ConditionPaths(){
		if _, ok := readerMap[path]; !ok {
			readerMap[path] = e.constructSeekableReader(path)
		}
	}
	return readerMap
}

func (e *Engine) constructReader(path string) reader.TimeValuePairReader {
	dataType, encoding, offsets, sizes,  _ := e.getPageInfo(path, false)
	return basic.NewSeriesReader(offsets, sizes, e.reader, dataType, encoding)
}

func (e *Engine) constructSeekableReader(path string) reader.ISeekableTimeValuePairReader {
	dataType, encoding, offsets, sizes,  headers := e.getPageInfo(path, true)
	return seek.NewSeekableSeriesReader(offsets, sizes, e.reader, headers, dataType, encoding)
}

func (e *Engine) getPageInfo(path string, needHeader bool) (dataType constant.TSDataType, encoding constant.TSEncoding,
											offsets []int64, sizes []int, pageHeaders []*header.PageHeader){
	pathSplits := strings.Split(path, constant.PATH_SEPARATOR)
	pathLevelLen := len(pathSplits)
	if pathLevelLen < 2 {
		log.Println(fmt.Println("Invalid path : %s", path))
		return 0, 0, nil, nil, nil
	}
	deviceId := strings.Join(pathSplits[0:pathLevelLen-1], constant.PATH_SEPARATOR)
	sensorId := pathSplits[pathLevelLen-1]

	dataType = e.getDataType(sensorId)
	if dataType == constant.INVALID {
		log.Println(fmt.Sprintf("No such timeseries in this file : %s", path))
		return 0, 0, nil, nil, nil
	}

	deviceMeta, ok := e.fileMeta.DeviceMap()[deviceId]
	if !ok {
		log.Println(fmt.Sprintf("No such timeseries in this file : %s", path))
		return 0, 0, nil, nil, nil
	}

	var headers []*header.PageHeader
	// find the offsets, sizes and headers(optional) of all pages of this path
	for ele := deviceMeta.RowGroupMetadataList().Front(); ele != nil; ele = ele.Next() {
		if rowGroupMeta, ok := ele.Value.(*metadata.RowGroupMetaData); ok {
			for c := rowGroupMeta.TimeSeriesChunkMetaDataList().Front(); c != nil; c = c.Next() {
				if chunkMeta, ok := c.Value.(*metadata.ChunkMetaData); ok {
					if chunkMeta.Sensor() != sensorId {
						continue
					}
					chunkHeader := e.reader.ReadChunkHeaderAt(chunkMeta.FileOffsetOfCorrespondingData())
					encoding = chunkHeader.GetEncodingType()
					pos := e.reader.Pos()
					for i := 0; i < chunkHeader.GetNumberOfPages(); i ++ {
						pageHeader := e.reader.ReadPageHeaderAt(dataType, pos)
						offsets = append(offsets, e.reader.Pos())
						sizes = append(sizes, pageHeader.GetCompressedSize())
						pos = e.reader.Pos() + int64(pageHeader.GetCompressedSize())
						if needHeader {
							headers = append(headers, pageHeader)
						}
					}
				}
			}
		} else {
			log.Println("Invalid RowgroupMetadata")
		}
	}
	return dataType, encoding, offsets, sizes, headers
}


func (e* Engine) getDataType(path string) constant.TSDataType {
	if tsMeta, ok := e.fileMeta.TimeSeriesMetadataMap()[path]; ok {
		return tsMeta.DataType()
	}
	return constant.INVALID
}