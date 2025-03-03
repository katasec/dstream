ingester {
    # Database provider type
    db_type = "sqlserver"

    # Connection string for the database
    db_connection_string = "{{ env "DSTREAM_DB_CONNECTION_STRING" }}"

    # Default polling intervals for tables
    poll_interval_defaults {
        poll_interval = "1s"
        max_poll_interval = "5s"
    }

    queue {
        type = "azure_service_bus"  # Type of queue (azure_service_bus, eventhub)
        name = "ingest-queue"
        connection_string = "{{ env "DSTREAM_INGEST_CONNECTION_STRING" }}"
    }

    # Lock configuration
    locks {
        type = "azure_blob"  # Specifies the lock provider type
        connection_string = "{{ env "DSTREAM_BLOB_CONNECTION_STRING" }}"  # Connection string to Azure Blob Storage
        container_name = "locks"  # The name of the container used for lock files
    }

    # List of tables
    tables = [
        "Persons"
    ]

    # Table-specific overrides
    tables_overrides {
        overrides {
            table_name = "Persons"
            poll_interval = "10s"
            max_poll_interval = "300s"
        }
    }
}

publisher {
    source {
        type = "azure_service_bus"
        connection_string = "{{ env "DSTREAM_PUBLISHER_CONNECTION_STRING" }}"  # Used if type is "eventhub" or "servicebus"            
    }

    # Output configuration
    output {
        type = "azure_service_bus"  #  "azure_service_bus"
        connection_string = "{{ env "DSTREAM_PUBLISHER_CONNECTION_STRING" }}"  # Used if type is "eventhub" or "servicebus"
    }
}
