package database

import (
	"database/sql"
	"fmt"
	udt "go_final_project/myLib/UDT"
	"os"
	"path/filepath"
)

var (
	CollectPtr udt.PtrLib // Глобальная переменная для хранения указателей
)

// Функция проверяет присутствие БД, если её нет - создаётся. Возвращает указатель на БД и ошибку
func CheckCreateDB() (*sql.DB, error) {

	// Определение расположения исполняемого файла
	// Проверка присутствия файла БД
	appPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("directory definition error: %w", err)
	}

	dbFile := filepath.Join(filepath.Dir(appPath), os.Getenv("DB_NAME"))
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		install = true // Установка признака в необходимости создания БД и её инициализации
	}

	// Подключение к БД
	// если БД нет, она создаётся и пингуется
	db, err := sql.Open(os.Getenv("DB_DRIVER"), os.Getenv("DB_NAME"))
	if err != nil {
		return nil, fmt.Errorf("error in open db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error in db ping: %w", err)
	}

	if install {

		str := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
 			"date" VARCHAR(250),
 			"title" VARCHAR(250) NOT NULL DEFAULT '',
 			"comment" VARCHAR(250) NOT NULL DEFAULT '',
 			"repeat" VARCHAR(128)
			);
    	`, os.Getenv("DB_TABLE_NAME"))

		stmt, err := db.Prepare(str)
		if err != nil {
			return nil, fmt.Errorf("error in create table prepare : %w", err)
		}
		defer func() { _ = stmt.Close() }()

		_, err = stmt.Exec()
		if err != nil {
			return nil, fmt.Errorf("error in create table execution : %w", err)
		}

		CollectPtr.PointerLogI.Println("the table has been created") // логирование события

		str = fmt.Sprintf("CREATE INDEX %s_date ON %s (date)", os.Getenv("DB_TABLE_NAME"), os.Getenv("DB_TABLE_NAME"))

		_, err = db.Exec(str)
		if err != nil {
			return nil, fmt.Errorf("error in create index execution : %w", err)
		}

		CollectPtr.PointerLogI.Println("the index has been created") // логирование события
	}

	return db, nil
}

// Функция выполняет запись строки в БД. Возвращает индекс добавленной записи и ошибку.
//
// Параметры:
//
// dbptr - указатель на БД
// data - данные подлежащие записи в БД
func WrRow(dbPtr *sql.DB, data udt.RxFormat) (int64, error) {

	res, err := dbPtr.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", data.Date),
		sql.Named("title", data.Title),
		sql.Named("comment", data.Comment),
		sql.Named("repeat", data.Repeat))

	if err != nil {
		return 0, fmt.Errorf("request - Error : %w", err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last index - Error : %w", err)
	}

	return lastID, nil
}

// Функция запрашивает у БД все строки по >= переданной даты. Возвращает строки и ошибку.
//
// Параметры:
//
// dbptr - указатель на БД
// date - дата
func RdRows(dbPtr *sql.DB, date string) ([]udt.TxFormatEl, error) {

	rows, err := dbPtr.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE date >= :date ORDER BY date LIMIT 15 ",
		sql.Named("date", date))

	if err != nil {
		return nil, fmt.Errorf("request error by id: %v, err: %v", date, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	// Выворд содержмимого запроса
	//
	t := make([]udt.TxFormatEl, 0)

	for rows.Next() {

		task := udt.TxFormatEl{}

		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

		if err != nil {
			return t, fmt.Errorf("scanning rows error: %v", err)
		}

		t = append(t, task)
	}

	if rows.Err() != nil {
		return t, fmt.Errorf("cursor error: %v", err)
	}

	return t, nil

}

// Функция запрашивает у БД строку по её id. Возвращает строки и ошибку.
//
// Параметры:
//
// dbptr - указатель на БД
// id - индекс
func RdRow(dbPtr *sql.DB, id int) (udt.TxFormatEl, error) {

	task := udt.TxFormatEl{}

	row := dbPtr.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id ",
		sql.Named("id", id))

	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return task, fmt.Errorf("scanning row error: %v", err)
	}

	return task, nil
}

// Функция обновляет сохранённые в БД данные. Возвращает ошибку.
//
// Параметры:
//
// dbptr - указатель на БД
// data - данные
func UpdRow(dbPtr *sql.DB, data udt.RxFormatFull) error {

	_, err := dbPtr.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
		sql.Named("date", data.Date),
		sql.Named("title", data.Title),
		sql.Named("comment", data.Comment),
		sql.Named("repeat", data.Repeat),
		sql.Named("id", data.ID))

	if err != nil {
		return fmt.Errorf("scanning row error: %v", err)
	}

	return nil
}

// Функция удаляет строку из БД по её id. Возвращает ошибку.
//
// Параметры:
//
// dbptr - указатель на БД
// id - индекс
func DlRow(dbPtr *sql.DB, id int) error {

	_, err := dbPtr.Exec("DELETE FROM scheduler WHERE id = :id",
		sql.Named("id", id))

	if err != nil {
		return fmt.Errorf("delete row error: %v", err)
	}

	return nil
}
