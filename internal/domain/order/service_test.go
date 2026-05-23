package order_test

import (
	"testing"

	"github.com/grizlaz/ya-gophermart/internal/domain/order"
	"github.com/stretchr/testify/assert"
)

func TestService_CheckLuhn(t *testing.T) {
	t.Run("test empty number", func(t *testing.T) {
		number := ""
		result := order.CheckLuhn(number)
		assert.False(t, result)
	})
	t.Run("test wrong number", func(t *testing.T) {
		number := "1"
		result := order.CheckLuhn(number)
		assert.False(t, result)
	})
	t.Run("test correct number", func(t *testing.T) {
		number := "12345678903"
		result := order.CheckLuhn(number)
		assert.True(t, result)
	})
	t.Run("test correct long number", func(t *testing.T) {
		number := "1234567891234567891234567891234567891234567891234567891234567891234567891234567891234567891234567891234567891234567891234567891234567891234567891234567891234567891234567891234567891234567897"
		result := order.CheckLuhn(number)
		assert.True(t, result)
	})
	t.Run("test wrong long number", func(t *testing.T) {
		number := "123456789123456789123456789123456789123456789123456789123456789123456789123456789123456789123456789123456789123456789123456789123456789123456789123456789123456789123456789123456789123456789"
		result := order.CheckLuhn(number)
		assert.False(t, result)
	})

}
