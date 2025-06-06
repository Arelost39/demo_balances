package processor

import (
	"fmt"
	"math"
	"partner_balance/internal/logger"
	db "partner_balance/internal/postgres"
	"partner_balance/internal/req/partner1"
	"partner_balance/internal/req/partner2"
	"partner_balance/internal/req/partner3"

	"partner_balance/internal/utils"
	"sort"
	"strings"
	"sync"
	"time"
)

type Partner struct {
	Name  string // название партнёра
	Token string // токен из конфига
}

type NetworkGroup struct {
	GroupName string    // например "adrich" или "clickmoney"
	Partners  []Partner // все активные партнёры этой сетки
}

// Вспомогательные функции
func RoundTo(x float64, n int) float64 {
	return math.Round(x*math.Pow(10, float64(n))) / math.Pow(10, float64(n))
}

func AverageArrow(balances []float64) []float64 {
    if len(balances) < 2 {
        return []float64{}
    }
    diffs := make([]float64, 0, len(balances)-1)
    for i := 1; i < len(balances); i++ {
        prev, curr := balances[i-1], balances[i]
        if prev != 0 && curr != 0 && curr > prev {
            diffs = append(diffs, curr-prev)
        }
    }
    return diffs
}

func AverageValue(balances []float64) float64 {
	if len(balances) == 0 {
		return 0.0
	}
	var value float64
	for _, v := range balances {
		value += v*4
	}
	return  RoundTo(value/float64(len(balances)), 2)
}

func FormatMapWithTimestamp(m map[string]string) string {
    now := time.Now().Format("02-01-2006 / 15:04")
    var sb strings.Builder
    sb.WriteString(now)
    sb.WriteByte('\n')

    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    for _, k := range keys {
        sb.WriteString(fmt.Sprintf("<b>%s</b>: %s\n\n", k, m[k]))
    }

    return sb.String()
}

// Загрузка списка сеток, их партнеров и токенов
func PartnerList() []NetworkGroup {
	var groups []NetworkGroup
	for groupName, partners := range utils.AppConfig.Networks {
		ng := NetworkGroup{GroupName: groupName}

		for partnerName, cfg := range partners {
			if !cfg.IsActive {
				continue
			}
			ng.Partners = append(ng.Partners, Partner{
				Name:  partnerName,
				Token: cfg.Token,
			})
		}

		if len(ng.Partners) > 0 {
			groups = append(groups, ng)
		}
	}
	return groups
}

// Вставка или обновление всех партнёров из списка групп в базу данных
func InsertPartners(groups []NetworkGroup) error {
	for _, ng := range groups {
		for _, p := range ng.Partners {
			// Вставляем или обновляем партнёра в базе данных
			if err := db.InsertPartner(p.Name, ng.GroupName, true); err != nil {
				return fmt.Errorf(
					"processor.InsertPartners: не удалось добавить партнёра %s в сеть %s: %v",
					p.Name, ng.GroupName, err,
				)
			}
			logger.Log.Debugf("Партнёр %s успешно добавлен/обновлён в сети %s", p.Name, ng.GroupName)
		}
	}
	logger.Log.Info("Операция InsertPartners завершена")
	return nil
}

func Router(partners Partner) (float64, error) {
	var result float64
	var err error
	switch partners.Name {
	case "Partner1":
		result, err = partner1.GetBalance(partners.Token)
	case "Partner2":
		result, err = partner2.GetBalance(partners.Token)
	case "Partner3":
		result, err = partner3.GetBalance(partners.Token)
	default:
		logger.Log.Errorf("Ошибка: неправильное название партнера (router): %s", partners.Name)
		return 0, fmt.Errorf("неправильное название партнера (router)")
	}
	if err != nil {
		logger.Log.Errorf("Ошибка получения баланса у партнера %s: %v", partners.Name, err)
		return 0, fmt.Errorf("ошибка получения баланса: %v", err)
	}
	return result, nil
}

// Вставка баланса в бд
func BalanceInsert() error {
	groups := PartnerList()
	var wg sync.WaitGroup

	for _, v := range groups {
		groupName := v.GroupName
		for _, partner := range v.Partners {
			wg.Add(1)
			go func(partner Partner, groupName string) error {
				defer wg.Done()
				balance, err := Router(partner)
				if err != nil {
					logger.Log.Errorf("Ошибка получения баланса партнера %s (группа %s): %v", partner.Name, groupName, err)
					return err
				}
				if err := db.InsertBalance(partner.Name, balance, groupName); err != nil {
					logger.Log.Errorf("Ошибка вставки баланса партнера %s (группа %s): %v", partner.Name, groupName, err)
					return err
				}
				logger.Log.Debugf("Успешно вставлен баланс партнера %s (группа %s): %.2f", partner.Name, groupName, balance)
				return nil
			}(partner, groupName)
		}
	}
	wg.Wait()
	logger.Log.Info("Завершена операция вставки балансов")
	return nil
}

