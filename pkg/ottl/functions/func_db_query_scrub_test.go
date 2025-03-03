package functions

import (
	"context"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/stretchr/testify/assert"
)

func Test_SqlQueryScrub(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Scrub single-quoted string and number",
			input:    `SELECT * FROM users WHERE name = "John Doe" AND age = 30;`,
			expected: `SELECT * FROM users WHERE name = ? AND age = ?;`,
		},
		{
			name:     "Scrub numbers and single-quoted strings in INSERT",
			input:    `INSERT INTO products (id, name, price) VALUES (42, 'Laptop', 999.99);`,
			expected: `INSERT INTO products (id, name, price) VALUES (?, ?, ?);`,
		},
		{
			name:     "Scrub timestamps and numbers in UPDATE",
			input:    `UPDATE logs SET message = 'Error occurred at 12:34:56' WHERE id = 1234;`,
			expected: `UPDATE logs SET message = ? WHERE id = ?;`,
		},
		{
			name:     "Scrub alphanumeric order_id and numeric customer_id",
			input:    `DELETE FROM orders WHERE order_id = 'ORD-5678' AND customer_id = 789;`,
			expected: `DELETE FROM orders WHERE order_id = ? AND customer_id = ?;`,
		},
		{
			name:     "Scrub numeric table name ",
			input:    `ALTER TABLE orders_23412312234234234 ADD COLUMN foo VARCHAR;`,
			expected: `ALTER TABLE orders_ ADD COLUMN foo VARCHAR;`,
		},
		{
			name:     "Scrub uuid table name",
			input:    `ALTER TABLE tbl_5960ff07_578a_4e49_a543_db92e8432860 ADD COLUMN foo VARCHAR;`,
			expected: `ALTER TABLE tbl_ ADD COLUMN foo VARCHAR;`,
		},
		{
			name:     "Scrub complex query with multiple replacements",
			input:    `SELECT c.customer_id, c.name, o.order_id, o.total_amount FROM customers c JOIN orders o ON c.customer_id = o.customer_id WHERE c.region = 'US' AND o.order_date >= '2024-01-01' AND o.total_amount > 500;`,
			expected: `SELECT c.customer_id, c.name, o.order_id, o.total_amount FROM customers c JOIN orders o ON c.customer_id = o.customer_id WHERE c.region = ? AND o.order_date >= ? AND o.total_amount > ?;`,
		},
		{
			name:     "Scrub subquery with multiple conditions",
			input:    `SELECT p.product_name, p.price, (SELECT COUNT(*) FROM reviews r WHERE r.product_id = p.product_id AND r.rating >= 4) AS positive_reviews FROM products p WHERE p.category = 'Electronics' AND p.stock > 10 AND p.price BETWEEN 100 AND 1000;`,
			expected: `SELECT p.product_name, p.price, (SELECT COUNT(*) FROM reviews r WHERE r.product_id = p.product_id AND r.rating >= ?) AS positive_reviews FROM products p WHERE p.category = ? AND p.stock > ? AND p.price BETWEEN ? AND ?;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprFunc := dbQueryScrub[any](&ottl.StandardStringGetter[any]{
				Getter: func(context.Context, any) (any, error) {
					return tt.input, nil
				},
			})
			result, err := exprFunc(context.Background(), nil)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
