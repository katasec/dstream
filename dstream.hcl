# Example of a plugin-based database connector task
# This would be implemented in a separate plugin repository
# task "database-connector" {
#   plugin_ref = "ghcr.io/katasec/dstream-database-connector:v1.0.0"
#   config {
#     connection_string = "{{ env "DATABASE_CONNECTION_STRING" }}"
#     tables = ["Table1", "Table2"]
#   }
#   
#   input {
#     provider = "database"
#     config {
#       polling_interval = "5s"
#     }
#   }
#   
#   output {
#     provider = "messaging"
#     config {
#       type = "azure_service_bus"
#       connection_string = "{{ env "MESSAGING_CONNECTION_STRING" }}"
#     }
#   }
# }

task "dotnet-counter" {
  type = "plugin"
  plugin_path = "../dstream-dotnet-test/out/dstream-dotnet-test"
  
  // Global configuration for the plugin
  config {
    interval = 5000  // Interval in milliseconds between counter increments
  }
  
  // Input configuration
  input {
    provider = "null"  // Null input provider as this plugin generates its own data
    config {}
  }
  
  // Output configuration
  output {
    provider = "console"  // Console output provider to display counter values
    config {
      format = "json"  // Output format (json or text)
    }
  }
}
