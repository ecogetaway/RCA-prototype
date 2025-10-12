package constructor

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/coroot/coroot/db"
	"github.com/coroot/coroot/model"
	"github.com/coroot/coroot/timeseries"
	"github.com/coroot/coroot/utils"
	promModel "github.com/prometheus/common/model"
	"golang.org/x/exp/maps"
	"k8s.io/klog"
)

const (
	qApplicationCustomSLI = "application_custom_sli"

	qRecordingRuleApplicationLogMessages        = "rr_application_log_messages"
	qRecordingRuleApplicationTCPSuccessful      = "rr_connection_tcp_successful"
	qRecordingRuleApplicationTCPActive          = "rr_connection_tcp_active"
	qRecordingRuleApplicationTCPFailed          = "rr_connection_tcp_failed"
	qRecordingRuleApplicationTCPConnectionTime  = "rr_connection_tcp_connection_time"
	qRecordingRuleApplicationTCPBytesSent       = "rr_connection_tcp_bytes_sent"
	qRecordingRuleApplicationTCPBytesReceived   = "rr_connection_tcp_bytes_received"
	qRecordingRuleApplicationTCPRetransmissions = "rr_connection_tcp_retransmissions"
	qRecordingRuleApplicationNetLatency         = "rr_connection_net_latency"
	qRecordingRuleApplicationL7Requests         = "rr_connection_l7_requests"
	qRecordingRuleApplicationL7Latency          = "rr_connection_l7_latency"
	qRecordingRuleApplicationTraffic            = "rr_application_traffic"
	qRecordingRuleApplicationL7Histogram        = "rr_application_l7_histogram"
	qRecordingRuleApplicationCategories         = "rr_application_categories"
	qRecordingRuleApplicationSLO                = "rr_application_slo"
)

var applicationAnnotations = maps.Keys(model.ApplicationAnnotationLabels)

var qConnectionAggregations = []string{
	qRecordingRuleApplicationTCPSuccessful,
	qRecordingRuleApplicationTCPActive,
	qRecordingRuleApplicationTCPFailed,
	qRecordingRuleApplicationTCPConnectionTime,
	qRecordingRuleApplicationTCPBytesSent,
	qRecordingRuleApplicationTCPBytesReceived,
	qRecordingRuleApplicationTCPRetransmissions,
	qRecordingRuleApplicationNetLatency,
	qRecordingRuleApplicationL7Requests,
	qRecordingRuleApplicationL7Latency,
	qRecordingRuleApplicationTraffic,
}

var (
	possibleNamespaceLabels  = []string{"namespace", "ns", "kubernetes_namespace", "kubernetes_ns", "k8s_namespace", "k8s_ns"}
	possiblePodLabels        = []string{"pod", "pod_name", "kubernetes_pod", "k8s_pod"}
	possibleDBInstanceLabels = []string{"address", "instance", "rds_instance_id", "ec_instance_id"}
)

type Query struct {
	Name   string
	Query  string
	Labels *utils.StringSet

	InstanceToInstance bool
}

func Q(name, query string, labels ...string) Query {
	ls := utils.NewStringSet(model.LabelMachineId, model.LabelSystemUuid, model.LabelContainerId, model.LabelDestination, model.LabelDestinationIP, model.LabelActualDestination)
	ls.Add(labels...)
	return Query{Name: name, Query: query, Labels: ls}
}

func qItoI(name, query string, labels ...string) Query {
	q := Q(name, query, append(labels, "app_id")...)
	q.InstanceToInstance = true
	return q
}

func qPod(name, query string, labels ...string) Query {
	return Q(name, query, slices.Concat([]string{"uid"}, labels)...)
}

func qRDS(name, query string, labels ...string) Query {
	return Q(name, query, slices.Concat([]string{"rds_instance_id"}, labels)...)
}

func qDB(name, query string, labels ...string) Query {
	return Q(name, query, slices.Concat(possibleDBInstanceLabels, possibleNamespaceLabels, possiblePodLabels, labels)...)
}

func qJVM(name, query string, labels ...string) Query {
	return Q(name, query, slices.Concat([]string{"jvm"}, labels)...)
}

func qDotNet(name, query string, labels ...string) Query {
	return Q(name, query, slices.Concat([]string{"application"}, labels)...)
}

func qFargateContainer(name, query string, labels ...string) Query {
	return Q(name, query, slices.Concat([]string{"kubernetes_io_hostname", "namespace", "pod", "container"}, labels)...)
}

func l7Req(metric string) string {
	return fmt.Sprintf(`sum by(app_id, destination, actual_destination, status) (rate(%s{app_id!=""}[$RANGE])) or rate(%s{app_id=""}[$RANGE])`, metric, metric)
}

func l7ReqWithMethod(metric string) string {
	return fmt.Sprintf(`sum by(app_id, destination, actual_destination, status, method) (rate(%s{app_id!=""}[$RANGE])) or rate(%s{app_id=""}[$RANGE])`, metric, metric)
}

func l7Latency(metric string) string {
	return fmt.Sprintf(`sum by(app_id, destination, actual_destination) (rate(%s{app_id!=""}[$RANGE])) or rate(%s{app_id=""}[$RANGE])`, metric, metric)
}

func l7Histogram(metric string) string {
	return fmt.Sprintf(`sum by(app_id, destination, actual_destination, le) (rate(%s{app_id!=""}[$RANGE])) or rate(%s{app_id=""}[$RANGE])`, metric, metric)
}

