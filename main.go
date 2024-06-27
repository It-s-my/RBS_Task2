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

func walk(root string) map[string]int64 {
	directorySizes := make(map[string]int64)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() {
			fmt.Printf("Visited directory: %s\n", path)
			var dirSize int64
			err := filepath.Walk(path, func(subPath string, subInfo os.FileInfo, subErr error) error {
				if subErr != nil {
					fmt.Printf("Error accessing path %q: %v\n", subPath, subErr)
					return subErr
				}
				dirSize += subInfo.Size()
				return nil
			})
			if err != nil {
				fmt.Printf("Error calculating size for directory %q: %v\n", path, err)
			}
			directorySizes[path] = dirSize
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", root, err)
	}

	return directorySizes
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
	directories := walk(*rootPtr)
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
		output := fmt.Sprintf("Name: %s, Type: %s, Size: %d bytes", fileInfo.Name, fileInfo.Type, fileInfo.Size)
		fmt.Println(output)
	}

}
