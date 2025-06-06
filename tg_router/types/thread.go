package types

// Thread представляет тред в Telegram
type Thread struct {
	Network     	string	`yaml:"network"`
	Description 	string	`yaml:"description"`
	ChatID      	int64  	`yaml:"chat_id"`
	ThreadID    	int64  	`yaml:"thread_id"`
}

// Config содержит конфигурацию тредов
type Config struct {
	Threads 	[]Thread	`yaml:"threads"`
}
