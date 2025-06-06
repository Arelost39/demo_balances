package db

import (
	"database/sql"
	"fmt"
	"os"
	"log"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("PG_HOST"),
		os.Getenv("PG_PORT"),
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_DBNAME"),
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("ошибка ping БД: %v", err)
	}

	return nil
}

func InsertBalance(partnerName string, balance float64, network string) error {
	var partnerID int
	selectQuery := fmt.Sprintf(`SELECT id FROM balance_%s.partners WHERE partner = $1`, network)

	err := DB.QueryRow(selectQuery, partnerName).Scan(&partnerID)
	if err != nil {
		return fmt.Errorf("партнёр %s не найден: %v", partnerName, err)
	}

	insertQuery := fmt.Sprintf(`
        INSERT INTO balance_%s.balances (partner_id, created_at, balance)
        VALUES ($1, CURRENT_TIMESTAMP, $2)
	`, network)

	_, err = DB.Exec(insertQuery, partnerID, balance)

	if err != nil {
		return fmt.Errorf("ошибка вставки баланса: %v", err)
	}

	return nil
}

func GetBalances(partnerName string, network string) ([]float64, error) {
	
	var partnerID int
	selectQuery := fmt.Sprintf(`SELECT id FROM balance_%s.partners WHERE partner = $1`, network)

	err := DB.QueryRow(selectQuery, partnerName).Scan(&partnerID)
	if err != nil {
		return nil, fmt.Errorf("партнёр %s не найден: %v", partnerName, err)
	}

	balanceQuery := fmt.Sprintf(`
		SELECT 
			b.balance
		FROM balance_%s.balances b
		JOIN balance_%s.partners p 
			ON b.partner_id = p.id
		WHERE 
			b.partner_id = $1
			AND b.created_at >= CURRENT_DATE - INTERVAL '3 days'
		ORDER BY 
			b.created_at DESC
	`, network, network)

	rows, err := DB.Query(balanceQuery, partnerID)
	if err != nil {			
		return nil, fmt.Errorf("ошибка запроса балансов: %v", err)
	}

	defer rows.Close()
	
	var balances []float64

	for rows.Next() {
		var balance float64
		err := rows.Scan(&balance)

		if err != nil {
			log.Printf("Ошибка чтения строки: %v", err)
			continue
		}

	balances = append(balances, balance)
	}
	return balances, nil
}

func DeleteOldData(network string) error {
	query := fmt.Sprintf(`
	DELETE FROM balance_%s.balances
	WHERE balance_%s.balances.created_at < CURRENT_DATE - INTERVAL '7 days'
	`, network, network)

	res, err := DB.Exec(query)
	if err != nil {
		log.Printf("Ошибка запроса на удаление: %v", err)
		return err
	}
	log.Printf("Успешно выполнено удаление старых данных для %s: %v", network, res)
	return nil
}

func InsertPartner(partnerName string, network string, isActive bool) error {
    insertQuery := fmt.Sprintf(`
        INSERT INTO balance_%s.partners (partner, is_active)
        VALUES ($1, $2)
        ON CONFLICT (partner) DO UPDATE
          SET is_active = EXCLUDED.is_active
    `, network)

    if _, err := DB.Exec(insertQuery, partnerName, isActive); err != nil {
        return fmt.Errorf("ошибка вставки партнёра '%s' в сеть '%s': %v", partnerName, network, err)
    }
    return nil
}