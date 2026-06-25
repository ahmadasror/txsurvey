// Package repository holds the pgx-backed data access layer (raw SQL, no ORM).
package repository

import "strings"

// prefixCols qualifies a comma-separated column list with a table alias, e.g.
// prefixCols("f", "id, title") -> "f.id, f.title". Used when a column list is
// reused both standalone and inside a JOIN.
func prefixCols(alias, cols string) string {
	parts := strings.Split(cols, ",")
	for i, p := range parts {
		parts[i] = alias + "." + strings.TrimSpace(p)
	}
	return strings.Join(parts, ", ")
}