var QUERIES = []Query{
	Q("node_agent_info", `node_agent_info`, "version"),

	Q("up", `up`, "job", "instance"),

	Q("node_info", `node_info`, "hostname", "kernel_version"),
	Q("node_cloud_info", `node_cloud_info`, "provider", "region", "availability_zone", "instance_type", "instance_life_cycle"),
	Q("node_uptime_seconds", `node_uptime_second`),
	Q("node_cpu_cores", `node_resources_cpu_logical_cores`, ""),
	Q("node_cpu_usage_percent", `sum(rate(node_resources_cpu_usage_seconds_total{mode!="idle"}[$RANGE])) without(mode) /sum(rate(node_resources_cpu_usage_seconds_total[$RANGE])) without(mode)*100`),
	Q("node_cpu_usage_by_mode", `rate(node_resources_cpu_usage_seconds_total{mode!="idle"}[$RANGE]) / ignoring(mode) group_left sum(rate(node_resources_cpu_usage_seconds_total[$RANGE])) without(mode)*100`, "mode"),
	Q("node_memory_total_bytes", `node_resources_memory_total_bytes`),
	Q("node_memory_available_bytes", `node_resources_memory_available_bytes`),
	Q("node_memory_free_bytes", `node_resources_memory_free_bytes`),
	Q("node_memory_cached_bytes", `node_resources_memory_cached_bytes`),
	Q("node_disk_read_time", `rate(node_resources_disk_read_time_seconds_total[$RANGE])`, "device"),
	Q("node_disk_write_time", `rate(node_resources_disk_write_time_seconds_total[$RANGE])`, "device"),
	Q("node_disk_reads", `rate(node_resources_disk_reads_total[$RANGE])`, "device"),
	Q("node_disk_writes", `rate(node_resources_disk_writes_total[$RANGE])`, "device"),
	Q("node_disk_read_bytes", `rate(node_resources_disk_read_bytes_total[$RANGE])`, "device"),
	Q("node_disk_written_bytes", `rate(node_resources_disk_written_bytes_total[$RANGE])`, "device"),
	Q("node_disk_io_time", `rate(node_resources_disk_io_time_seconds_total[$RANGE])`, "device"),
	Q("node_net_up", `node_net_interface_up`, "interface"),
	Q("node_net_ip", `node_net_interface_ip`, "interface", "ip"),
	Q("node_net_rx_bytes", `rate(node_net_received_bytes_total[$RANGE])`, "interface"),
	Q("node_net_tx_bytes", `rate(node_net_transmitted_bytes_total[$RANGE])`, "interface"),
	Q("node_gpu_info", `node_gpu_info`, "gpu_uuid", "name"),
	Q("node_gpu_memory_total_bytes", `node_resources_gpu_memory_total_bytes`, "gpu_uuid"),
	Q("node_gpu_memory_used_bytes", `node_resources_gpu_memory_used_bytes`, "gpu_uuid"),
	Q("node_gpu_memory_utilization_percent_avg", `node_resources_gpu_memory_utilization_percent_avg`, "gpu_uuid"),
	Q("node_gpu_memory_utilization_percent_peak", `node_resources_gpu_memory_utilization_percent_peak`, "gpu_uuid"),
	Q("node_gpu_utilization_percent_avg", `node_resources_gpu_utilization_percent_avg`, "gpu_uuid"),
	Q("node_gpu_utilization_percent_peak", `node_resources_gpu_utilization_percent_peak`, "gpu_uuid"),
	Q("node_gpu_temperature_celsius", `node_resources_gpu_temperature_celsius`, "gpu_uuid"),
	Q("node_gpu_power_usage_watts", `node_resources_gpu_power_usage_watts`, "gpu_uuid"),

	Q("ip_to_fqdn", `sum by(fqdn, ip) (ip_to_fqdn)`, "ip", "fqdn"),

	Q("fargate_node_machine_cpu_cores", `machine_cpu_cores{eks_amazonaws_com_compute_type="fargate"}`, "eks_amazonaws_com_compute_type", "kubernetes_io_hostname", "topology_kubernetes_io_region", "topology_kubernetes_io_zone"),
	Q("fargate_node_machine_memory_bytes", `machine_memory_bytes{eks_amazonaws_com_compute_type="fargate"}`, "eks_amazonaws_com_compute_type", "kubernetes_io_hostname", "topology_kubernetes_io_region", "topology_kubernetes_io_zone"),

	qFargateContainer("fargate_container_spec_cpu_limit_cores", `container_spec_cpu_quota{eks_amazonaws_com_compute_type="fargate"}/container_spec_cpu_period{eks_amazonaws_com_compute_type="fargate"}`),
	qFargateContainer("fargate_container_cpu_usage_seconds", `rate(container_cpu_usage_seconds_total{eks_amazonaws_com_compute_type="fargate"}[$RANGE])`),
	qFargateContainer("fargate_container_cpu_cfs_throttled_seconds", `rate(container_cpu_cfs_throttled_seconds_total{eks_amazonaws_com_compute_type="fargate"}[$RANGE])`),
	qFargateContainer("fargate_container_spec_memory_limit_bytes", `container_spec_memory_limit_bytes{eks_amazonaws_com_compute_type="fargate"}`),
	qFargateContainer("fargate_container_memory_rss", `container_memory_rss{eks_amazonaws_com_compute_type="fargate"}`),
	qFargateContainer("fargate_container_memory_cache", `container_memory_cache{eks_amazonaws_com_compute_type="fargate"}`),
	qFargateContainer("fargate_container_oom_events_total", `container_oom_events_total{eks_amazonaws_com_compute_type="fargate"}`, "job", "instance"),

	Q("kube_node_info", `kube_node_info`, "node", "kernel_version"),
	Q("kube_service_info", `kube_service_info`, "namespace", "service", "cluster_ip"),
	Q("kube_service_spec_type", `kube_service_spec_type`, "namespace", "service", "type"),
	Q("kube_endpoint_address", `kube_endpoint_address`, "namespace", "endpoint", "ip"),
	Q("kube_service_status_load_balancer_ingress", `kube_service_status_load_balancer_ingress`, "namespace", "service", "ip"),
	Q("kube_deployment_spec_replicas", `kube_deployment_spec_replicas`, "namespace", "deployment"),
	Q("kube_daemonset_status_desired_number_scheduled", `kube_daemonset_status_desired_number_scheduled`, "namespace", "daemonset"),
	Q("kube_statefulset_replicas", `kube_statefulset_replicas`, "namespace", "statefulset"),
	Q("kube_deployment_annotations", `kube_deployment_annotations`, append(applicationAnnotations, "namespace", "deployment")...),
	Q("kube_statefulset_annotations", `kube_statefulset_annotations`, append(applicationAnnotations, "namespace", "statefulset")...),
	Q("kube_daemonset_annotations", `kube_daemonset_annotations`, append(applicationAnnotations, "namespace", "daemonset")...),
	Q("kube_cronjob_annotations", `kube_cronjob_annotations`, append(applicationAnnotations, "namespace", "cronjob")...),

	qPod("kube_pod_info", `kube_pod_info`, "namespace", "pod", "created_by_name", "created_by_kind", "node", "pod_ip", "host_ip"),
	qPod("kube_pod_annotations", hasNotEmptyLabel("kube_pod_annotations", applicationAnnotations), applicationAnnotations...),
	qPod("kube_pod_labels", `kube_pod_labels`,
		"label_postgres_operator_crunchydata_com_cluster", "label_postgres_operator_crunchydata_com_role",
		"label_cluster_name", "label_team", "label_application", "label_spilo_role",
		"label_role",
		"label_k8s_enterprisedb_io_cluster",
		"label_cnpg_io_cluster",
		"label_stackgres_io_cluster_name",
		"label_app_kubernetes_io_managed_by",
		"label_app_kubernetes_io_instance",
		"label_helm_sh_chart",
		"label_app_kubernetes_io_name",
		"label_app_kubernetes_io_component", "label_app_kubernetes_io_part_of",
	),
	qPod("kube_pod_status_phase", `kube_pod_status_phase > 0`, "phase"),
	qPod("kube_pod_status_ready", `kube_pod_status_ready{condition="true"}`),
	qPod("kube_pod_status_scheduled", `kube_pod_status_scheduled{condition="true"} > 0`),
	qPod("kube_pod_init_container_info", `kube_pod_init_container_info`, "namespace", "pod", "container"),
	qPod("kube_pod_container_resource_requests", `kube_pod_container_resource_requests`, "namespace", "pod", "container", "resource"),
	qPod("kube_pod_container_status_ready", `kube_pod_container_status_ready > 0`, "namespace", "pod", "container"),
	qPod("kube_pod_container_status_running", `kube_pod_container_status_running > 0`, "namespace", "pod", "container"),
	qPod("kube_pod_container_status_waiting", `kube_pod_container_status_waiting > 0`, "namespace", "pod", "container"),
	qPod("kube_pod_container_status_waiting_reason", `kube_pod_container_status_waiting_reason > 0`, "namespace", "pod", "container", "reason"),
	qPod("kube_pod_container_status_terminated", `kube_pod_container_status_terminated > 0`, "namespace", "pod", "container"),
	qPod("kube_pod_container_status_terminated_reason", `kube_pod_container_status_terminated_reason > 0`, "namespace", "pod", "container", "reason"),
	qPod("kube_pod_container_status_last_terminated_reason", `kube_pod_container_status_last_terminated_reason`, "namespace", "pod", "container", "reason"),

	Q("container_info", `container_info`, "image", "systemd_triggered_by"),
	Q("container_application_type", `container_application_type`, "application_type"),
	Q("container_cpu_limit", `container_resources_cpu_limit_cores`),
	Q("container_cpu_usage", `rate(container_resources_cpu_usage_seconds_total[$RANGE])`),
	Q("container_cpu_delay", `rate(container_resources_cpu_delay_seconds_total[$RANGE])`),
	Q("container_throttled_time", `rate(container_resources_cpu_throttled_seconds_total[$RANGE])`),
	Q("container_memory_limit", `container_resources_memory_limit_bytes`),
	Q("container_memory_rss", `container_resources_memory_rss_bytes`),
	Q("container_memory_cache", `container_resources_memory_cache_bytes`),
	Q("container_memory_pressure", `rate(container_resources_memory_pressure_waiting_seconds_total[$RANGE])`, "kind"),
	Q("container_oom_kills_total", `container_oom_kills_total % 10000000`, "job", "instance"),
	Q("container_restarts", `container_restarts_total % 10000000`, "job", "instance"),
	Q("container_volume_size", `container_resources_disk_size_bytes`, "mount_point", "volume", "device"),
	Q("container_volume_used", `container_resources_disk_used_bytes`, "mount_point", "volume", "device"),
	Q("container_gpu_usage_percent", `container_resources_gpu_usage_percent`, "gpu_uuid"),
	Q("container_gpu_memory_usage_percent", `container_resources_gpu_memory_usage_percent`, "gpu_uuid"),
	Q("container_net_tcp_listen_info", `container_net_tcp_listen_info`, "listen_addr", "proxy"),

	qItoI("container_net_latency", `avg by(app_id, destination_ip) (container_net_latency_seconds{app_id!=""}) or container_net_latency_seconds{app_id=""}`),
	qItoI("container_net_tcp_successful_connects", `sum by(app_id, destination, actual_destination) (rate(container_net_tcp_successful_connects_total{app_id!=""}[$RANGE])) or rate(container_net_tcp_successful_connects_total{app_id=""}[$RANGE])`),
	qItoI("container_net_tcp_failed_connects", `sum by(app_id, destination, actual_destination) (rate(container_net_tcp_failed_connects_total{app_id!=""}[$RANGE])) or rate(container_net_tcp_failed_connects_total{app_id=""}[$RANGE])`),
	qItoI("container_net_tcp_active_connections", `sum by(app_id, destination, actual_destination) (container_net_tcp_active_connections{app_id!=""}) or container_net_tcp_active_connections{app_id=""}`),
	qItoI("container_net_tcp_connection_time_seconds", `sum by(app_id, destination, actual_destination) (rate(container_net_tcp_connection_time_seconds_total{app_id!=""}[$RANGE])) or rate(container_net_tcp_connection_time_seconds_total{app_id=""}[$RANGE])`),
	qItoI("container_net_tcp_bytes_sent", `sum by(app_id, destination, actual_destination, az, region) (rate(container_net_tcp_bytes_sent_total{app_id!=""}[$RANGE])) or rate(container_net_tcp_bytes_sent_total{app_id=""}[$RANGE])`, "region", "az"),
	qItoI("container_net_tcp_bytes_received", `sum by(app_id, destination, actual_destination, az, region) (rate(container_net_tcp_bytes_received_total{app_id!=""}[$RANGE])) or rate(container_net_tcp_bytes_received_total{app_id=""}[$RANGE])`, "region", "az"),
	qItoI("container_net_tcp_retransmits", `sum by(app_id, destination, actual_destination) (rate(container_net_tcp_retransmits_total{app_id!=""}[$RANGE])) or rate(container_net_tcp_retransmits_total{app_id=""}[$RANGE])`),

	Q("container_log_messages", `container_log_messages_total % 10000000`, "level", "pattern_hash", "sample", "job", "instance"),

	qItoI("container_http_requests_count", l7Req("container_http_requests_total"), "status"),
	qItoI("container_http_requests_latency_total", l7Latency("container_http_requests_duration_seconds_total_sum")),
	qItoI("container_http_requests_histogram", l7Histogram("container_http_requests_duration_seconds_total_bucket"), "le"),
	qItoI("container_postgres_queries_count", l7Req("container_postgres_queries_total"), "status"),
	qItoI("container_postgres_queries_latency_total", l7Latency("container_postgres_queries_duration_seconds_total_sum")),
	qItoI("container_postgres_queries_histogram", l7Histogram("container_postgres_queries_duration_seconds_total_bucket"), "le"),
	qItoI("container_redis_queries_count", l7Req("container_redis_queries_total"), "status"),
	qItoI("container_redis_queries_latency_total", l7Latency("container_redis_queries_duration_seconds_total_sum")),
	qItoI("container_redis_queries_histogram", l7Histogram("container_redis_queries_duration_seconds_total_bucket"), "le"),
	qItoI("container_memcached_queries_count", l7Req("container_memcached_queries_total"), "status"),
	qItoI("container_memcached_queries_latency_total", l7Latency("container_memcached_queries_duration_seconds_total_sum")),
	qItoI("container_memcached_queries_histogram", l7Histogram("container_memcached_queries_duration_seconds_total_bucket"), "le"),
	qItoI("container_mysql_queries_count", l7Req("container_mysql_queries_total"), "status"),
	qItoI("container_mysql_queries_latency_total", l7Latency("container_mysql_queries_duration_seconds_total_sum")),
	qItoI("container_mysql_queries_histogram", l7Histogram("container_mysql_queries_duration_seconds_total_bucket"), "le"),
	qItoI("container_mongo_queries_count", l7Req("container_mongo_queries_total"), "status"),
	qItoI("container_mongo_queries_latency_total", l7Latency("container_mongo_queries_duration_seconds_total_sum")),
	qItoI("container_mongo_queries_histogram", l7Histogram("container_mongo_queries_duration_seconds_total_bucket"), "le"),
	qItoI("container_kafka_requests_count", l7Req("container_kafka_requests_total"), "status"),
	qItoI("container_kafka_requests_latency_total", l7Latency("container_kafka_requests_duration_seconds_total_sum")),
	qItoI("container_kafka_requests_histogram", l7Histogram("container_kafka_requests_duration_seconds_total_bucket"), "le"),
	qItoI("container_cassandra_queries_count", l7Req("container_cassandra_queries_total"), "status"),
	qItoI("container_cassandra_queries_latency_total", l7Latency("container_cassandra_queries_duration_seconds_total_sum")),
	qItoI("container_cassandra_queries_histogram", l7Histogram("container_cassandra_queries_duration_seconds_total_bucket"), "le"),
	qItoI("container_clickhouse_queries_count", l7Req("container_clickhouse_queries_total"), "status"),
	qItoI("container_clickhouse_queries_latency_total", l7Latency("container_clickhouse_queries_duration_seconds_total_sum")),
	qItoI("container_clickhouse_queries_histogram", l7Histogram("container_clickhouse_queries_duration_seconds_total_bucket"), "le"),
	qItoI("container_zookeeper_requests_count", l7Req("container_zookeeper_requests_total"), "status"),
	qItoI("container_zookeeper_requests_latency_total", l7Latency("container_zookeeper_requests_duration_seconds_total_sum")),
	qItoI("container_zookeeper_requests_histogram", l7Histogram("container_zookeeper_requests_duration_seconds_total_bucket"), "le"),
	qItoI("container_foundationdb_requests_count", l7Req("container_foundationdb_requests_total"), "status"),
	qItoI("container_foundationdb_requests_latency_total", l7Latency("container_foundationdb_requests_duration_seconds_total_sum")),
	qItoI("container_foundationdb_requests_histogram", l7Histogram("container_foundationdb_requests_duration_seconds_total_bucket"), "le"),
	qItoI("container_rabbitmq_messages", l7ReqWithMethod("container_rabbitmq_messages_total"), "status", "method"),
	qItoI("container_nats_messages", l7ReqWithMethod("container_nats_messages_total"), "status", "method"),

	Q("l7_requests_by_dest", "sum by(actual_destination, status) (rate(container_mongo_queries_total[$RANGE]) or rate(container_mysql_queries_total[$RANGE]))", "status"),
	Q("l7_total_latency_by_dest", "sum by(actual_destination) (rate(container_mongo_queries_duration_seconds_total_sum[$RANGE]) or rate(container_mysql_queries_duration_seconds_total_sum[$RANGE]))"),

	Q("container_dns_requests_total", `sum by(app_id, request_type, domain, status) (rate(container_dns_requests_total{app_id!=""}[$RANGE])) or rate(container_dns_requests_total{app_id=""}[$RANGE])`, "app_id", "request_type", "domain", "status"),
	Q("container_dns_requests_latency", `sum by(app_id, le) (rate(container_dns_requests_duration_seconds_total_bucket{app_id!=""}[$RANGE])) or rate(container_dns_requests_duration_seconds_total_bucket{app_id=""}[$RANGE]) `, "app_id", "le"),

	Q("aws_discovery_error", `aws_discovery_error`, "error"),
	qRDS("aws_rds_info", `aws_rds_info`, "cluster_id", "ipv4", "port", "engine", "engine_version", "instance_type", "storage_type", "region", "availability_zone", "multi_az"),
	qRDS("aws_rds_status", `aws_rds_status`, "status"),
	qRDS("aws_rds_cpu_cores", `aws_rds_cpu_cores`),
	qRDS("aws_rds_cpu_usage_percent", `aws_rds_cpu_usage_percent`, "mode"),
	qRDS("aws_rds_memory_total_bytes", `aws_rds_memory_total_bytes`),
	qRDS("aws_rds_memory_cached_bytes", `aws_rds_memory_cached_bytes`),
	qRDS("aws_rds_memory_free_bytes", `aws_rds_memory_free_bytes`),
	qRDS("aws_rds_storage_provisioned_iops", `aws_rds_storage_provisioned_iops`),
	qRDS("aws_rds_allocated_storage_gibibytes", `aws_rds_allocated_storage_gibibytes`),
	qRDS("aws_rds_fs_total_bytes", `aws_rds_fs_total_bytes{mount_point="/rdsdbdata"}`),
	qRDS("aws_rds_fs_used_bytes", `aws_rds_fs_used_bytes{mount_point="/rdsdbdata"}`),
	qRDS("aws_rds_io_util_percent", `aws_rds_io_util_percent`, "device"),
	qRDS("aws_rds_io_ops_per_second", `aws_rds_io_ops_per_second`, "device", "operation"),
	qRDS("aws_rds_io_await_seconds", `aws_rds_io_await_seconds`, "device"),
	qRDS("aws_rds_net_rx_bytes_per_second", `aws_rds_net_rx_bytes_per_second`, "interface"),
	qRDS("aws_rds_net_tx_bytes_per_second", `aws_rds_net_tx_bytes_per_second`, "interface"),
	qRDS("aws_rds_log_messages_total", `aws_rds_log_messages_total % 10000000`, "level", "pattern_hash", "sample", "job", "instance"),

	Q("aws_elasticache_info", `aws_elasticache_info`, "ec_instance_id", "cluster_id", "ipv4", "port", "engine", "engine_version", "instance_type", "region", "availability_zone"),
	Q("aws_elasticache_status", `aws_elasticache_status`, "ec_instance_id", "status"),

	qDB("pg_up", `pg_up`),
	qDB("pg_scrape_error", `pg_scrape_error`, "error", "warning"),
	qDB("pg_info", `pg_info`, "server_version"),
	qDB("pg_setting", `pg_setting`, "name", "unit"),
	qDB("pg_connections", `pg_connections{db!="postgres"}`, "db", "user", "state", "query", "wait_event_type"),
	qDB("pg_lock_awaiting_queries", `pg_lock_awaiting_queries`, "db", "user", "blocking_query"),
	qDB("pg_latency_seconds", `pg_latency_seconds`, "summary"),
	qDB("pg_top_query_calls_per_second", `pg_top_query_calls_per_second`, "db", "user", "query"),
	qDB("pg_top_query_time_per_second", `pg_top_query_time_per_second`, "db", "user", "query"),
	qDB("pg_top_query_io_time_per_second", `pg_top_query_io_time_per_second`, "db", "user", "query"),
	qDB("pg_db_queries_per_second", `pg_db_queries_per_second`, "db"),
	qDB("pg_wal_current_lsn", `pg_wal_current_lsn`),
	qDB("pg_wal_receive_lsn", `pg_wal_receive_lsn`),
	qDB("pg_wal_reply_lsn", `pg_wal_reply_lsn`),

	qDB("mysql_up", `mysql_up`),
	qDB("mysql_scrape_error", `mysql_scrape_error`, "error", "warning"),
	qDB("mysql_info", `mysql_info`, "server_uuid", "server_version"),
	qDB("mysql_top_query_calls_per_second", `mysql_top_query_calls_per_second`, "schema", "query"),
	qDB("mysql_top_query_time_per_second", `mysql_top_query_time_per_second`, "schema", "query"),
	qDB("mysql_top_query_lock_time_per_second", `mysql_top_query_lock_time_per_second`, "schema", "query"),
	qDB("mysql_replication_io_status", `mysql_replication_io_status`, "source_server_uuid", "last_error", "state"),
	qDB("mysql_replication_sql_status", `mysql_replication_sql_status`, "source_server_uuid", "last_error", "state"),
	qDB("mysql_replication_lag_seconds", `mysql_replication_lag_seconds`, "source_server_uuid"),
	qDB("mysql_connections_max", `mysql_connections_max`),
	qDB("mysql_connections_current", `mysql_connections_current`),
	qDB("mysql_connections_total", `rate(mysql_connections_total[$RANGE])`),
	qDB("mysql_connections_aborted_total", `rate(mysql_connections_aborted_total[$RANGE])`),
	qDB("mysql_traffic_received_bytes_total", `rate(mysql_traffic_received_bytes_total[$RANGE])`),
	qDB("mysql_traffic_sent_bytes_total", `rate(mysql_traffic_sent_bytes_total[$RANGE])`),
	qDB("mysql_queries_total", `rate(mysql_queries_total[$RANGE])`),
	qDB("mysql_slow_queries_total", `rate(mysql_slow_queries_total[$RANGE])`),
	qDB("mysql_top_table_io_wait_time_per_second", `mysql_top_table_io_wait_time_per_second`, "schema", "table", "operation"),

	qDB("redis_up", `redis_up`),
	qDB("redis_scrape_error", `redis_exporter_last_scrape_error`, "err"),
	qDB("redis_instance_info", `redis_instance_info`, "redis_version", "role"),
	qDB("redis_commands_duration_seconds_total", `rate(redis_commands_duration_seconds_total[$RANGE])`, "cmd"),
	qDB("redis_commands_total", `rate(redis_commands_total[$RANGE])`, "cmd"),

	qDB("mongo_up", `mongo_up`),
	qDB("mongo_scrape_error", `mongo_scrape_error`, "error", "warning"),
	qDB("mongo_info", `mongo_info`, "server_version"),
	qDB("mongo_rs_status", `mongo_rs_status`, "rs", "role"),
	qDB("mongo_rs_last_applied_timestamp_ms", `timestamp(mongo_rs_last_applied_timestamp_ms) - mongo_rs_last_applied_timestamp_ms/1000`),

	qDB("memcached_up", `memcached_up`),
	qDB("memcached_version", `memcached_version`, "version"),
	qDB("memcached_limit_bytes", `memcached_limit_bytes`),
	qDB("memcached_items_evicted_total", `rate(memcached_items_evicted_total[$RANGE])`),
	qDB("memcached_commands_total", `rate(memcached_commands_total[$RANGE])`, "command", "status"),

	qJVM("container_jvm_info", `container_jvm_info`, "java_version"),
	qJVM("container_jvm_heap_size_bytes", `container_jvm_heap_size_bytes`),
	qJVM("container_jvm_heap_used_bytes", `container_jvm_heap_used_bytes`),
	qJVM("container_jvm_gc_time_seconds", `rate(container_jvm_gc_time_seconds[$RANGE])`, "gc"),
	qJVM("container_jvm_safepoint_time_seconds", `rate(container_jvm_safepoint_time_seconds[$RANGE])`),
	qJVM("container_jvm_safepoint_sync_time_seconds", `rate(container_jvm_safepoint_sync_time_seconds[$RANGE])`),

	qDotNet("container_dotnet_info", `container_dotnet_info`, "runtime_version"),
	qDotNet("container_dotnet_memory_allocated_bytes_total", `rate(container_dotnet_memory_allocated_bytes_total[$RANGE])`),
	qDotNet("container_dotnet_exceptions_total", `rate(container_dotnet_exceptions_total[$RANGE])`),
	qDotNet("container_dotnet_memory_heap_size_bytes", `container_dotnet_memory_heap_size_bytes`, "generation"),
	qDotNet("container_dotnet_gc_count_total", `rate(container_dotnet_gc_count_total[$RANGE])`, "generation"),
	qDotNet("container_dotnet_heap_fragmentation_percent", `container_dotnet_heap_fragmentation_percent`),
	qDotNet("container_dotnet_monitor_lock_contentions_total", `rate(container_dotnet_monitor_lock_contentions_total[$RANGE])`),
	qDotNet("container_dotnet_thread_pool_completed_items_total", `rate(container_dotnet_thread_pool_completed_items_total[$RANGE])`),
	qDotNet("container_dotnet_thread_pool_queue_length", `container_dotnet_thread_pool_queue_length`),
	qDotNet("container_dotnet_thread_pool_size", `container_dotnet_thread_pool_size`),

	Q("container_python_thread_lock_wait_time_seconds", `rate(container_python_thread_lock_wait_time_seconds[$RANGE])`),
	Q("container_nodejs_event_loop_blocked_time_seconds", `rate(container_nodejs_event_loop_blocked_time_seconds_total[$RANGE])`),

	qPod("fluxcd_git_repository_info", `fluxcd_git_repository_info`, "name", "namespace", "suspended", "url", "interval"),
	qPod("fluxcd_git_repository_status", `fluxcd_git_repository_status`, "name", "namespace", "type", "reason"),
	qPod("fluxcd_oci_repository_info", `fluxcd_oci_repository_info`, "name", "namespace", "suspended", "url", "interval"),
	qPod("fluxcd_oci_repository_status", `fluxcd_oci_repository_status`, "name", "namespace", "type", "reason"),
	qPod("fluxcd_helm_repository_info", `fluxcd_helm_repository_info`, "name", "namespace", "suspended", "url", "interval"),
	qPod("fluxcd_helm_repository_status", `fluxcd_helm_repository_status`, "name", "namespace", "type", "reason"),
	qPod("fluxcd_helm_chart_info", `fluxcd_helm_chart_info`, "name", "namespace", "suspended", "chart", "interval", "version", "source_kind", "source_name", "source_namespace"),
	qPod("fluxcd_helm_chart_status", `fluxcd_helm_chart_status`, "name", "namespace", "type", "reason"),
	qPod("fluxcd_helm_release_info", `fluxcd_helm_release_info`, "name", "namespace", "suspended", "chart", "interval", "version", "source_kind", "source_name", "source_namespace", "chart_ref_kind", "chart_ref_name", "chart_ref_namespace", "target_namespace"),
	qPod("fluxcd_helm_release_status", `fluxcd_helm_release_status`, "name", "namespace", "type", "reason"),
	qPod("fluxcd_kustomization_info", `fluxcd_kustomization_info`, "name", "namespace", "suspended", "interval", "path", "source_kind", "source_name", "source_namespace", "target_namespace", "last_applied_revision", "last_attempted_revision"),
	qPod("fluxcd_kustomization_status", `fluxcd_kustomization_status`, "name", "namespace", "type", "reason"),
	qPod("fluxcd_kustomization_dependency_info", `fluxcd_kustomization_dependency_info`, "name", "namespace", "depends_on_name", "depends_on_namespace"),
	qPod("fluxcd_kustomization_inventory_entry_info", `fluxcd_kustomization_inventory_entry_info`, "name", "namespace", "entry_id"),
	qPod("fluxcd_resourceset_info", `fluxcd_resourceset_info`, "name", "namespace", "last_applied_revision"),
	qPod("fluxcd_resourceset_status", `fluxcd_resourceset_status`, "name", "namespace", "type", "reason"),
	qPod("fluxcd_resourceset_dependency_info", `fluxcd_resourceset_dependency_info`, "name", "namespace", "depends_on_name", "depends_on_namespace", "depends_on_kind"),
	qPod("fluxcd_resourceset_inventory_entry_info", `fluxcd_resourceset_inventory_entry_info`, "name", "namespace", "entry_id"),
}

