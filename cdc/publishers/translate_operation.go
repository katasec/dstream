package publishers

// TranslateOperation converts an operation code into a human-readable string.
func TranslateOperation(operation int) string {
	switch operation {
	case 1:
		return "delete"
	case 2:
		return "create"
	case 4:
		return "update"
	default:
		return "unknown"
	}
}
