package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Структура для хранения информации о файлах
type FileInfo struct {
	Name string
	Type string
	Size int64
}

func sortBySizeAsc(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size < files[j].Size
	})
}

func sortBySizeDesc(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})
}

func main() {
	// Определение и парсинг флагов
	rootPtr := flag.String("root", "", "Specify the root directory")
	sortPtr := flag.String("sort", "asc", "Specify the sort order (asc or desc)")

	flag.Usage = func() {
		fmt.Println("  --root string\tSpecify the root directory")
		fmt.Println("  --sort string\tSpecify the sort order (asc or desc)")
	}

	flag.Parse()

	if !flag.Parsed() {
		flag.Usage()
		os.Exit(1)
	}

	// Проверка правильности ввода флагов
	if *sortPtr != "asc" && *sortPtr != "desc" {
		fmt.Println("Error: Invalid sort order. Please specify 'asc' or 'desc' for the --sort flag.")
		flag.Usage()
		os.Exit(1)
	}

	// Получение списка файлов в указанной директории
	files, err := filepath.Glob(filepath.Join(*rootPtr, "*"))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fileInfos := make([]FileInfo, 0)

	// Чтение информации о файлах
	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err != nil {
			fmt.Println("Ошибка чтения информации о файле:", err)
			continue
		}

		fileType := "file"
		if fileInfo.IsDir() {
			fileType = "directory"
		}

		fileInfos = append(fileInfos, FileInfo{
			Name: filepath.Base(file),
			Type: fileType,
			Size: fileInfo.Size(),
		})
	}

	// Сортировка файлов
	if *sortPtr == "asc" {
		sortBySizeAsc(fileInfos)
	} else if *sortPtr == "desc" {
		sortBySizeDesc(fileInfos)
	}

	// Вывод информации о файлах
	for _, fileInfo := range fileInfos {
		output := fmt.Sprintf("Name: %s, Type: %s, Size: %d bytes", fileInfo.Name, fileInfo.Type, fileInfo.Size)
		fmt.Println(output)
	}
}
