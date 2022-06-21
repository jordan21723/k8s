package util

import (
	"io/ioutil"
	"os"
)

func ReadTxtFile(url string) ([]byte, error) {
	file, err := os.Open(url)
	if err != nil {
		return []byte(""), err
	}
	defer file.Close()
	stringData, err := ioutil.ReadAll(file)
	if err != nil {
		return []byte(""), err
	}
	return stringData, nil
}

func WriteTxtToFile(filePath, txtToWrite string) error {
	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.WriteString(txtToWrite)
	if err != nil {
		return err
	}
	return nil
}

func WriteTxtToFileByte(filePath string, txtToWrite []byte) error {
	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write(txtToWrite)
	if err != nil {
		return err
	}
	return nil
}

func AppendTxtToFile(filePath, txtToWrite string, permission os.FileMode) error {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, permission)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.WriteString(txtToWrite)
	if err != nil {
		return err
	}
	return nil
}

func CreateDirIfNotExists(dirLocation string) error {
	if _, err := os.Stat(dirLocation); os.IsNotExist(err) {
		return os.MkdirAll(dirLocation, os.ModeDir|0755)
	}
	return nil
}
