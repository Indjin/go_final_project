package httph

import (
	"encoding/json"
	"errors"
	"fmt"
	udt "go_final_project/myLib/UDT"
	dbA "go_final_project/myLib/dataBase"
	genDate "go_final_project/myLib/dateGen"
	"io"
	"net/http"
	"strconv"
	"time"
)

var (
	CollectPtr udt.PtrLib // Глобальная переменная для хранения указателей
)

// Обработчик правил повторения
func NextDateH(w http.ResponseWriter, r *http.Request) {

	CollectPtr.PointerLogI.Println("get request accepted to generation date")

	// Чтение значений Query параметров
	//
	nowV := r.URL.Query().Get("now")
	dateV := r.URL.Query().Get("date")
	repeatV := r.URL.Query().Get("repeat")

	// Проверка корректности содержимого nowV
	//
	nowD, err := time.Parse("20060102", nowV)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("can't parse the value %v to 20060102 format\n", nowV)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Генерация даты по данным из запроса
	//
	repStr, err := genDate.NextDate(nowD, dateV, repeatV)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("can't do generation the next date by now = %v, date = %v, repeat = %v\n", nowV, dateV, repeatV)
		http.Error(w, `"Bad request"`, http.StatusBadRequest)
		return
	}

	CollectPtr.PointerLogI.Println("request - Ok")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(repStr))

}

// Обработчик добавления задач в БД
func AddTasksH(w http.ResponseWriter, r *http.Request) {

	CollectPtr.PointerLogI.Println("post request accepted to add task")

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Подготовка аварийного сообщения в случае ошибки обработчика
	//
	/*
		dataSendError, err := generationErrMsg("Не указан заголовок задачи")
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Println("can't do generation error message")
			http.Error(w, `"Не указан заголовок задачи"`, http.StatusInternalServerError)
			return
		}
	*/

	// Чтение тела
	//
	bodyReq, err := io.ReadAll(r.Body)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Println("can't do read the body of request")
		http.Error(w, `"Bad Request"`, http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	var msgRx udt.RxFormat

	err = json.Unmarshal(bodyReq, &msgRx)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("unmarshal function have error: %v", err)
		//w.Write(dataSendError)
		err = sendErrMsg("Не указан заголовок задачи", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	// Проверка содержимого поля title на отсутствие данных
	//
	if msgRx.Title == "" {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Println("the field title is empty")
		//w.Write(dataSendError)
		err = sendErrMsg("Не указан заголовок задачи", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	// Проверка принятых данных в поле date
	//
	tn := time.Now()

	if msgRx.Date == "" {
		msgRx.Date = tn.Format("20060102")
	} else { // Проверка корректности данных в поле date
		_, err = time.Parse("20060102", msgRx.Date)
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("can't do parese %v, by 20060102 err: %v\n", msgRx.Date, err)
			//w.Write(dataSendError)
			err = sendErrMsg("Не указан заголовок задачи", w)
			if err != nil {
				CollectPtr.PointerLogI.Println("error send message")
				CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
			}
			return
		}
	}

	// Проверка принятых данных в поле repeat
	// и вычисление новой даты
	//
	if msgRx.Repeat == "" {
		msgRx.Date = tn.Format("20060102")
	} else { // При указанном правиле повторения
		msgRx.Date, err = genDate.NextDate(tn, msgRx.Date, msgRx.Repeat)
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("can't do generation the new date by date = %v, repeat = %v\n", msgRx.Date, msgRx.Repeat)
			//w.Write(dataSendError)
			err = sendErrMsg("Не указан заголовок задачи", w)
			if err != nil {
				CollectPtr.PointerLogI.Println("error send message")
				CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
			}
			return
		}
	}

	// Добавление записи в БД
	//
	/*
		sqlDB := CollectPtr.PointerDB

		res, err := sqlDB.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
			sql.Named("date", msgRx.Date),
			sql.Named("title", msgRx.Title),
			sql.Named("comment", msgRx.Comment),
			sql.Named("repeat", msgRx.Repeat))

		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("error insert in table new data: %v\n", msgRx)
			http.Error(w, `"Internal Server Error"`, http.StatusInternalServerError)
			return
		}

		lastID, err := res.LastInsertId()
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Println("can't get id of the last insert object")
			http.Error(w, `"Internal Server Error"`, http.StatusInternalServerError)
			return
		}
	*/
	lastID, err := dbA.WrRow(CollectPtr.PointerDB, msgRx)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("%v\n", err)
		http.Error(w, `"Internal Server Error"`, http.StatusInternalServerError)
		return
	}

	CollectPtr.PointerLogI.Printf("was added to the database < Date: %v, Title: %v, Comment: %v, Repeat: %v >\n", msgRx.Date, msgRx.Title, msgRx.Comment, msgRx.Repeat)

	var sID udt.ReportAddFormat

	sID.ID = strconv.Itoa(int(lastID))

	dataSend, err := json.Marshal(sID)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("error marshal: %v\n", sID)
		http.Error(w, `"Internal Server Error"`, http.StatusInternalServerError)
		return
	}

	CollectPtr.PointerLogI.Println("request - Ok")
	w.WriteHeader(http.StatusCreated)
	w.Write(dataSend)
}

