task "ingester-mssql" {
  plugin_ref = "ghcr.io/katasec/dstream-ingester-mssql:v0.0.24"
  config {
    db_connection_string = "blah blah"
    tables = ["Orders", "Customers"]

    ingest_queue {
      provider = "azure"
      type              = "azure_service_bus"
      name              = "dstream-ingest"
      connection_string = "xx"
    }

    lock {
      provider = "azure"
      type              = "azure_blob"
      connection_string = "xx"
      container_name    = "locks"
    }

    polling {
      interval     = "10s"
      max_interval = "300s"
    }    
  }
}

task "ingest-time" {
  type       = "ingester"
  plugin_ref = "ghcr.io/katasec/dstream-ingester-time:v0.0.1"

  config {
    interval = "5s"
  }
}




