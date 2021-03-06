package header

import (
	_ "bufio"
	_ "log"
	_ "os"
	"tsfile/common/constant"
	"tsfile/common/utils"
	"tsfile/file/metadata/statistics"
)

type PageHeader struct {
	uncompressedSize int
	compressedSize   int
	numberOfValues   int
	max_timestamp    int64
	min_timestamp    int64
	statistics       statistics.Statistics
	serializedSize   int
}

func (p *PageHeader) Min_timestamp() int64 {
	return p.min_timestamp
}

func (p *PageHeader) Max_timestamp() int64 {
	return p.max_timestamp
}

func (h *PageHeader) Deserialize(reader *utils.FileReader, dataType constant.TSDataType) {
	h.uncompressedSize = int(reader.ReadInt())
	h.compressedSize = int(reader.ReadInt())
	h.numberOfValues = int(reader.ReadInt())
	h.max_timestamp = reader.ReadLong()
	h.min_timestamp = reader.ReadLong()
	h.statistics = statistics.Deserialize(reader, dataType)

	h.serializedSize = 3*constant.INT_LEN + 2*constant.LONG_LEN + h.statistics.GetSerializedSize()
}

func (h *PageHeader) GetUncompressedSize() int {
	return h.uncompressedSize
}

func (h *PageHeader) GetCompressedSize() int {
	return h.compressedSize
}

func (h *PageHeader) GetNumberOfValues() int {
	return h.numberOfValues
}

func (h *PageHeader) GetSerializedSize() int {
	return h.serializedSize
}
