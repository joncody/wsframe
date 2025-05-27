package wsframe

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type DBConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (wfa *App) prepareTables() {
	const query = `CREATE TABLE IF NOT EXISTS %s (
				       id bigserial PRIMARY KEY,
					   key text UNIQUE NOT NULL,
					   value json
                  )`

	tables := []string{"auth"}
	for _, details := range wfa.Routes {
		if details.Table == "" || strings.HasPrefix(details.Table, "$") {
			continue
		}
		tables = append(tables, details.Table)
	}
	for _, table := range tables {
		if _, err := wfa.Driver.Exec(fmt.Sprintf(query, table)); err != nil {
			log.Fatal(err)
		}
	}
}

func (wfa *App) GetRow(table, key string) map[string]interface{} {
	value := make([]byte, 0)
	ifacevalue := make(map[string]interface{})
	row := wfa.Driver.QueryRow(fmt.Sprintf(`SELECT value FROM %s WHERE key = $1`, table), key)
	if err := row.Scan(&value); err != nil {
		log.Println(err)
		return ifacevalue
	}
	if err := json.Unmarshal(value, &ifacevalue); err != nil {
		log.Println(err)
		return ifacevalue
	}
	return ifacevalue
}

func (wfa *App) GetRows(table string) []map[string]interface{} {
	ifacevalues := make([]map[string]interface{}, 0)
	rows, err := wfa.Driver.Query(fmt.Sprintf(`SELECT value FROM %s`, table))
	if err != nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		value := make([]byte, 0)
		ifacevalue := make(map[string]interface{})
		if err := rows.Scan(&value); err != nil {
			log.Println(err)
			return ifacevalues
		}
		if err := json.Unmarshal(value, &ifacevalue); err != nil {
			log.Println(err)
			return ifacevalues
		}
		ifacevalues = append(ifacevalues, ifacevalue)
	}
	return ifacevalues
}

func (wfa *App) InsertRow(table, key, value string) error {
	if _, err := wfa.Driver.Exec(fmt.Sprintf(`INSERT INTO %s (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`, table), key, value); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
