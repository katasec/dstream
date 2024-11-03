# Database provider type
db_type = "sqlserver"

# Connection string for the database
db_connection_string = "{{ env "DSTREAM_DB_CONNECTION_STRING"  }}"

# Output provider type
output {
    type = "console"
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
