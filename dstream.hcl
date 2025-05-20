dstream {
  ingest {
    provider = "azure"

    ingest_queue {
      type              = "azure_service_bus"
      name              = "dstream-ingest"
      connection_string = "{{ env "DSTREAM_INGEST_CONNECTION_STRING" }}"
    }

    lock {
      type              = "azure_blob"
      connection_string = "{{ env "DSTREAM_LOCK_CONNECTION_STRING" }}"
      container_name    = "locks"
    }

    polling {
      interval     = "10s"
      max_interval = "300s"
    }
  }

  plugin_registry = "ghcr.io/katasec"

  required_plugins {
    name    = "ingester-time"
    version = "v0.0.1"
  }

  required_plugins {
    name    = "router"
    version = "0.0.1"
  }
}

task "ingester-mssql" "mssql_orders" {
  tables = ["Orders", "Customers"]

  config {
    db_connection_string = "{{ env "MSSQL_CONN" }}"
  }
}

task "ingester" "ingest_time" {
  type       = "ingester"
  plugin_ref = "ghcr.io/katasec/dstream-ingester-time:v0.0.1"

  config {
    interval = "5s"
  }
}
