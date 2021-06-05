package osm

import (
	"github.com/flywave/go-pbf"
)

const (
	BLOB_RAW       pbf.TagType = 1
	BLOB_RAW_SIZE  pbf.TagType = 2
	BLOB_ZLIB_DATA pbf.TagType = 3
	BLOB_LZMA_DATA pbf.TagType = 4
)

const (
	BLOB_HEADER_TYPE       pbf.TagType = 1
	BLOB_HEADER_INDEX_DATA pbf.TagType = 2
	BLOB_HEADER_DATA_SIZE  pbf.TagType = 3
)

const (
	HEADER_BLOCK_BBOX                        pbf.TagType = 1
	HEADER_BLOCK_REQUIRED_FEATURES           pbf.TagType = 4
	HEADER_BLOCK_OPTIONAL_FEATURES           pbf.TagType = 5
	HEADER_BLOCK_WRITING_PROGRAM             pbf.TagType = 16
	HEADER_BLOCK_SOURCE                      pbf.TagType = 17
	HEADER_BLOCK_REPLICATION_TIMESTAMP       pbf.TagType = 32
	HEADER_BLOCK_REPLICATION_SEQUENCE_NUMBER pbf.TagType = 33
	HEADER_BLOCK_REPLICATION_BASE_URL        pbf.TagType = 34
)

const (
	HEADER_BBOX_LEFT   pbf.TagType = 1
	HEADER_BBOX_RIGHT  pbf.TagType = 2
	HEADER_BBOX_TOP    pbf.TagType = 3
	HEADER_BBOX_BOTTOM pbf.TagType = 4
)

const (
	PRIMITIVE_BLOCK_STRINGTABLE      pbf.TagType = 1
	PRIMITIVE_BLOCK_PRIMITIVEGROUP   pbf.TagType = 2
	PRIMITIVE_BLOCK_GRANULARITY      pbf.TagType = 17
	PRIMITIVE_BLOCK_LAT_OFFSET       pbf.TagType = 19
	PRIMITIVE_BLOCK_LON_OFFSET       pbf.TagType = 20
	PRIMITIVE_BLOCK_DATA_GRANULARITY pbf.TagType = 18
)

const (
	PRIMITIVE_GROUP_NODES     pbf.TagType = 1
	PRIMITIVE_GROUP_DENSE     pbf.TagType = 2
	PRIMITIVE_GROUP_WAYS      pbf.TagType = 3
	PRIMITIVE_GROUP_RELATION  pbf.TagType = 4
	PRIMITIVE_GROUP_CHANGESET pbf.TagType = 5
)

const (
	STRINGTABLE_S pbf.TagType = 1
)

const (
	INFO_VERSION   pbf.TagType = 1
	INFO_TIMESTAMP pbf.TagType = 2
	INFO_CHANGESET pbf.TagType = 3
	INFO_UID       pbf.TagType = 4
	INFO_USER_SID  pbf.TagType = 5
	INFO_VISIBLE   pbf.TagType = 6
)

const (
	DENSE_INFO_VERSION   pbf.TagType = 1
	DENSE_INFO_TIMESTAMP pbf.TagType = 2
	DENSE_INFO_CHANGESET pbf.TagType = 3
	DENSE_INFO_UID       pbf.TagType = 4
	DENSE_INFO_USER_SID  pbf.TagType = 5
	DENSE_INFO_VISIBLE   pbf.TagType = 6
)

const (
	CHANGESET_ID pbf.TagType = 1
)

const (
	NODE_ID   pbf.TagType = 1
	NODE_KEYS pbf.TagType = 2
	NODE_VALS pbf.TagType = 3
	NODE_INFO pbf.TagType = 4
	NODE_LAT  pbf.TagType = 8
	NODE_LON  pbf.TagType = 9
)

const (
	DENSE_NODE_ID        pbf.TagType = 1
	DENSE_NODE_INFO      pbf.TagType = 5
	DENSE_NODE_LAT       pbf.TagType = 8
	DENSE_NODE_LON       pbf.TagType = 9
	DENSE_NODE_KEYS_VALS pbf.TagType = 10
)

const (
	WAY_ID   pbf.TagType = 1
	WAY_KEYS pbf.TagType = 2
	WAY_VALS pbf.TagType = 3
	WAY_INFO pbf.TagType = 4
	WAY_REFS pbf.TagType = 8
)

const (
	RELATION_ID        pbf.TagType = 1
	RELATION_KEYS      pbf.TagType = 2
	RELATION_VALS      pbf.TagType = 3
	RELATION_INFO      pbf.TagType = 4
	RELATION_ROLES_SID pbf.TagType = 8
	RELATION_MEMIDS    pbf.TagType = 9
	RELATION_TYPES     pbf.TagType = 10
)
