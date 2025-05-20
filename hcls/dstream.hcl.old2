task "ingester" "ingest_mssql" {
  plugin_path = "./dstream-ingester-mssql"

  config {
    db_connection_string = "{{ env "DSTREAM_DB_CONNECTION_STRING" }}"

    poll_interval_defaults {
      poll_interval     = "10s"
      max_poll_interval = "300s"
    }

    queue {
      type              = "azure_service_bus"
      name              = "ingest-queue"
      connection_string = "{{ env "DSTREAM_INGEST_CONNECTION_STRING" }}"
    }

    locks {
      type              = "azure_blob"
      connection_string = "{{ env "DSTREAM_BLOB_CONNECTION_STRING" }}"
      container_name    = "locks"
    }

    tables = [
      "Persons",
      "Cars",
      "Hello",
      "Doesnotexist"
    ]

    tables_overrides {
      overrides {
        table_name        = "Persons"
        poll_interval     = "10s"
        max_poll_interval = "300s"
      }
    }
  }
}

task "router" "route_to_publisher" {
  plugin_path = "./dstream-router"

  config {
    source {
      type              = "azure_service_bus"
      connection_string = "{{ env "DSTREAM_PUBLISHER_CONNECTION_STRING" }}"
    }

    output {
      type              = "azure_service_bus"
      connection_string = "{{ env "DSTREAM_PUBLISHER_CONNECTION_STRING" }}"
    }
  }
}

task "ingester" "ingest_time" {
  type        = "ingester"
  plugin_path = "./out/dstream-ingester-time"

  config {
    interval = "5s" # optional, default can be hardcoded in plugin
  }
}
