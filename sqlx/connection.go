package sqlx

import (
	"fmt"
	"strings"
)

// ConnectionString erzeugt aus einem Dateinamen und Optionen einen DSN
func ConnectionString(filename string, options ...Option) string {
	var opts []string
	for _, o := range options {
		opts = append(opts, o.String())
	}
	return fmt.Sprintf("file:%v?%v", filename, strings.Join(opts, "&"))
}
