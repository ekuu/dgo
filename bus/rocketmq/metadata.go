package rocketmq

type Metadata map[string]string

const (
	MetadataRocketmqTag         = "rocketmq-tag"
	MetadataRocketmqKey         = "rocketmq-key"
	MetadataRocketmqShardingKey = "rocketmq-shardingkey"
)