var RecordingRules = map[string]func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues{
	qRecordingRuleApplicationLogMessages: func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		var res []*model.MetricValues
		for _, app := range w.Applications {
			appId := app.Id.String()
			for severity, msgs := range app.LogMessages {
				if len(msgs.Patterns) == 0 {
					if msgs.Messages.Reduce(timeseries.NanSum) > 0 {
						ls := model.Labels{"application": appId, "level": severity.String()}
						res = append(res, &model.MetricValues{Labels: ls, LabelsHash: promModel.LabelsToSignature(ls), Values: msgs.Messages})
					}
				} else {
					for _, pattern := range msgs.Patterns {
						if pattern.Messages.Reduce(timeseries.NanSum) > 0 {
							ls := model.Labels{"application": appId, "level": severity.String()}
							ls["multiline"] = fmt.Sprintf("%t", pattern.Multiline)
							ls["similar"] = strings.Join(pattern.SimilarPatternHashes.Items(), " ")
							ls["sample"] = pattern.Sample
							ls["words"] = pattern.Pattern.String()
							res = append(res, &model.MetricValues{Labels: ls, LabelsHash: promModel.LabelsToSignature(ls), Values: pattern.Messages})
						}
					}
				}
			}
		}
		return res
	},
	qRecordingRuleApplicationTCPSuccessful: func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		return aggConnections(w, func(c *model.Connection) *timeseries.TimeSeries { return c.SuccessfulConnections })
	},
	qRecordingRuleApplicationTCPActive: func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		return aggConnections(w, func(c *model.Connection) *timeseries.TimeSeries { return c.Active })
	},
	qRecordingRuleApplicationTCPFailed: func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		return aggConnections(w, func(c *model.Connection) *timeseries.TimeSeries { return c.FailedConnections })
	},
	qRecordingRuleApplicationTCPConnectionTime: func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		return aggConnections(w, func(c *model.Connection) *timeseries.TimeSeries { return c.ConnectionTime })
	},
	qRecordingRuleApplicationTCPBytesSent: func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		return aggConnections(w, func(c *model.Connection) *timeseries.TimeSeries { return c.BytesSent })
	},
	qRecordingRuleApplicationTCPBytesReceived: func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		return aggConnections(w, func(c *model.Connection) *timeseries.TimeSeries { return c.BytesReceived })
	},
	qRecordingRuleApplicationTCPRetransmissions: func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		return aggConnections(w, func(c *model.Connection) *timeseries.TimeSeries { return c.Retransmissions })
	},

	qRecordingRuleApplicationNetLatency: func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		var res []*model.MetricValues

		for _, app := range w.Applications {
			byDest := map[*model.Application][]*timeseries.TimeSeries{}
			for _, instance := range app.Instances {
				for _, u := range instance.Upstreams {
					dest := u.RemoteApplication()
					if dest == nil {
						continue
					}
					if !u.Rtt.IsEmpty() {
						byDest[dest] = append(byDest[dest], u.Rtt)
					}
				}
			}
			appId := app.Id.String()
			for dest, rtts := range byDest {
				sum := timeseries.NewAggregate(timeseries.NanSum).Add(rtts...).Get()
				count := timeseries.NewAggregate(timeseries.NanSum)
				for _, rtt := range rtts {
					count.Add(rtt.Map(timeseries.Defined))
				}
				ls := model.Labels{"app": appId, "dest": dest.Id.String(), "agg": "avg"}
				res = append(res, &model.MetricValues{Labels: ls, LabelsHash: promModel.LabelsToSignature(ls), Values: timeseries.Div(sum, count.Get())})
			}
		}
		return res
	},

	qRecordingRuleApplicationL7Requests: func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		var res []*model.MetricValues
		type key struct {
			status   string
			protocol model.Protocol
			dest     *model.Application
		}

		for _, app := range w.Applications {
			byProtoStatus := map[key]*timeseries.Aggregate{}
			for _, instance := range app.Instances {
				for _, u := range instance.Upstreams {
					dest := u.RemoteApplication()
					if dest == nil {
						continue
					}
					for proto, byStatus := range u.RequestsCount {
						for status, ts := range byStatus {
							k := key{dest: dest, status: status, protocol: proto}
							agg := byProtoStatus[k]
							if agg == nil {
								agg = timeseries.NewAggregate(timeseries.NanSum)
								byProtoStatus[k] = agg
							}
							agg.Add(ts)

						}
					}
				}
			}
			appId := app.Id.String()
			for k, agg := range byProtoStatus {
				ts := agg.Get()
				if !ts.IsEmpty() {
					ls := model.Labels{"app": appId, "dest": k.dest.Id.String(), "proto": string(k.protocol), "status": k.status}
					res = append(res, &model.MetricValues{Labels: ls, LabelsHash: promModel.LabelsToSignature(ls), Values: ts})
				}
			}
		}
		return res
	},

	qRecordingRuleApplicationL7Latency: func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		var res []*model.MetricValues
		type key struct {
			protocol model.Protocol
			dest     *model.Application
		}

		for _, app := range w.Applications {
			byProto := map[key]*timeseries.Aggregate{}
			for _, instance := range app.Instances {
				for _, u := range instance.Upstreams {
					dest := u.RemoteApplication()
					if dest == nil {
						continue
					}
					for proto, ts := range u.RequestsLatency {
						k := key{dest: dest, protocol: proto}
						agg := byProto[k]
						if agg == nil {
							agg = timeseries.NewAggregate(timeseries.NanSum)
							byProto[k] = agg
						}
						agg.Add(ts)

					}
				}
			}
			appId := app.Id.String()
			for k, agg := range byProto {
				ts := agg.Get()
				if !ts.IsEmpty() {
					ls := model.Labels{"app": appId, "dest": k.dest.Id.String(), "proto": string(k.protocol)}
					res = append(res, &model.MetricValues{Labels: ls, LabelsHash: promModel.LabelsToSignature(ls), Values: ts})
				}
			}
		}
		return res
	},

	qRecordingRuleApplicationL7Histogram: func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		var res []*model.MetricValues
		type key struct {
			le   float32
			dest *model.Application
		}
		sum := map[key]*timeseries.Aggregate{}
		for _, app := range w.Applications {
			for _, instance := range app.Instances {
				for _, u := range instance.Upstreams {
					dest := u.RemoteApplication()
					if dest == nil {
						continue
					}
					if !dest.Category.Auxiliary() && app.Category.Auxiliary() {
						continue
					}
					for _, byLe := range u.RequestsHistogram {
						for le, ts := range byLe {
							k := key{dest: dest, le: le}
							agg := sum[k]
							if agg == nil {
								agg = timeseries.NewAggregate(timeseries.NanSum)
								sum[k] = agg
							}
							agg.Add(ts)
						}
					}
				}
			}
		}
		for k, agg := range sum {
			ts := agg.Get()
			if !ts.IsEmpty() {
				ls := model.Labels{"app": k.dest.Id.String(), "le": fmt.Sprintf("%f", k.le)}
				res = append(res, &model.MetricValues{Labels: ls, LabelsHash: promModel.LabelsToSignature(ls), Values: ts})
			}
		}
		return res
	},

	qRecordingRuleApplicationTraffic: func(db *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		var res []*model.MetricValues

		for _, app := range w.Applications {
			appId := app.Id.String()
			if ts := app.TrafficStats.InternetEgress.Get(); !ts.IsEmpty() {
				ls := model.Labels{"app": appId, "kind": string(model.TrafficKindInternetEgress)}
				res = append(res, &model.MetricValues{Labels: ls, LabelsHash: promModel.LabelsToSignature(ls), Values: ts})
			}
			if ts := app.TrafficStats.CrossAZEgress.Get(); !ts.IsEmpty() {
				ls := model.Labels{"app": appId, "kind": string(model.TrafficKindCrossAZEgress)}
				res = append(res, &model.MetricValues{Labels: ls, LabelsHash: promModel.LabelsToSignature(ls), Values: ts})
			}
			if ts := app.TrafficStats.CrossAZIngress.Get(); !ts.IsEmpty() {
				ls := model.Labels{"app": appId, "kind": string(model.TrafficKindCrossAZIngress)}
				res = append(res, &model.MetricValues{Labels: ls, LabelsHash: promModel.LabelsToSignature(ls), Values: ts})
			}
		}
		return res
	},

	qRecordingRuleApplicationCategories: func(database *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		var needSave bool
		for _, app := range w.Applications {
			if _, ok := p.Settings.ApplicationCategorySettings[app.Category]; !ok {
				if p.Settings.ApplicationCategorySettings == nil {
					p.Settings.ApplicationCategorySettings = map[model.ApplicationCategory]*db.ApplicationCategorySettings{}
				}
				p.Settings.ApplicationCategorySettings[app.Category] = nil
				needSave = true
			}
		}
		if needSave {
			if err := database.SaveProjectSettings(p); err != nil {
				klog.Errorln("failed to save project settings:", err)
			}
		}
		return nil
	},

	qRecordingRuleApplicationSLO: func(database *db.DB, p *db.Project, w *model.World) []*model.MetricValues {
		for _, app := range w.Applications {
			updateAvailabilitySLOFromAnnotations(database, p, w, app)
			updateLatencySLOFromAnnotations(database, p, w, app)
		}
		return nil
	},
}

