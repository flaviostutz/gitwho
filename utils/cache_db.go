package utils

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

type CacheDB struct {
	db         *sql.DB
	ttlSeconds int
	tableName  string
}

func NewCacheDB(dbFile string, tableName string, ttlSeconds int) (*CacheDB, error) {
	if dbFile == "" {
		return nil, fmt.Errorf("cacheFile was not defined")
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		"CACHE_KEY" TEXT NOT NULL PRIMARY KEY,
		"CACHE_VALUE" TEXT NOT NULL,
		"LAST_ACCESS" TIMESTAMP
		);`, tableName)

	_, err = db.Exec(sql)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Cleaning up old cache entries")
	sql = fmt.Sprintf(`DELETE FROM %s WHERE LAST_ACCESS <= DATETIME(CURRENT_TIMESTAMP, '-%d second');`, tableName, ttlSeconds)
	_, err = db.Exec(sql)
	if err != nil {
		return nil, err
	}

	return &CacheDB{
			db:         db,
			ttlSeconds: ttlSeconds,
			tableName:  tableName},
		nil
}

func (c *CacheDB) PutValue(cacheKey string, cacheContents string) error {
	logrus.Debugf("Saving cache contents")

	sql := fmt.Sprintf(`DELETE FROM %s WHERE CACHE_KEY = ?;`, c.tableName)
	statement, err := c.db.Prepare(sql)
	if err != nil {
		return err
	}
	_, err = statement.Exec(cacheKey)
	if err != nil {
		return err
	}

	sql = fmt.Sprintf(`INSERT INTO %s (CACHE_KEY, CACHE_VALUE, LAST_ACCESS) VALUES (?, ?, CURRENT_TIMESTAMP);`, c.tableName)
	statement, err = c.db.Prepare(sql)
	if err != nil {
		return err
	}
	_, err = statement.Exec(cacheKey, cacheContents)
	if err != nil {
		return err
	}
	return nil
}

func (c *CacheDB) GetValue(cacheKey string) (*string, error) {
	logrus.Debugf("Getting cache contents")

	sql := fmt.Sprintf(`SELECT CACHE_VALUE FROM %s WHERE CACHE_KEY = ? AND LAST_ACCESS > DATETIME(CURRENT_TIMESTAMP, '-%d second');`, c.tableName, c.ttlSeconds)
	statement, err := c.db.Prepare(sql)
	if err != nil {
		return nil, err
	}
	rows, err := statement.Query(cacheKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result *string
	for rows.Next() {
		resultStr := ""
		err = rows.Scan(&resultStr)
		if err != nil {
			return nil, err
		}
		result = &resultStr
	}

	if result != nil {
		// mark last accessed time
		sql := fmt.Sprintf(`UPDATE %s SET LAST_ACCESS = CURRENT_TIMESTAMP WHERE CACHE_KEY = ?`, c.tableName)
		stmt, err := c.db.Prepare(sql)
		if err != nil {
			return nil, err
		}
		_, err = stmt.Exec(cacheKey)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (c *CacheDB) Close() {
	c.db.Close()
}
