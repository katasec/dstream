# MSSQL CDC Test Task
# Test the MSSQL ingester provider with DStream orchestration
task "mssql-test" {
  type = "providers"  
  input {
    provider_ref = "ghcr.io/katasec/dstream-ingester-mssql:v0.0.55"
    config {
      db_connection_string = "server=localhost,1433;user id=sa;password=Passw0rd123;database=TestDB;encrypt=disable"
      poll_interval = "5s"
      max_poll_interval = "30s"
      tables = ["Persons", "Cars"]
      lock_config = {
        type = "none"
      }
    }
  }
  
  output {
    provider_ref = "ghcr.io/katasec/dstream-log-output-provider:v0.1.0"
    config {
      logLevel = "info"
    }
  }
}



# OCI Container Registry Example (Production)
# Uses semantic versioned providers distributed as OCI artifacts
task "oci-counter-demo" {
  type = "providers"
  
  input {
    provider_ref = "ghcr.io/katasec/dstream-counter-input-provider:v0.2.0"
    config {
      interval = 1000    # Generate counter every 1 second
      maxCount = 5       # Stop after 5 iterations
    }
  }
  
  output {
    provider_ref = "ghcr.io/katasec/dstream-console-output-provider:v0.2.0"
    config {
      outputFormat = "simple"  # Use simple output format
    }
  }
}

# MSSQL CDC -> Azure Service Bus (Capability Reset target)
# This is the original production capability rebuilt on the new architecture
task "mssql-to-asb" {
  type = "providers"

  input {
    provider_ref = "ghcr.io/katasec/dstream-ingester-mssql:v0.0.57"
    config {
      db_connection_string = "server=localhost,1433;user id=sa;password=Passw0rd123;database=TestDB;encrypt=disable"
      poll_interval   = "5s"
      max_poll_interval = "30s"
      tables = ["Persons", "Cars"]
      lock_config = {
        type = "none"
      }
    }
  }

  output {
    provider_ref = "ghcr.io/katasec/dstream-out-asb:v0.2.0"
    config {
      connectionString = "{{ env `ASB_CONNECTION_STRING` }}"
      resourceGroup    = "rg-dstream-dev"
      namespace        = "sb-dstream-dev"
      sourceHost       = "localhost"
      sourceDatabase   = "TestDB"
      tables           = ["dbo.Cars", "dbo.Persons"]
    }
  }
}

# Usage:
#   go run . run mssql-test --log-level debug
#   ASB_CONNECTION_STRING="Endpoint=sb://..." go run . init mssql-to-asb
#   ASB_CONNECTION_STRING="Endpoint=sb://..." go run . run mssql-to-asb