func aggConnections(w *model.World, tsF func(c *model.Connection) *timeseries.TimeSeries) []*model.MetricValues {
	var res []*model.MetricValues
	for _, app := range w.Applications {
		byDest := map[*model.Application]*timeseries.Aggregate{}
		for _, instance := range app.Instances {
			for _, u := range instance.Upstreams {
				dest := u.RemoteApplication()
				if dest == nil {
					continue
				}
				agg := byDest[dest]
				if agg == nil {
					agg = timeseries.NewAggregate(timeseries.NanSum)
					byDest[dest] = agg
				}
				agg.Add(tsF(u))
			}
		}
		appId := app.Id.String()
		for dest, agg := range byDest {
			ts := agg.Get()
			if !ts.IsEmpty() {
				ls := model.Labels{"app": appId, "dest": dest.Id.String()}
				res = append(res, &model.MetricValues{Labels: ls, LabelsHash: promModel.LabelsToSignature(ls), Values: ts})
			}
		}
	}
	return res
}

func updateAvailabilitySLOFromAnnotations(database *db.DB, p *db.Project, w *model.World, app *model.Application) {
	objectiveStr := app.GetAnnotation(model.ApplicationAnnotationSLOAvailabilityObjective)
	cfg, _ := w.CheckConfigs.GetAvailability(app.Id)
	if objectiveStr == "" {
		return
	}
	cfgSaved := cfg
	cfg.Source = model.CheckConfigSourceKubernetesAnnotations
	cfg.Custom = false
	cfg.Error = ""
	objective, err := parseObjective(objectiveStr)
	if err != nil {
		cfg.Error = fmt.Sprintf("Invalid annotation 'coroot.com/slo-availability-objective': %s", err)
	}
	if cfg.Error != "" {
		cfg.ObjectivePercentage = 0 // disable
	} else {
		cfg.ObjectivePercentage = objective
	}
	if cfg == cfgSaved {
		return
	}
	if err = database.SaveCheckConfig(p.Id, app.Id, model.Checks.SLOAvailability.Id, []model.CheckConfigSLOAvailability{cfg}); err != nil {
		klog.Errorln(err)
	}
}

