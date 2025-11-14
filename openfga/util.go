package openfga

import "strings"

func splitObject(obj string) (objectType, objectID string) {
	parts := strings.SplitN(obj, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return obj, ""
}
