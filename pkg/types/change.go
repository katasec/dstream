package types

// ChangeEvent represents a change in a database table
type ChangeEvent struct {
	TableName string                 `json:"table_name"`
	Operation string                 `json:"operation"` // insert, update, delete
	LSN       string                 `json:"lsn"`
	Data      map[string]interface{} `json:"data"`
}