func updateLatencySLOFromAnnotations(database *db.DB, p *db.Project, w *model.World, app *model.Application) {
	objectiveStr := app.GetAnnotation(model.ApplicationAnnotationSLOLatencyObjective)
	thresholdStr := app.GetAnnotation(model.ApplicationAnnotationSLOLatencyThreshold)
	if objectiveStr == "" && thresholdStr == "" {
		return
	}
	cfg, _ := w.CheckConfigs.GetLatency(app.Id, app.Category)
	cfgSaved := cfg
	cfg.Source = model.CheckConfigSourceKubernetesAnnotations
	cfg.Custom = false
	cfg.Error = ""
	var err error
	objective := cfg.ObjectivePercentage
	threshold := cfg.ObjectiveBucket
	if objectiveStr != "" {
		objective, err = parseObjective(objectiveStr)
		if err != nil {
			cfg.Error = fmt.Sprintf("Invalid annotation 'coroot.com/slo-latency-objective': %s", err)
		}
	}
	if objective > 0 && thresholdStr != "" {
		threshold, err = parseThreshold(thresholdStr)
		if err != nil && cfg.Error == "" {
			cfg.Error = fmt.Sprintf("Invalid annotation 'coroot.com/slo-latency-threshold': %s", err)
		}
	}
	if cfg.Error != "" {
		cfg.ObjectivePercentage = 0 // disable
	} else {
		cfg.ObjectivePercentage = objective
		cfg.ObjectiveBucket = threshold
	}
	if cfg == cfgSaved {
		return
	}
	if err = database.SaveCheckConfig(p.Id, app.Id, model.Checks.SLOLatency.Id, []model.CheckConfigSLOLatency{cfg}); err != nil {
		klog.Errorln(err)
	}
}

func hasNotEmptyLabel(metricName string, labelNames []string) string {
	var parts []string
	for _, labelName := range labelNames {
		parts = append(parts, fmt.Sprintf(`%s{%s != ""}`, metricName, labelName))
	}
	return strings.Join(parts, " or ")
}

func parseObjective(s string) (float32, error) {
	s = strings.TrimSpace(strings.TrimRight(strings.TrimSpace(s), "%"))
	v, err := strconv.ParseFloat(s, 32)
	return float32(v), err
}

func parseThreshold(s string) (float32, error) {
	d, err := time.ParseDuration(strings.TrimSpace(s))
	if err != nil {
		return 0, err
	}
	v := model.RoundUpToDefaultBucket(float32(d.Seconds()))
	return v, err
}
