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
    provider_ref = "ghcr.io/writeameer/dstream-console-output-provider:v0.3.0"
    config {
      outputFormat = "simple"
    }
  }
}



# OCI Container Registry Example (Production)
# Uses semantic versioned providers distributed as OCI artifacts
task "oci-counter-demo" {
  type = "providers"
  
  input {
    provider_ref = "ghcr.io/writeameer/dstream-counter-input-provider:v0.3.0"
    config {
      interval = 1000    # Generate counter every 1 second
      maxCount = 5       # Stop after 5 iterations
    }
  }
  
  output {
    provider_ref = "ghcr.io/writeameer/dstream-console-output-provider:v0.3.0"
    config {
      outputFormat = "simple"  # Use simple output format
    }
  }
}

# Usage:
#   go run . run mssql-test --log-level debug