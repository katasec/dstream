task "ingester-mssql" {
  plugin_ref = "ghcr.io/katasec/dstream-ingester-mssql:v0.0.53"
  config {
    db_connection_string = "{{ env "DSTREAM_DB_CONNECTION_STRING" }}"
    tables = ["Cars", "Persons"]

    ingest_queue {
      provider = "azure"
      type              = "azure_service_bus"
      name              = "dstream-ingest"
      connection_string = "{{ env "DSTREAM_INGEST_CONNECTION_STRING" }}"
    }

    lock {
      provider = "azure"
      type              = "azure_blob"
      connection_string = "{{ env "DSTREAM_BLOB_CONNECTION_STRING"}}"
      container_name    = "locks"
    }

    polling {
      interval     = "5s"
      max_interval = "30s"
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

task "dotnet-counter" {
  type = "plugin"
  plugin_path = "../dstream-dotnet-test/out/dstream-dotnet-test"
  
  // Global configuration for the plugin
  config {
    interval = "5000"  // Interval in milliseconds between counter increments
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
