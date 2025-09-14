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

# Legacy single plugin task (keep for compatibility)
task "dotnet-counter-plugin" {
  type = "plugin"
  plugin_path = "../dstream-dotnet-sdk/samples/dstream-dotnet-test/out/dstream-dotnet-test"
   
  // Global configuration for the plugin
  config {
    interval = 500  // Interval in milliseconds between counter increments
  }
  
  // Input configuration
  input {
    provider = "null"  // Null input provider as this plugin generates its own data
    config {
      interval = 1000
    }
  }
  
  // Output configuration
  output {
    provider = "console"  // Console output provider to display counter values
    config {
      format = "json"  // Output format (json or text)
    }
  }
}

# New independent provider task
task "counter-to-console" {
  type = "providers"  # New type for independent provider orchestration
  
  input {
    provider_path = "../dstream-counter-input-provider/bin/Release/net9.0/osx-x64/counter-input-provider"
    config {
      interval = 1000   # Generate counter every 1 second
      max_count = 50    # Stop after 50 iterations
    }
  }
  
  output {
    provider_path = "../dstream-console-output-provider/bin/Release/net9.0/osx-x64/console-output-provider"
    config {
      outputFormat = "structured"  # Use structured output format
    }
  }
}
