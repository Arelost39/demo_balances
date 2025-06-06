package logger

import (
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

// InitLogger настраивает logrus для вывода в консоль и в разные файлы с ежедневной ротацией.
// Логи уровней Info, Debug и Warn будут писаться в один файл (info.log),
// а логи уровней Error, Fatal и Panic – в другой (error.log).
func InitLogger() {

	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	Log.SetLevel(logrus.DebugLevel)

	Log.SetOutput(os.Stdout)

	// Создаем ротацию для логов информационных уровней
	infoWriter, err := rotatelogs.New(
		"./logs/info.log.%Y-%m-%d",
		rotatelogs.WithLinkName("./logs/info.log"),      // всегда ссылка на последний лог
		rotatelogs.WithMaxAge(7*24*time.Hour),            // хранить логи 7 дней
		rotatelogs.WithRotationTime(24*time.Hour),        // ежедневная ротация
	)
	if err != nil {
		Log.Errorf("Ошибка инициализации info лог-файла: %v", err)
	}

	// Создаем ротацию для логов уровней ошибок
	errorWriter, err := rotatelogs.New(
		"./logs/error.log.%Y-%m-%d",
		rotatelogs.WithLinkName("./logs/error.log"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		Log.Errorf("Ошибка инициализации error лог-файла: %v", err)
	}

	// Настраиваем lfshook, чтобы разные уровни логов писались в разные файлы.
	pathMap := lfshook.WriterMap{
		logrus.DebugLevel: infoWriter,
		logrus.InfoLevel:  infoWriter,
		logrus.WarnLevel:  infoWriter,
		logrus.ErrorLevel: errorWriter,
		logrus.FatalLevel: errorWriter,
		logrus.PanicLevel: errorWriter,
	}

	// Добавляем hook для записи логов по уровню с использованием lfshook
	Log.AddHook(lfshook.NewHook(pathMap, &logrus.TextFormatter{
		FullTimestamp: true,
	}))
}
