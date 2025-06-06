package types

// Конфигурациии чатов для бота
type Thread struct {
	Network     	string	`yaml:"network"`
	Description 	string	`yaml:"description"`
	ChatID      	int64  	`yaml:"chat_id"`
	ThreadID    	int64  	`yaml:"thread_id"`
}

// массив конфигураций
type Config struct {
	Threads 	[]Thread	`yaml:"threads"`
}