// Функция получения среднего спенда из бд
func GetStatistic() map[string]map[string]float64 {
	stats := make(map[string]map[string]float64)
	groups := PartnerList()
	for _, v := range groups {
		if stats[v.GroupName] == nil {
			stats[v.GroupName] = make(map[string]float64)
		}
		for _, x := range v.Partners {
			dbData, err := db.GetBalances(x.Name, v.GroupName)
			if err != nil {
				logger.Log.Errorf("Ошибка GetStatistic %s --- %s: %v", x.Name, v.GroupName, err)
				return nil
			}
			averData := AverageArrow(dbData)
			averValue := AverageValue(averData)
			stats[v.GroupName][x.Name] = averValue
		}
	}
	return stats
}

// Функция формирования строки с текущим балансом
func CallBalanceList(networkName string) string {
	logger.Log.Infof("Вызов CallBalance для сети: %s", networkName)
	balances := make(map[string]string)
	for _, group := range PartnerList() {
		if group.GroupName != networkName {
			continue
		}
		for _, partner := range group.Partners {
			balance, err := Router(partner)
			if err != nil {
				logger.Log.Errorf("Ошибка получения баланса партнера %s: %v", partner.Name, err)
				continue
			}
			balances[partner.Name] = fmt.Sprintf("%.2f", RoundTo(balance, 2))
			logger.Log.Debugf("Получен баланс партнера %s: %.2f", partner.Name, balance)
		}
	}
	result := FormatMapWithTimestamp(balances)
	logger.Log.Infof("Завершена операция CallBalance для сети %s", networkName)
	return result
}

// формирование списка партнеров, которых надо будет пополнить
func CompareBalances(networkName string) string {
	logger.Log.Infof("Вызов CompareBalances для сети: %s", networkName)
	stats := GetStatistic()
	alerts := make(map[string]string)

	for _, group := range PartnerList() {
		if group.GroupName != networkName {
			continue
		}
		for _, partner := range group.Partners {
			avgVal, exists := stats[networkName][partner.Name]
			if !exists {
				logger.Log.Warnf("Не найдена статистика для партнёра %s", partner.Name)
				continue
			}

			threshold := avgVal * 1.3

			bal, err := Router(partner)
			if err != nil {
				logger.Log.Errorf("Ошибка получения баланса партнёра %s: %v", partner.Name, err)
				continue
			}
			if bal < threshold {
				alerts[partner.Name] = fmt.Sprintf("%.2f (spend %.2f) ⚠️", RoundTo(bal, 2), RoundTo(avgVal, 2))
				logger.Log.Warnf("Баланс партнёра %s ниже порога (%.2f < %.2f)", partner.Name, bal, threshold)
			} else {
				alerts[partner.Name] = fmt.Sprintf("%.2f (spend %.2f)", RoundTo(bal, 2), RoundTo(avgVal, 2))
				logger.Log.Infof("Баланс партнёра %s в норме (%.2f < %.2f)", partner.Name, bal, threshold)
			}
		}
	}

	if len(alerts) == 0 {
		logger.Log.Infof("Нет алертов для сети %s, сообщение не будет отправлено", networkName)
		return ""
	}
	result := FormatMapWithTimestamp(alerts)
	logger.Log.Infof("Завершена операция CompareBalances для сети %s", networkName)
	return result
	// ответ будет иметь такой формат
	//–––––––––––––––––––––––––––––––––––––
	//01-01-2025 / 17:16
	//Partner1: 120.03 (spend 11.66)
	//
	//Partner2: 231.39 (spend 10.05)
	// 			.	.	.
	//PartnerN: 22.22 (spend 55.05) ⚠️
	//–––––––––––––––––––––––––––––––––––––
	//
}

func CallBalanceListWithStat(networkName string) string {
	logger.Log.Infof("Вызов CompareBalances для сети: %s", networkName)
	stats := GetStatistic()
	alerts := make(map[string]string)

	for _, group := range PartnerList() {
		if group.GroupName != networkName {
			continue
		}
		for _, partner := range group.Partners {
			avgVal, exists := stats[networkName][partner.Name]
			if !exists {
				logger.Log.Warnf("Не найдена статистика для партнёра %s", partner.Name)
				continue
			}

			threshold := avgVal * 1.3

			bal, err := Router(partner)
			if err != nil {
				logger.Log.Errorf("Ошибка получения баланса партнёра %s: %v", partner.Name, err)
				continue
			}
			if bal < threshold {
				alerts[partner.Name] = fmt.Sprintf("%.2f (spend %.2f) ⚠️", RoundTo(bal, 2), RoundTo(avgVal, 2))
				logger.Log.Warnf("Баланс партнёра %s ниже порога (%.2f < %.2f)", partner.Name, bal, threshold)
			} else {
				alerts[partner.Name] = fmt.Sprintf("%.2f (spend %.2f)", RoundTo(bal, 2), RoundTo(avgVal, 2))
				logger.Log.Infof("Баланс партнёра %s в норме (%.2f < %.2f)", partner.Name, bal, threshold)
			}
		}
	}

	if len(alerts) == 0 {
		logger.Log.Infof("Нет алертов для сети %s, сообщение не будет отправлено", networkName)
		return ""
	}
	result := FormatMapWithTimestamp(alerts)
	logger.Log.Infof("Завершена операция CompareBalances для сети %s", networkName)
	return result
}
