package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
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

func walk(root string) map[string]int64 {
	directorySizes := make(map[string]int64)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Ошибка доступа%q: %v\n", path, err)
			return err
		}

		if info.IsDir() {
			fmt.Printf("Проход директории: %s\n", path)
			var dirSize int64
			err := filepath.Walk(path, func(subPath string, subInfo os.FileInfo, subErr error) error {
				if subErr != nil {
					fmt.Printf("Ошибка доступа  %q: %v\n", subPath, subErr)
					return subErr
				}
				dirSize += subInfo.Size()
				return nil
			})
			if err != nil {
				fmt.Printf("ошибка размера каталога %q: %v\n", path, err)
			}
			directorySizes[path] = dirSize
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Ошибка прохода%q: %v\n", root, err)
	}

	return directorySizes
}

func main() {
	start := time.Now()
	// Определение и парсинг флагов
	rootPtr := flag.String("root", "", "Укажите корневую папку")
	sortPtr := flag.String("sort", "asc", "Укажите порядок сортировки (asc или desc)")

	flag.Usage = func() {
		fmt.Println("  --root string\tУкажите корневую папку\"")
		fmt.Println("  --sort string\tУкажите порядок сортировки (asc или desc)")
	}

	flag.Parse()

	if !flag.Parsed() {
		flag.Usage()
		os.Exit(1)
	}

	// Проверка правильности ввода флагов
	if *sortPtr != "asc" && *sortPtr != "desc" {
		fmt.Println("Ошибка: неверный порядок сортировки. Укажите asc или desc для флага --sort.")
		flag.Usage()
		os.Exit(1)
	}

	// Получение списка файлов в указанной директории
	files, err := filepath.Glob(filepath.Join(*rootPtr, "*"))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	/*
		происходит получение списка файлов в указанной директории, используя filepath.Glob.
		Эта функция возвращает список файлов в указанной директории, соответствующих шаблону * (все файлы).
		Если происходит ошибка при получении списка файлов, выводится сообщение об ошибке и программа завершается.
	*/

	fileInfos := make([]FileInfo, 0) //Создается пустой срез fileInfos, который будет использоваться для хранения информации о файлах.
	directories := walk(*rootPtr)
	/*
		Вызывается функция walk, которая рекурсивно обходит все поддиректории начиная с указанной корневой директории
		и возвращает карту, где ключами являются пути к директориям, а значениями - их размеры.
	*/

	//--------------------------------------------------------------------------------------------------------------
	/*
		происходит чтение информации о каждом файле из списка файлов:
		    Для каждого файла вызывается os.Stat, чтобы получить информацию о файле (размер, тип и имя).
		    Если происходит ошибка при чтении информации о файле, выводится сообщение об ошибке и программа переходит к следующему файлу.
		    Определяется тип файла (файл или директория) и размер файла.
		    Если файл является директорией, то размер файла устанавливается равным размеру соответствующей директории
		    из карты directories, возвращенной функцией walk.
		    Информация о файле (имя, тип, размер) добавляется в срез fileInfos в виде структуры FileInfo.

		После завершения цикла чтения информации о файлах, срез fileInfos содержит информацию о всех файлах и директориях
		в указанной директории, включая их имена, типы (файл или директория) и размеры.
	*/
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
		// Получение размера папки, если это директория
		var fileSize int64
		if fileInfo.IsDir() {
			fileSize = directories[file]
		} else {
			fileSize = fileInfo.Size()
		}

		fileInfos = append(fileInfos, FileInfo{
			Name: filepath.Base(file),
			Type: fileType,
			Size: fileSize,
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
		var sizeStr string
		switch {
		case fileInfo.Size >= 1000*1000*1000:
			sizeStr = fmt.Sprintf("%.2f GB", float64(fileInfo.Size)/(1000*1000*1000))
		case fileInfo.Size >= 1000*1000:
			sizeStr = fmt.Sprintf("%.2f MB", float64(fileInfo.Size)/(1000*1000))
		case fileInfo.Size >= 1000:
			sizeStr = fmt.Sprintf("%.2f KB", float64(fileInfo.Size)/1000)
		default:
			sizeStr = fmt.Sprintf("%d bytes", fileInfo.Size)
		}

		output := fmt.Sprintf("Name: %s, Type: %s, Size: %s", fileInfo.Name, fileInfo.Type, sizeStr)
		fmt.Println(output)
	}

	//остановка счётчика и вывод
	elapsed := time.Since(start)
	fmt.Printf("Время выполнения программы: %s\n", elapsed)

}
