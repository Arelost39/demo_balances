package utils

import (
	"log"
	"time"
)

var LocalLocation *time.Location

// InitLocation загружает нужную локацию и сохраняет её в глобальной переменной.
func InitLocation() {
	// Используем "Europe/Kaliningrad"
	loc, err := time.LoadLocation("Europe/Kaliningrad")
	if err != nil {
		log.Printf("Ошибка загрузки локации 'Europe/Kaliningrad': %v. Используется UTC.", err)
		LocalLocation = time.UTC
	} else {
		LocalLocation = loc
	}
}

// LocalNow возвращает текущее время в заданном часовом поясе.
func LocalNow() time.Time {
	if LocalLocation == nil {
		InitLocation()
	}
	return time.Now().In(LocalLocation)
}
