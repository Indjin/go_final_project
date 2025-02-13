package main

import (
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"

	dbA "go_final_project/myLib/dataBase"
	httpH "go_final_project/myLib/httpH"
	logerA "go_final_project/myLib/loger"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {

	// Загрузка переменных окружения
	//
	err := godotenv.Load("./.example")
	if err != nil {
		log.Fatal(err)
	}

	// Логеры
	//
	logI, logE, fileI, fileE, err := logerA.CreateLogers(os.Getenv("LOG_LOCATION"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = fileI.Close()
		_ = fileE.Close()
	}()

	httpH.CollectPtr.PointerLogI = logI // передача указателя на логер общей информации
	httpH.CollectPtr.PointerLogE = logE // передача указателя на логер ошибок
	dbA.CollectPtr.PointerLogI = logI   // передача указателя на логер общей информации
	dbA.CollectPtr.PointerLogE = logE   // передача указателя на логер ошибок

	// Подключение БД - Создание БД
	// Отключение от БД
	//
	sqlDB, err := dbA.CheckCreateDB()
	if err != nil {
		logE.Println("can't connect db")
		os.Exit(1)
	}
	defer func() {
		err = sqlDB.Close()
		if err != nil {
			logE.Println("can't close connect db")
		}
	}()

	httpH.CollectPtr.PointerDB = sqlDB // передача указателя на БД

	// Обработчики URL
	//
	r := chi.NewRouter()

	fs := http.FileServer(http.Dir(os.Getenv("HTTP_WEB")))
	r.Mount("/", fs) // добавляем обработку для статических файлов

	r.Get("/api/nextdate", httpH.NextDateH)    // Формирование новой даты
	r.Post("/api/task", httpH.AddTasksH)       // Добавление задач
	r.Get("/api/tasks", httpH.ReadTasksH)      // Чтение задач
	r.Get("/api/task", httpH.GetTasksH)        // Чтение параметров задачи по его ID
	r.Put("/api/task", httpH.SaveTasksH)       // Сохранение данных задачи по его ID
	r.Post("/api/task/done", httpH.DoneTasksH) // Завершение задачи
	r.Delete("/api/task", httpH.DelTasksH)     // Удаление задачи

	// Запуск сервера
	//
	logI.Println("starting the server on port:", os.Getenv("HTTP_PORT"))

	err = http.ListenAndServe(":"+os.Getenv("HTTP_PORT"), r)
	if err != nil {
		logE.Println("can't do start server")
		os.Exit(1)
	}

}
