# Dstream Global Configuration
dstream {
  output {
    queue {
      type              = "azure_service_bus"
      name              = "dstream-ingest"
      connection_string = "{{ env "DSTREAM_INGEST_CONNECTION_STRING" }}"
    }

    locks {
      type              = "azure_blob"
      connection_string = "{{ env "DSTREAM_LOCK_CONNECTION_STRING" }}"
      container_name    = "locks"
    }

    poll_interval_defaults {
      poll_interval     = "10s"
      max_poll_interval = "300s"
    }
  }

  required_plugins {
    "ingester-mssql" = {
      source  = "ghcr.io/katasec/dstream-ingester-mssql"
      version = "v1.0.0"
    }

    "ingester-postgres" = {
      source  = "ghcr.io/katasec/dstream-ingester-postgres"
      version = "v1.0.0"
    }
  }
}

# Dstream Task Configuration
task "ingester" "ingest_time" {
  type        = "ingester"
  plugin_path = "./out/dstream-ingester-time"

  config {
    interval = "5s" # optional, default can be hardcoded in plugin
  }
}