// Обработчик чтения задач из БД
func ReadTasksH(w http.ResponseWriter, r *http.Request) {

	CollectPtr.PointerLogI.Println("get request accepted to read tasks all")

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Подготовка аварийного сообщения в случае ошибки обработчика
	//
	/*
		dataSendError, err := generationErrMsg("ошибка запроса данных")
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("can't do generation error message: %v\n", err)
			http.Error(w, `"Internal Server Error"`, http.StatusInternalServerError)
			return
		}
	*/

	tn := time.Now()
	tnS := tn.Format("20060102")

	// Формирование запроса
	//
	/*
		sqlDB := CollectPtr.PointerDB

		rows, err := sqlDB.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE date >= :date ORDER BY date LIMIT 15 ",
			sql.Named("date", tnS))

		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("request error by id: %v, err: %v\n", tnS, err)
			w.Write(dataSendError)
			return
		}
		defer func() {
			_ = rows.Close()
		}()

		// Выворд содержмимого запроса
		//
		var t udt.TxFormat = udt.TxFormat{}

		for rows.Next() {

			task := udt.TxFormatEl{}

			err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

			if err != nil {
				CollectPtr.PointerLogI.Println("request - Error")
				CollectPtr.PointerLogE.Printf("scanning rows error: %v\n", err)
				w.Write(dataSendError)
				return
			}

			t.Tasks = append(t.Tasks, task)
		}

		if rows.Err() != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("scanning rows error: %v\n", err)
			w.Write(dataSendError)
			return
		}

		if t.Tasks == nil {
			t.Tasks = make([]udt.TxFormatEl, 0)
		}
	*/
	var t udt.TxFormat //= udt.TxFormat{}

	tasks, err := dbA.RdRows(CollectPtr.PointerDB, tnS)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("%v\n", err)
		http.Error(w, `"Internal Server Error"`, http.StatusInternalServerError)
		return
	}

	t.Tasks = tasks

	b, err := json.Marshal(t)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("marshal: %v has the error: %v\n", t, err)
		//w.Write(dataSendError)
		err = sendErrMsg("ошибка запроса данных", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	CollectPtr.PointerLogI.Println("request - Ok")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// Возврат всех параметров задачи по его ID
func GetTasksH(w http.ResponseWriter, r *http.Request) {

	CollectPtr.PointerLogI.Println("get request accepted to read the task by id")

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Подготовка аварийного сообщения если id не найдет
	//
	/*
		idNotFound, err := generationErrMsg("Задача не найдена")
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("can't do generation error message: %v\n", err)
			http.Error(w, `"Internal Server Error"`, http.StatusInternalServerError)
			return
		}
	*/

	// Подготовка аварийного сообщения если id не указан
	//
	/*
		idNot, err := generationErrMsg("Не указан идентификатор")
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("can't do generation error message: %v\n", err)
			http.Error(w, `"Internal Server Error"`, http.StatusInternalServerError)
			return
		}
	*/

	// Чтение параметра запроса
	//
	id := r.URL.Query().Get("id")

	idN, err := qualId(id)
	if err != nil {
		//w.Write(idNot)
		err = sendErrMsg("Не указан идентификатор", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	// Формирование запроса
	//

	//task := udt.TxFormatEl{}
	/*
		sqlDB := CollectPtr.PointerDB

		row := sqlDB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id ",
			sql.Named("id", idN))

		err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("scanning rows error: %v\n", err)
			w.Write(idNotFound)
			return
		}
	*/
	task, err := dbA.RdRow(CollectPtr.PointerDB, idN)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("%v\n", err)
		//w.Write(idNotFound)
		err = sendErrMsg("Задача не найдена", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	// Подготовка данных для передачи
	//
	b, err := json.Marshal(task)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("marshal: %v has the error: %v\n", task, err)
		http.Error(w, `"Internal Server Error"`, http.StatusInternalServerError)
		return
	}

	CollectPtr.PointerLogI.Println("request - Ok")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

// Сохранение данных задачи по его ID
func SaveTasksH(w http.ResponseWriter, r *http.Request) {

	CollectPtr.PointerLogI.Println("put request accepted to save the task by id")

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Подготовка аварийного сообщения в случае ошибки обработчика
	//
	/*
		dataSendError, err := generationErrMsg("Задача не найдена")
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("can't do generation error message: %v\n", err)
			http.Error(w, `"Internal Server Error"`, http.StatusInternalServerError)
			return
		}
	*/

	// Чтение тела
	//
	bodyReq, err := io.ReadAll(r.Body)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("read the body request err: %v\n", err)
		http.Error(w, `"Bad Request"`, http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	var msgRx udt.RxFormatFull

	err = json.Unmarshal(bodyReq, &msgRx)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("unmarshal body request error: %v\n", err)
		//w.Write(dataSendError)
		err = sendErrMsg("Задача не найдена", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	// Проверка поля id
	//
	id, err := qualId(msgRx.ID)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("bad quality id error: %v\n", err)
		//w.Write(dataSendError)
		err = sendErrMsg("Задача не найдена", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	// Проверка принятых данных в поле date
	//
	msgRx.Date, err = qualDate(msgRx.Date, msgRx.Repeat)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("bad quality date error: %v\n", err)
		//w.Write(dataSendError)
		err = sendErrMsg("Задача не найдена", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	// Проверка поля title
	//
	if msgRx.Title == "" {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Println("field title is empty")
		//w.Write(dataSendError)
		err = sendErrMsg("Задача не найдена", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	// Добавление данных задачи в БД
	//
	/*
		sqlDB := CollectPtr.PointerDB

		_, err = sqlDB.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
			sql.Named("date", msgRx.Date),
			sql.Named("title", msgRx.Title),
			sql.Named("comment", msgRx.Comment),
			sql.Named("repeat", msgRx.Repeat),
			sql.Named("id", id))

		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("update data <Date: %v, Title: %v, Comment: %v, Repeat: %v> by id: %v has error: %v\n", msgRx.Date, msgRx.Title, msgRx.Comment, msgRx.Repeat, id, err)
			w.Write(dataSendError)
			return
		}
	*/
	err = dbA.UpdRow(CollectPtr.PointerDB, msgRx)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("update data <Date: %v, Title: %v, Comment: %v, Repeat: %v> by id: %v has error: %v\n", msgRx.Date, msgRx.Title, msgRx.Comment, msgRx.Repeat, id, err)
		//w.Write(dataSendError)
		err = sendErrMsg("Задача не найдена", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	CollectPtr.PointerLogI.Println("request - Ok")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))

}

// Завершение задачи
func DoneTasksH(w http.ResponseWriter, r *http.Request) {

	CollectPtr.PointerLogI.Println("post request accepted to done the task by id")

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Подготовка аварийного сообщения в случае ошибки обработчика
	//
	/*
		dataSendError, err := generationErrMsg("Задача не найдена")
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("can't do generation error message: %v\n", err)
			http.Error(w, `"Internal Server Error"`, http.StatusInternalServerError)
			return
		}
	*/

	// Чтение параметра запроса
	//
	idQ := r.URL.Query().Get("id")

	id, err := qualId(idQ)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("bad quality id error: %v\n", err)
		//w.Write(dataSendError)
		err = sendErrMsg("Задача не найдена", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	// Получение данных задачи по его ID
	//
	/*
		var msg udt.RxFormatFull

		sqlDB := CollectPtr.PointerDB

		row := sqlDB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = :id",
			sql.Named("id", id))

		err = row.Scan(&msg.ID, &msg.Date, &msg.Title, &msg.Comment, &msg.Repeat)

		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("row scan error: %v\n", err)
			w.Write(dataSendError)
			return
		}
	*/
	sqlDB := CollectPtr.PointerDB

	msg, err := dbA.RdRow(sqlDB, id)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("%v\n", err)
		//w.Write(dataSendError)
		err = sendErrMsg("Задача не найдена", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	// Анализ содержимого в поле date после проверки
	// Если поле пустое - происходит удаление задачи по принятому параметру id
	// Если поле не пустое - сохраняются новые данные по принятому параметру id
	//
	if msg.Repeat == "" { // Удаление записи

		/*
			_, err = sqlDB.Exec("DELETE FROM scheduler WHERE id = :id",
				sql.Named("id", id))

			if err != nil {
				CollectPtr.PointerLogI.Println("request - Error")
				CollectPtr.PointerLogE.Printf("delete row by id: <%v> has error %v\n", id, err)
				w.Write(dataSendError)
				return
			}
		*/

		err = dbA.DlRow(sqlDB, id)
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("%v\n", err)
			//w.Write(dataSendError)
			err = sendErrMsg("Задача не найдена", w)
			if err != nil {
				CollectPtr.PointerLogI.Println("error send message")
				CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
			}
			return
		}

	} else { // Обновление записи

		// Проверка поля repeat
		//
		tn := time.Now()

		msg.Date, err = genDate.NextDate(tn, msg.Date, msg.Repeat)
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("error generation new date by Date: %v Repeat: %v. err: %v\n", msg.Date, msg.Repeat, err)
			//w.Write(dataSendError)
			err = sendErrMsg("Задача не найдена", w)
			if err != nil {
				CollectPtr.PointerLogI.Println("error send message")
				CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
			}
			return
		}

		/*
			_, err = sqlDB.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
				sql.Named("date", msg.Date),
				sql.Named("title", msg.Title),
				sql.Named("comment", msg.Comment),
				sql.Named("repeat", msg.Repeat),
				sql.Named("id", id))

			if err != nil {
				CollectPtr.PointerLogI.Println("request - Error")
				CollectPtr.PointerLogE.Printf("update data <Date: %v, Title: %v, Comment: %v, Repeat: %v> by id: %v has error: %v\n", msg.Date, msg.Title, msg.Comment, msg.Repeat, id, err)
				w.Write(dataSendError)
				return
			}
		*/
		var msg2 udt.RxFormatFull

		msg2.ID = msg.ID
		msg2.Date = msg.Date
		msg2.Title = msg.Title
		msg2.Comment = msg.Comment
		msg2.Repeat = msg.Repeat

		err = dbA.UpdRow(CollectPtr.PointerDB, msg2)
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("update data <Date: %v, Title: %v, Comment: %v, Repeat: %v> by id: %v has error: %v\n", msg2.Date, msg2.Title, msg2.Comment, msg2.Repeat, id, err)
			//w.Write(dataSendError)
			err = sendErrMsg("Задача не найдена", w)
			if err != nil {
				CollectPtr.PointerLogI.Println("error send message")
				CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
			}
			return
		}

	}

	CollectPtr.PointerLogI.Println("request - Ok")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))

}

// Удаление задачи
func DelTasksH(w http.ResponseWriter, r *http.Request) {

	CollectPtr.PointerLogI.Println("delete request accepted to delete the task by id")

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Подготовка аварийного сообщения в случае ошибки обработчика
	//
	/*
		dataSendError, err := generationErrMsg("Задача не найдена")
		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("can't do generation error message: %v\n", err)
			http.Error(w, `"Internal Server Error"`, http.StatusInternalServerError)
			return
		}
	*/

	// Чтение параметра запроса
	//
	idQ := r.URL.Query().Get("id")

	id, err := qualId(idQ)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("bad quality id error: %v\n", err)
		//w.Write(dataSendError)
		err = sendErrMsg("Задача не найдена", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	// Удаление задачи из БД
	//
	sqlDB := CollectPtr.PointerDB

	/*
		_, err = sqlDB.Exec("DELETE FROM scheduler WHERE id = :id",
			sql.Named("id", id))

		if err != nil {
			CollectPtr.PointerLogI.Println("request - Error")
			CollectPtr.PointerLogE.Printf("delete row by id: <%v> has error %v\n", id, err)
			w.Write(dataSendError)
			return
		}
	*/

	err = dbA.DlRow(sqlDB, id)
	if err != nil {
		CollectPtr.PointerLogI.Println("request - Error")
		CollectPtr.PointerLogE.Printf("%v\n", err)
		//w.Write(dataSendError)
		err = sendErrMsg("Задача не найдена", w)
		if err != nil {
			CollectPtr.PointerLogI.Println("error send message")
			CollectPtr.PointerLogE.Printf("send message error has the error: %v\n", err)
		}
		return
	}

	// Успешное завершение
	//
	CollectPtr.PointerLogI.Println("request - Ok")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))

}

// Функция возвращает массив байт и ошибку,
//
// Параметры:
//
// str - строка по данным которой формируется массив байт
/*
func generationErrMsg(str string) (b []byte, err error) {

	errMsg := udt.ErrMsgFormat{
		Error: str,
	}
	b, err = json.Marshal(errMsg)
	if err != nil {
		return nil, fmt.Errorf("error in generation error message : %w", err)
	}

	return b, nil
}
*/

// Функция производит проверку содержимого id. Возвращает id в формате int и признак качества отсутствием ошибки
//
// Параметры:
//
// id - строка с номером
func qualId(idQ string) (int, error) {

	if idQ == "" {
		return 0, errors.New("error with an empty value of idQ")
	}

	id, err := strconv.Atoi(idQ)
	if err != nil {
		return 0, fmt.Errorf("can't convert %s to int: %w", idQ, err)
	}

	return id, nil
}

// Функция производит проверку содержимого date. Возвращает дату в формате string и признак качества отсутствием ошибки
//
// Параметры:
//
// date - дата подлежащая обработке
// repeat - условие повторения
func qualDate(date, repeat string) (string, error) {

	tn := time.Now()

	if date == "" {
		return "", errors.New("error with an empty value of date")
	}

	_, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("can't convert %s to format 20060102 : %w", date, err)
	}

	dateNew, err := genDate.NextDate(tn, date, repeat)
	if err != nil {
		return "", fmt.Errorf("can't generation new date by %s and repeat by %s : %w", date, repeat, err)

	}

	return dateNew, nil
}

// Функция передаёт текст ошибки в http.ResponseWriter.
//
// Параметры;
//
// msg - передаваемое сообщение
func sendErrMsg(msg string, w http.ResponseWriter) error {

	errMsg := udt.ErrMsgFormat{
		Error: msg,
	}

	b, err := json.Marshal(errMsg)
	if err != nil {
		return fmt.Errorf("error in generation error message : %w", err)
	}

	w.Write(b)
	return nil
}
