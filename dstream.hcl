# Database provider type
db_type = "sqlserver"

# Connection string for the database
db_connection_string = "{{ env "DSTREAM_DB_CONNECTION_STRING"  }}"

# Event Hub Connection String
azure_event_hub_connection_string = "{{ env "DSTREAM_EH_CONNECTION_STRING"  }}"
azure_event_hub_name = "{{ env "DSTREAM_EH_NAME"  }}"
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



tables {
    name = "Dano"
    poll_interval = "5s"
    max_poll_interval = "2m"
}