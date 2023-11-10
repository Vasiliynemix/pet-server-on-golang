package logging

import (
	"archive/zip"
	"fmt"
	"go.uber.org/zap"
	"io"
	"os"
	"strings"
	"time"
)

func ClearLogFiles(pathToDebugLog string, pathToInfoLog string, logger *Logger) {
	logger.Info("Clearing log files started")

	maxFileSize := int64(500 * 1024 * 1024) // 500MB

	fileInfo := openFile(pathToInfoLog, logger)
	fileDebug := openFile(pathToDebugLog, logger)

	go monitorFile(fileInfo, maxFileSize, make(chan bool), pathToInfoLog, pathToDebugLog, logger)
	go monitorFile(fileDebug, maxFileSize, make(chan bool), pathToInfoLog, pathToDebugLog, logger)
}

func openFile(fileName string, logger *Logger) *os.File {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		logger.Info("Error opening file", zap.Error(err))
	}
	return file
}

func monitorFile(
	file *os.File,
	maxSize int64,
	quit chan bool,
	pathToInfoLog string,
	pathToDebugLog string,
	logger *Logger,
) {
	for {
		fileInfo, err := file.Stat()
		if err != nil {
			logger.Error("Error getting file info", zap.Error(err))
			break
		}

		if fileInfo.Size() >= maxSize {
			logger.Info("Clearing log file and creating new zip file...", zap.String("file", file.Name()))

			sourceFilePath := file.Name()
			zipFilePath := generateZipFileName(sourceFilePath, pathToInfoLog, pathToDebugLog)
			if err = createZipFile(sourceFilePath, zipFilePath); err != nil {
				logger.Info("Error creating zip file", zap.Error(err))
			}

			err = file.Truncate(0)
			if err != nil {
				logger.Info("Error clearing log file", zap.Error(err))
				break
			}
		}

		time.Sleep(5 * time.Second)
	}

	quit <- true
}

func generateZipFileName(sourceFilePath string, pathToInfoLog string, pathToDebugLog string) string {
	fileName := strings.Split(sourceFilePath, "/")[2]

	var logPath string

	switch fileName {
	case "debug.log":
		logPath = pathToDebugLog
	case "info.log":
		logPath = pathToInfoLog
	}

	oldFileName := strings.Split(logPath, "/")[2]
	newFileName := fmt.Sprintf("%s_%s", time.Now().Format("02-01-2006_15:04:05"), fileName)

	filePath := strings.Replace(logPath, oldFileName, newFileName, 1)
	zipFilePath := strings.Replace(filePath, ".log", ".zip", 1)
	return zipFilePath
}

func createZipFile(sourceFilePath, zipFilePath string) error {
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	info, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
