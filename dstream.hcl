# DStream Configuration Examples
# =============================

# 1. OCI Container Registry Example (Production)
# Uses semantic versioned providers distributed as OCI artifacts
task "oci-counter-demo" {
  type = "providers"
  
  input {
    provider_ref = "ghcr.io/katasec/dstream-counter-input-provider:v0.2.0"
    config {
      interval = 1000    # Generate counter every 1 second
      max_count = 5      # Stop after 5 iterations
    }
  }
  
  output {
    provider_ref = "ghcr.io/katasec/dstream-console-output-provider:v0.2.0"
    config {
      outputFormat = "simple"  # Use simple output format
    }
  }
}

# 2. Local Provider Path Example (Development)
# Uses locally built provider binaries for development and testing
task "local-counter-demo" {
  type = "providers"
  
  input {
    provider_path = "../dstream-providers/counter-input-provider/bin/Release/net9.0/darwin-arm64/publish/counter-input-provider"
    config {
      interval = 2000    # Generate counter every 2 seconds
      max_count = 3      # Stop after 3 iterations
    }
  }
  
  output {
    provider_path = "../dstream-providers/console-output-provider/bin/Release/net9.0/darwin-arm64/publish/console-output-provider"
    config {
      outputFormat = "structured"  # Use structured output format
    }
  }
}

# 3. Legacy Plugin Example (Backward Compatibility)
# Single .NET plugin using gRPC communication
task "legacy-plugin-demo" {
  type = "plugin"
  plugin_path = "../dstream-dotnet-sdk/samples/dstream-dotnet-test/out/dstream-dotnet-test"
   
  config {
    interval = 500  # Plugin-level configuration
  }
  
  input {
    provider = "null"
    config {
      interval = 1000
    }
  }
  
  output {
    provider = "console"
    config {
      format = "json"
    }
  }
}

# 4. Production CDC Pipeline Example (Future)
# Real-world SQL Server CDC to Azure Service Bus
task "production-cdc-pipeline" {
  type = "providers"
  
  input {
    provider_ref = "ghcr.io/katasec/mssql-cdc-provider:v1.2.0"
    config {
      connection_string = "{{ env "DATABASE_CONNECTION_STRING" }}"
      tables = ["Orders", "Customers", "Inventory"]
      polling_interval = 1000
      batch_size = 100
    }
  }
  
  output {
    provider_ref = "ghcr.io/katasec/azure-servicebus-provider:v1.1.0"
    config {
      connection_string = "{{ env "MESSAGING_CONNECTION_STRING" }}"
      queue_name = "data-events"
      message_retention = "7d"
    }
  }
}

# 5. Infrastructure Lifecycle Test (Local Development)
# Demonstrates providers with infrastructure management
task "test-infrastructure" {
  type = "providers"
  
  input {
    provider_path = "../dstream-providers/counter-input-provider/bin/Release/net9.0/darwin-arm64/publish/counter-input-provider"
    config {
      interval = 2000     # Generate every 2 seconds
      max_count = 10      # Generate 10 messages
    }
  }
  
  output {
    provider_path = "../dstream-dotnet-sdk/samples/test-infrastructure-provider/bin/Release/net9.0/osx-arm64/publish/test-infrastructure-provider"
    config {
      testValue = "azure-service-bus"
      resourceCount = 3   # This output provider will manage 3 infrastructure resources
    }
  }
}

# Usage Examples:
#
# Run OCI-based demo (production):
#   ./dstream run oci-counter-demo
#
# Run local development demo:
#   ./dstream run local-counter-demo
#
# Run legacy plugin demo:
#   ./dstream run legacy-plugin-demo
#
# Show help and available commands:
#   ./dstream --help
#
# Run with debug logging:
#   ./dstream run oci-counter-demo --log-level debug
#
# Infrastructure lifecycle commands:
#   ./dstream init test-infrastructure    # Initialize resources
#   ./dstream plan test-infrastructure    # Show planned changes  
#   ./dstream status test-infrastructure  # Show current status
#   ./dstream destroy test-infrastructure # Destroy resources
