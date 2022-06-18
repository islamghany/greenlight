package marshing

import (
	"encoding/json"
)

func MarshalBinary(data map[string]interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func UnmarshalBinary(data []byte, dest interface{}) error {
	if err := json.Unmarshal(data, dest); err != nil {
		return err
	}
	return nil
}
