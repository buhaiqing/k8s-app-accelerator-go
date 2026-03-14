package config

// K8sResource K8s 资源配置（CPU/内存）
type K8sResource struct {
	LimitsCPU      string `yaml:"limits_cpu" json:"limits_cpu"`
	LimitsMemory   string `yaml:"limits_memory" json:"limits_memory"`
	RequestsCPU    string `yaml:"requests_cpu" json:"requests_cpu"`
	RequestsMemory string `yaml:"requests_memory" json:"requests_memory"`
}

// ResourceGroup 对应 resources.yaml
type ResourceGroup struct {
	// K8s 资源配置（按环境）
	Production K8sResource `yaml:"production" json:"production"`
	Default    K8sResource `yaml:"default" json:"default"`
	Int        K8sResource `yaml:"int" json:"int"`

	// 其他资源
	RDS           []RDSResource      `yaml:"rds" json:"rds"`
	PostgreSQL    []PGResource       `yaml:"pg" json:"pg"`
	Redis         []RedisResource    `yaml:"redis" json:"redis"`
	MongoDB       []MongoResource    `yaml:"mongo" json:"mongo"`
	Elasticsearch []ESResource       `yaml:"es" json:"es"`
	OSS           []OSSResource      `yaml:"oss" json:"oss"`
	MQ            []MQResource       `yaml:"mq" json:"mq"`
	Oracle        []OracleResource   `yaml:"oracle" json:"oracle"`
	RabbitMQ      []RabbitMQResource `yaml:"rabbitmq" json:"rabbitmq"`
	OTS           []OTSResource      `yaml:"ots" json:"ots"`
	DTFlow        []DTFlowResource   `yaml:"dtflow" json:"dtflow"`
}

// RDSResource RDS 资源
type RDSResource struct {
	Name               string `yaml:"name"`
	DatasourceURL      string `yaml:"datasource_url"`
	DatasourceDB       string `yaml:"datasource_db"`
	DatasourcePort     string `yaml:"datasource_port"`
	DatasourceUsername string `yaml:"datasource_username"`
	DatasourcePassword string `yaml:"datasource_password"`
}

// PGResource PostgreSQL 资源
type PGResource struct {
	Name                 string `yaml:"name"`
	DatasourcePgURL      string `yaml:"datasource_pg_url"`
	DatasourcePgPort     string `yaml:"datasource_pg_port"`
	DatasourcePgDB       string `yaml:"datasource_pg_db"`
	DatasourcePgUsername string `yaml:"datasource_pg_username"`
	DatasourcePgPassword string `yaml:"datasource_pg_password"`
}

// RedisResource Redis 资源
type RedisResource struct {
	Name          string `yaml:"name"`
	RedisIP       string `yaml:"redisIp"`
	RedisPort     string `yaml:"redisPort"`
	RedisDb       string `yaml:"redisDb"`
	RedisPassword string `yaml:"redisPassword"`
}

// MongoResource MongoDB 资源
type MongoResource struct {
	Name            string `yaml:"name"`
	MongoDB         string `yaml:"mongo_db"`
	MongoDBURL      string `yaml:"mongodb_url"`
	MongoDBUsername string `yaml:"mongodb_username"`
	MongoDBPassword string `yaml:"mongodb_password"`
}

// ESResource Elasticsearch 资源
type ESResource struct {
	Name          string `yaml:"name"`
	ESInstanceID  string `yaml:"es_instanceId"`
	ESDomain      string `yaml:"es_domain"`
	ESRegionID    string `yaml:"es_regionId"`
	ESSecret      string `yaml:"es_secret"`
	ESAccessKeyID string `yaml:"es_accessKeyId"`
	ESURL         string `yaml:"es_url"`
	ESUsername    string `yaml:"es_username"`
	ESPassword    string `yaml:"es_password"`
}

