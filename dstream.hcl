# Database provider type
db_type = "sqlserver"

# Connection string for the database
db_connection_string = "{{ env "DSTREAM_DB_CONNECTION_STRING" }}"

# Output configuration
output {
    type = "servicebus"  # Possible values: "console", "eventhub", "servicebus"
    connection_string = "{{ env "DSTREAM_PUBLISHER_CONNECTION_STRING" }}"  # Used if type is "eventhub" or "servicebus"
}

# Lock configuration
locks {
    type = "azure_blob_db"  # Specifies the lock provider type
    connection_string = "{{ env "DSTREAM_BLOB_CONNECTION_STRING" }}"  # Connection string to Azure Blob Storage
    container_name = "locks"  # The name of the container used for lock files
}

# Table configurations with polling intervals
tables {
    name = "Persons"
    poll_interval = "5s"
    max_poll_interval = "1m"
}

tables {
    name = "Cars"
    poll_interval = "5s"
    max_poll_interval = "2m"
}

tables {
    name = "Dano"
    poll_interval = "5s"
    max_poll_interval = "2m"
}
