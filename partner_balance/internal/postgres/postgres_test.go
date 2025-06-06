package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitDB(t *testing.T) {
	// Устанавливаем переменные окружения для теста
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_PORT", "5432")
	os.Setenv("PG_USER", "user")
	os.Setenv("PG_PASSWORD", "password")
	os.Setenv("PG_DBNAME", "blocker")

	// Инициализируем подключение
	err := InitDB()
	assert.NoError(t, err, "Ошибка подключения к БД")

	// Проверяем ping
	err = DB.Ping()
	assert.NoError(t, err, "Ping не удался")

	// Проверочный простой запрос
	var one int
	err = DB.QueryRow("SELECT 1").Scan(&one)
	assert.NoError(t, err)
	assert.Equal(t, 1, one)
}

func TestInsert(t *testing.T) {

	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_PORT", "5432")
	os.Setenv("PG_USER", "user")
	os.Setenv("PG_PASSWORD", "password")
	os.Setenv("PG_DBNAME", "blocker")

	err := InitDB()
	assert.NoError(t, err, "Ошибка подключения к БД")

	err = InsertBalance("UebanAds", 1200, "admeking")
	assert.NoError(t, err, "Данные записаны")
}

func TestGetBalances(t *testing.T) {

	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_PORT", "5432")
	os.Setenv("PG_USER", "user")
	os.Setenv("PG_PASSWORD", "password")
	os.Setenv("PG_DBNAME", "blocker")

	err := InitDB()
	assert.NoError(t, err, "Ошибка подключения к БД")

	data, err := GetBalances("OctoTest", "admeking")
	assert.NoError(t, err)

	testData := []float64{1000, 1200, 1400}
	assert.Equal(t, testData, data)
}