// OSSResource OSS 资源
type OSSResource struct {
	Name                string `yaml:"name" json:"name"`
	OSSRoleArn          string `yaml:"oss_roleArn" json:"oss_roleArn"`
	OSSStsEndpoint      string `yaml:"oss_stsEndpoint" json:"oss_stsEndpoint"`
	BucketName          string `yaml:"bucketName" json:"bucketName"`
	OSSEndpoint         string `yaml:"ossEndpoint" json:"ossEndpoint"`
	OSSImageBaseURL     string `yaml:"ossimageBaseUrl" json:"ossimageBaseUrl"`
	OSSInternalEndpoint string `yaml:"ossinternalEndpoint" json:"ossinternalEndpoint"`
	OSSAccessKeyID      string `yaml:"oss_accessKeyId" json:"oss_accessKeyId"`
	OSSAccessKeySecret  string `yaml:"oss_accessKeySecret" json:"oss_accessKeySecret"`
	COSBucketName       string `yaml:"cos_bucketName" json:"cos_bucketName"`
	COSAppID            string `yaml:"cos_appId" json:"cos_appId"`
	COSRegion           string `yaml:"cos_region" json:"cos_region"`
	COSSecretID         string `yaml:"cos_secretId" json:"cos_secretId"`
	COSSecretKey        string `yaml:"cos_secretKey" json:"cos_secretKey"`
}

// MQResource MQ 资源
type MQResource struct {
	Name                string `yaml:"name"`
	CMQSecretID         string `yaml:"cmq_secretId"`
	CMQSecretKey        string `yaml:"cmq_secretKey"`
	CMQQueueEndpoint    string `yaml:"cmq_queueEndpoint"`
	CMQTopicEndpoint    string `yaml:"cmq_topicEndpoint"`
	MNSAccessKeyID      string `yaml:"mns_accessKeyId"`
	MNSAccessKeySecret  string `yaml:"mns_accessKeySecret"`
	MNSEndpoint         string `yaml:"mns_endpoint"`
	MNSStuffix          string `yaml:"mns_stuffix"`
	RocketMQAccessKey   string `yaml:"rocketmq_accesskey"`
	RocketMQSecretKey   string `yaml:"rocketmq_secretkey"`
	RocketMQInstanceID  string `yaml:"rocketmq_instanceId"`
	RocketMQRegionID    string `yaml:"rocketmq_regionId"`
	RocketMQNamesrvaddr string `yaml:"rocketmq_namesrvaddr"`
}

// OracleResource Oracle 资源
type OracleResource struct {
	Name       string `yaml:"name"`
	OracleURL  string `yaml:"oracle_url"`
	OraclePort string `yaml:"oracle_port"`
	OracleDB   string `yaml:"oracle_db"`
	OracleUser string `yaml:"oracle_user"`
	OraclePass string `yaml:"oracle_pass"`
}

// RabbitMQResource RabbitMQ 资源
type RabbitMQResource struct {
	Name           string `yaml:"name"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	Port           int    `yaml:"port"`
	Host           string `yaml:"host"`
	RetryDelayTime int    `yaml:"retry_delay_time"`
}

// OTSResource OTS 资源
type OTSResource struct {
	Name               string `yaml:"name"`
	OTSEndPoint        string `yaml:"ots_endPoint"`
	OTSAccessKeyID     string `yaml:"ots_accessKeyId"`
	OTSAccessKeySecret string `yaml:"ots_accessKeySecret"`
	OTSInstanceName    string `yaml:"ots_instanceName"`
}

// DTFlowResource DTFlow 资源
type DTFlowResource struct {
	Name               string `yaml:"name"`
	DFTConfigServerURL string `yaml:"dft_config_serverUrl"`
	DFTConfigTenant    string `yaml:"dft_config_tenant"`
	DFTConfigUsername  string `yaml:"dft_config_username"`
	DFTConfigPassword  string `yaml:"dft_config_password"`
}
