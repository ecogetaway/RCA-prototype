package model

import "strings"

type ApplicationType string

const (
	ApplicationTypeUnknown         ApplicationType = ""
	ApplicationTypePostgres        ApplicationType = "postgres"
	ApplicationTypePgbouncer       ApplicationType = "pgbouncer"
	ApplicationTypeMysql           ApplicationType = "mysql"
	ApplicationTypeCassandra       ApplicationType = "cassandra"
	ApplicationTypeElasticsearch   ApplicationType = "elasticsearch"
	ApplicationTypeOpensearch      ApplicationType = "opensearch"
	ApplicationTypeMemcached       ApplicationType = "memcached"
	ApplicationTypeRedis           ApplicationType = "redis"
	ApplicationTypeKeyDB           ApplicationType = "keydb"
	ApplicationTypeValkey          ApplicationType = "valkey"
	ApplicationTypeDragonfly       ApplicationType = "dragonfly"
	ApplicationTypeMongodb         ApplicationType = "mongodb"
	ApplicationTypeMongos          ApplicationType = "mongos"
	ApplicationTypeRabbitmq        ApplicationType = "rabbitmq"
	ApplicationTypeKafka           ApplicationType = "kafka"
	ApplicationTypeZookeeper       ApplicationType = "zookeeper"
	ApplicationTypeRDS             ApplicationType = "aws-rds"
	ApplicationTypeElastiCache     ApplicationType = "aws-elasticache"
	ApplicationTypeNats            ApplicationType = "nats"
	ApplicationTypeDotNet          ApplicationType = "dotnet"
	ApplicationTypeGolang          ApplicationType = "golang"
	ApplicationTypePHP             ApplicationType = "php"
	ApplicationTypeJava            ApplicationType = "java"
	ApplicationTypeNodeJS          ApplicationType = "nodejs"
	ApplicationTypePython          ApplicationType = "python"
	ApplicationTypeRuby            ApplicationType = "ruby"
	ApplicationTypeEnvoy           ApplicationType = "envoy"
	ApplicationTypePrometheus      ApplicationType = "prometheus"
	ApplicationTypeVictoriaMetrics ApplicationType = "victoria-metrics"
	ApplicationTypeVictoriaLogs    ApplicationType = "victoria-logs"
	ApplicationTypeClickHouse      ApplicationType = "clickhouse"
	ApplicationTypeFoundationDB    ApplicationType = "foundationdb"
	ApplicationTypeCorootCE        ApplicationType = "coroot-community-edition"
	ApplicationTypeCorootEE        ApplicationType = "coroot-enterprise-edition"
)

func (at ApplicationType) IsDatabase() bool {
	switch at {
	case ApplicationTypeCassandra, ApplicationTypeMemcached,
		ApplicationTypeZookeeper, ApplicationTypeElasticsearch, ApplicationTypeOpensearch, ApplicationTypePostgres,
		ApplicationTypeMysql, ApplicationTypeRedis, ApplicationTypeKeyDB, ApplicationTypeValkey, ApplicationTypeDragonfly,
		ApplicationTypeMongodb, ApplicationTypePrometheus, ApplicationTypeVictoriaMetrics, ApplicationTypeVictoriaLogs,
		ApplicationTypeClickHouse, ApplicationTypeFoundationDB:
		return true
	}
	return false
}

func (at ApplicationType) InstrumentationType() ApplicationType {
	switch at {
	case ApplicationTypeMongos:
		return ApplicationTypeMongodb
	case ApplicationTypeValkey, ApplicationTypeKeyDB, ApplicationTypeDragonfly:
		return ApplicationTypeRedis
	}
	return at
}

func (at ApplicationType) IsQueue() bool {
	switch at {
	case ApplicationTypeKafka, ApplicationTypeRabbitmq, ApplicationTypeNats:
		return true
	}
	return false
}

func (at ApplicationType) IsLanguage() bool {
	switch at {
	case ApplicationTypeGolang,
		ApplicationTypeDotNet,
		ApplicationTypePHP,
		ApplicationTypeJava,
		ApplicationTypeNodeJS,
		ApplicationTypePython,
		ApplicationTypeRuby:
		return true
	}
	return false
}

func (at ApplicationType) Weight() int {
	switch {
	case at.IsDatabase():
		return 1
	case at.IsQueue():
		return 2
	case at.IsLanguage():
		return 4
	case at == ApplicationTypeEnvoy: // when using service meshes, Envoy is deployed as a sidecar to each container
		return 5
	}
	return 3
}

func (at ApplicationType) AuditReport() AuditReportName {
	switch at {
	case ApplicationTypePostgres:
		return AuditReportPostgres
	case ApplicationTypeMysql:
		return AuditReportMysql
	case ApplicationTypeRedis, ApplicationTypeKeyDB, ApplicationTypeValkey, ApplicationTypeDragonfly:
		return AuditReportRedis
	case ApplicationTypeMongodb, ApplicationTypeMongos:
		return AuditReportMongodb
	case ApplicationTypeMemcached:
		return AuditReportMemcached
	case ApplicationTypeJava:
		return AuditReportJvm
	case ApplicationTypeDotNet:
		return AuditReportDotNet
	case ApplicationTypePython:
		return AuditReportPython
	case ApplicationTypeNodeJS:
		return AuditReportNodejs
	}
	return ""
}

func (at ApplicationType) IsCorootComponent() bool {
	return strings.HasPrefix(string(at), "coroot")
}

func (at ApplicationType) Name() string {
	switch {
	case at.IsCorootComponent():
		return "coroot"
	}
	return string(at)
}

func (at ApplicationType) Icon() string {
	switch {
	case strings.HasPrefix(string(at), "kube"):
		return "kubernetes"
	case at.IsCorootComponent():
		return "coroot"
	case at == ApplicationTypePgbouncer:
		return "postgres"
	case at == ApplicationTypeMongos:
		return "mongodb"
	case at == ApplicationTypeKeyDB || at == ApplicationTypeDragonfly:
		return "redis"
	case at == ApplicationTypeVictoriaMetrics || at == ApplicationTypeVictoriaLogs:
		return "victoriametrics"
	}
	return string(at)
}
