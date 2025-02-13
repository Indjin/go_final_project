package loger

import (
	"fmt"
	"log"
	"os"
)

// Функция создания логеров. Возвращает указатели на информационный, предупредительный и аларм логеры, их файлы и ошибку.
//
// Параметры:
//
// location - место размещения логеров
func CreateLogers(location string) (ptrI, ptrE *log.Logger, fileI, fileE *os.File, err error) {

	fileI, err = os.OpenFile(location+"log_info.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error open loger information file: %w", err)
	}

	fileE, err = os.OpenFile(location+"log_error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error open loger error file: %w", err)
	}

	// Создание логеров
	ptrI = log.New(fileI, "INFO:\t", log.Ldate|log.Ltime|log.Lshortfile)
	ptrE = log.New(fileE, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)

	return ptrI, ptrE, fileI, fileE, nil

}
