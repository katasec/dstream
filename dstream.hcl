# Database provider type
db_type = "sqlserver"

# Connection string for the database
db_connection_string = "{{ env "DSTREAM_DB_CONNECTION_STRING"  }}"

# Output configuration
output {
    type = "console"  # Possible values: "console", "eventhub", "servicebus"
    connection_string = "{{ env "DSTREAM_PUBLISHER_CONNECTION_STRING"  }}"  # Used if type is "eventhub" or "servicebus"
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
