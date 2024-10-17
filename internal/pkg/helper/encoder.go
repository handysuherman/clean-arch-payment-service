package helper

import (
	"database/sql"
	"fmt"
)

func EncodePgxUUID(id [16]byte) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16])
}

func SqlNString(a string) sql.NullString {
	return sql.NullString{
		String: a,
		Valid:  true,
	}
}
