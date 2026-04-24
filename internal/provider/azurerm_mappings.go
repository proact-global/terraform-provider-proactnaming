// Copyright (c) HashiCorp, Inc.

package provider

import (
	"strings"
)

// azurermResourceTypeMap maps lowercase AzureNamingTool resource paths to the
// corresponding azurerm Terraform provider resource type name(s).
//
// Keys use the format:
//   - "namespace/resourcetype"           (lowercase, no "Microsoft." prefix)
//   - "namespace/resourcetype|property"  (lowercase, for property-specific disambiguation)
//
// Values are slices of azurerm resource type names.
var azurermResourceTypeMap = map[string][]string{
	// AnalysisServices
	"analysisservices/servers": {"azurerm_analysis_services_server"},

	// ApiManagement
	"apimanagement/service": {"azurerm_api_management"},

	// AppConfiguration
	"appconfiguration/configurationstores": {"azurerm_app_configuration"},

	// Authorization
	"authorization/locks":               {"azurerm_management_lock"},
	"authorization/policyassignments":   {"azurerm_policy_assignment"},
	"authorization/policydefinitions":   {"azurerm_policy_definition"},
	"authorization/policysetdefinitions": {"azurerm_policy_set_definition"},

	// Automation
	"automation/automationaccounts":           {"azurerm_automation_account"},
	"automation/automationaccounts/runbooks":  {"azurerm_automation_runbook"},
	"automation/automationaccounts/schedules": {"azurerm_automation_schedule"},
	"automation/automationaccounts/variables": {
		"azurerm_automation_variable_bool",
		"azurerm_automation_variable_datetime",
		"azurerm_automation_variable_int",
		"azurerm_automation_variable_string",
	},
	"automation/automationaccounts/webhooks": {"azurerm_automation_webhook"},

	// Batch
	"batch/batchaccounts":              {"azurerm_batch_account"},
	"batch/batchaccounts/applications": {"azurerm_batch_application"},
	"batch/batchaccounts/certificates": {"azurerm_batch_certificate"},
	"batch/batchaccounts/pools":        {"azurerm_batch_pool"},

	// Cache
	"cache/redis":              {"azurerm_redis_cache"},
	"cache/redis/firewallrules": {"azurerm_redis_firewall_rule"},

	// Cdn
	"cdn/profiles":           {"azurerm_cdn_profile"},
	"cdn/profiles/endpoints": {"azurerm_cdn_endpoint"},

	// CognitiveServices
	"cognitiveservices/accounts": {"azurerm_cognitive_account"},

	// Compute
	"compute/availabilitysets":    {"azurerm_availability_set"},
	"compute/diskencryptionsets":  {"azurerm_disk_encryption_set"},
	"compute/disks":               {"azurerm_managed_disk"},
	"compute/disks|os disk":       {"azurerm_managed_disk"},
	"compute/disks|data disk":     {"azurerm_managed_disk"},
	"compute/galleries":           {"azurerm_shared_image_gallery"},
	"compute/galleries/images":    {"azurerm_shared_image"},
	"compute/galleries/images/versions": {"azurerm_shared_image_version"},
	"compute/images":              {"azurerm_image"},
	"compute/snapshots":           {"azurerm_snapshot"},
	"compute/virtualmachines": {
		"azurerm_linux_virtual_machine",
		"azurerm_windows_virtual_machine",
	},
	"compute/virtualmachines|linux":   {"azurerm_linux_virtual_machine"},
	"compute/virtualmachines|windows": {"azurerm_windows_virtual_machine"},
	"compute/virtualmachinescalesets": {
		"azurerm_linux_virtual_machine_scale_set",
		"azurerm_windows_virtual_machine_scale_set",
	},
	"compute/virtualmachinescalesets|linux":   {"azurerm_linux_virtual_machine_scale_set"},
	"compute/virtualmachinescalesets|windows": {"azurerm_windows_virtual_machine_scale_set"},

	// ContainerInstance
	"containerinstance/containergroups": {"azurerm_container_group"},

	// ContainerRegistry
	"containerregistry/registries":          {"azurerm_container_registry"},
	"containerregistry/registries/tasks":    {"azurerm_container_registry_task"},
	"containerregistry/registries/tokens":   {"azurerm_container_registry_token"},
	"containerregistry/registries/webhooks": {"azurerm_container_registry_webhook"},

	// ContainerService
	"containerservice/managedclusters": {"azurerm_kubernetes_cluster"},

	// DataBox
	"databox/jobs": {"azurerm_databox_job"},

	// Databricks
	"databricks/workspaces": {"azurerm_databricks_workspace"},

	// DataFactory
	"datafactory/factories":                    {"azurerm_data_factory"},
	"datafactory/factories/dataflows":          {"azurerm_data_factory_data_flow"},
	"datafactory/factories/pipelines":          {"azurerm_data_factory_pipeline"},
	"datafactory/factories/integrationruntimes": {
		"azurerm_data_factory_integration_runtime_azure",
		"azurerm_data_factory_integration_runtime_azure_ssis",
		"azurerm_data_factory_integration_runtime_self_hosted",
	},

	// DataLakeAnalytics
	"datalakeanalytics/accounts": {"azurerm_data_lake_analytics_account"},

	// DataLakeStore
	"datalakestore/accounts": {"azurerm_data_lake_store"},

	// DataMigration
	"datamigration/services": {"azurerm_database_migration_service"},

	// DBforMariaDB
	"dbformariadb/servers":           {"azurerm_mariadb_server"},
	"dbformariadb/servers/databases": {"azurerm_mariadb_database"},

	// DBforMySQL
	"dbformysql/servers": {
		"azurerm_mysql_server",
		"azurerm_mysql_flexible_server",
	},
	"dbformysql/servers/databases": {"azurerm_mysql_database", "azurerm_mysql_flexible_database"},

	// DBforPostgreSQL
	"dbforpostgresql/servers": {
		"azurerm_postgresql_server",
		"azurerm_postgresql_flexible_server",
	},
	"dbforpostgresql/servers/databases": {"azurerm_postgresql_database", "azurerm_postgresql_flexible_database"},

	// Devices
	"devices/iothubs":               {"azurerm_iothub"},
	"devices/provisioningservices":  {"azurerm_iothub_dps"},

	// DevTestLab
	"devtestlab/labs": {"azurerm_dev_test_lab"},

	// DocumentDB
	"documentdb/databaseaccounts": {"azurerm_cosmosdb_account"},

	// EventGrid
	"eventgrid/domains": {"azurerm_eventgrid_domain"},
	"eventgrid/domains/topics": {"azurerm_eventgrid_domain_topic"},
	"eventgrid/topics":  {"azurerm_eventgrid_topic"},

	// EventHub
	"eventhub/clusters":              {"azurerm_eventhub_cluster"},
	"eventhub/namespaces":            {"azurerm_eventhub_namespace"},
	"eventhub/namespaces/eventhubs":  {"azurerm_eventhub"},

	// HDInsight
	"hdinsight/clusters": {
		"azurerm_hdinsight_hadoop_cluster",
		"azurerm_hdinsight_hbase_cluster",
		"azurerm_hdinsight_kafka_cluster",
		"azurerm_hdinsight_spark_cluster",
	},
	"hdinsight/clusters|spark cluster":       {"azurerm_hdinsight_spark_cluster"},
	"hdinsight/clusters|storm cluster":       {"azurerm_hdinsight_storm_cluster"},
	"hdinsight/clusters|ml services cluster": {"azurerm_hdinsight_ml_services_cluster"},
	"hdinsight/clusters|hadoop cluster":      {"azurerm_hdinsight_hadoop_cluster"},
	"hdinsight/clusters|hbase cluster":       {"azurerm_hdinsight_hbase_cluster"},
	"hdinsight/clusters|kafka cluster":       {"azurerm_hdinsight_kafka_cluster"},

	// Insights (Azure Monitor / Application Insights)
	"insights/actiongroups":        {"azurerm_monitor_action_group"},
	"insights/activitylogalerts":   {"azurerm_monitor_activity_log_alert"},
	"insights/components":          {"azurerm_application_insights"},
	"insights/metricalerts":        {"azurerm_monitor_metric_alert"},
	"insights/scheduledqueryrules": {"azurerm_monitor_scheduled_query_rules_alert"},

	// IoTCentral
	"iotcentral/iotapps": {"azurerm_iotcentral_application"},

	// KeyVault
	"keyvault/vaults":         {"azurerm_key_vault"},
	"keyvault/vaults/secrets": {"azurerm_key_vault_secret"},

	// Kusto (Azure Data Explorer)
	"kusto/clusters":           {"azurerm_kusto_cluster"},
	"kusto/clusters/databases": {"azurerm_kusto_database"},

	// Logic
	"logic/workflows":                    {"azurerm_logic_app_workflow"},
	"logic/integrationaccounts":          {"azurerm_logic_app_integration_account"},
	"logic/integrationserviceenvironments": {"azurerm_integration_service_environment"},

	// MachineLearningServices
	"machinelearningservices/workspaces":         {"azurerm_machine_learning_workspace"},
	"machinelearningservices/workspaces/computes": {"azurerm_machine_learning_compute_cluster"},

	// ManagedIdentity
	"managedidentity/userassignedidentities": {"azurerm_user_assigned_identity"},

	// Management
	"management/managementgroups": {"azurerm_management_group"},

	// Maps
	"maps/accounts": {"azurerm_maps_account"},

	// Media
	"media/mediaservices": {"azurerm_media_services_account"},

	// Network
	"network/applicationgateways":       {"azurerm_application_gateway"},
	"network/applicationsecuritygroups": {"azurerm_application_security_group"},
	"network/azurefirewalls":            {"azurerm_firewall"},
	"network/bastionhosts":              {"azurerm_bastion_host"},
	"network/connections":               {"azurerm_virtual_network_gateway_connection"},
	"network/dnszones":                  {"azurerm_dns_zone"},
	"network/expressroutecircuits":      {"azurerm_express_route_circuit"},
	"network/firewallpolicies":          {"azurerm_firewall_policy"},
	"network/frontdoors":                {"azurerm_frontdoor"},
	"network/frontdoorwebapplicationfirewallpolicies": {"azurerm_frontdoor_firewall_policy"},
	"network/loadbalancers":             {"azurerm_lb"},
	"network/loadbalancers|internal":    {"azurerm_lb"},
	"network/loadbalancers|external":    {"azurerm_lb"},
	"network/localnetworkgateways":      {"azurerm_local_network_gateway"},
	"network/networkinterfaces":         {"azurerm_network_interface"},
	"network/networksecuritygroups":     {"azurerm_network_security_group"},
	"network/networkwatchers":           {"azurerm_network_watcher"},
	"network/privatednszones":           {"azurerm_private_dns_zone"},
	"network/privatelinkservices":       {"azurerm_private_link_service"},
	"network/publicipaddresses":         {"azurerm_public_ip"},
	"network/publicipprefixes":          {"azurerm_public_ip_prefix"},
	"network/routetables":               {"azurerm_route_table"},
	"network/trafficmanagerprofiles":    {"azurerm_traffic_manager_profile"},
	"network/virtualnetworkgateways":    {"azurerm_virtual_network_gateway"},
	"network/virtualnetworks":           {"azurerm_virtual_network"},
	"network/virtualnetworks/subnets":   {"azurerm_subnet"},
	"network/virtualnetworks/virtualnetworkpeerings": {"azurerm_virtual_network_peering"},
	"network/virtualwans":               {"azurerm_virtual_wan"},
	"network/vpngateways":               {"azurerm_vpn_gateway"},
	"network/vpngateways/vpnconnections": {"azurerm_vpn_gateway_connection"},
	"network/vpnsites":                  {"azurerm_vpn_site"},

	// NotificationHubs
	"notificationhubs/namespaces":                  {"azurerm_notification_hub_namespace"},
	"notificationhubs/namespaces/notificationhubs": {"azurerm_notification_hub"},

	// OperationalInsights
	"operationalinsights/workspaces": {"azurerm_log_analytics_workspace"},
	"operationalinsights/clusters":   {"azurerm_log_analytics_cluster"},

	// Portal
	"portal/dashboards": {"azurerm_portal_dashboard"},

	// Purview
	"purview/accounts": {"azurerm_purview_account"},

	// RecoveryServices
	"recoveryservices/vaults": {"azurerm_recovery_services_vault"},

	// Relay
	"relay/namespaces": {"azurerm_relay_namespace"},

	// Resources
	"resources/resourcegroups":   {"azurerm_resource_group"},
	"resources/deployments":      {"azurerm_resource_group_template_deployment"},
	"resources/templatespecs":    {"azurerm_template_deployment"},

	// Search
	"search/searchservices": {"azurerm_search_service"},

	// ServiceBus
	"servicebus/namespaces":                         {"azurerm_servicebus_namespace"},
	"servicebus/namespaces/queues":                  {"azurerm_servicebus_queue"},
	"servicebus/namespaces/topics":                  {"azurerm_servicebus_topic"},
	"servicebus/namespaces/topics/subscriptions":    {"azurerm_servicebus_subscription"},

	// ServiceFabric
	"servicefabric/clusters": {"azurerm_service_fabric_cluster"},

	// SignalRService
	"signalrservice/signalr": {"azurerm_signalr_service"},

	// Sql
	"sql/managedinstances":      {"azurerm_mssql_managed_instance"},
	"sql/servers":               {"azurerm_mssql_server"},
	"sql/servers|azure sql database server": {"azurerm_mssql_server"},
	"sql/servers|azure sql data warehouse":  {"azurerm_mssql_server"},
	"sql/servers/databases":     {"azurerm_mssql_database"},
	"sql/servers/elasticpools":  {"azurerm_mssql_elasticpool"},

	// Storage
	"storage/storageaccounts":                        {"azurerm_storage_account"},
	"storage/storageaccounts/blobservices/containers": {"azurerm_storage_container"},
	"storage/storageaccounts/fileservices/shares":     {"azurerm_storage_share"},

	// StorageSync
	"storagesync/storagesyncservices": {"azurerm_storage_sync"},

	// StreamAnalytics
	"streamanalytics/streamingjobs": {"azurerm_stream_analytics_job"},

	// Synapse
	"synapse/workspaces": {"azurerm_synapse_workspace"},
	"synapse/workspaces/sqlpools|azure synapse analytics spark pool":       {"azurerm_synapse_spark_pool"},
	"synapse/workspaces/sqlpools|azure synapse analytics sql dedicated pool": {"azurerm_synapse_sql_pool"},

	// Web
	"web/serverfarms": {"azurerm_service_plan"},
	"web/sites": {
		"azurerm_linux_web_app",
		"azurerm_windows_web_app",
		"azurerm_linux_function_app",
		"azurerm_windows_function_app",
	},
	"web/sites|web app":            {"azurerm_linux_web_app", "azurerm_windows_web_app"},
	"web/sites|function app":       {"azurerm_linux_function_app", "azurerm_windows_function_app"},
	"web/sites|static web app":     {"azurerm_static_site"},
	"web/sites|azure static web apps": {"azurerm_static_site"},
	"web/sites|app service environment": {"azurerm_app_service_environment_v3"},
	"web/sites/slots":              {"azurerm_linux_web_app_slot", "azurerm_windows_web_app_slot"},
}

// getAzurermResourceType returns the azurerm Terraform provider resource type name(s)
// for a given AzureNamingTool resource path and optional property value.
// Returns a comma-separated string of matching resource types, or an empty string if unknown.
func getAzurermResourceType(resource, property string) string {
	key := strings.ToLower(resource)

	// Try property-specific match first.
	if property != "" {
		propKey := key + "|" + strings.ToLower(property)
		if matches, ok := azurermResourceTypeMap[propKey]; ok {
			return strings.Join(matches, ",")
		}
	}

	// Fall back to resource-only match.
	if matches, ok := azurermResourceTypeMap[key]; ok {
		return strings.Join(matches, ",")
	}

	return ""
}
