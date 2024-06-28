package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

//Объявляем константы, которые помогут переводить размер файлов из байтов.
const thousand float64 = 1000
const GB int64 = 1000 * 1000 * 1000
const MB int64 = 1000 * 1000
const KB int64 = 1000

//Структура, в которой будут записаны название, тип и размер файлов.
type FileInfo struct {
	Name string
	Type string
	Size int64
}

//Функция принимает срез файлов для сортировки от меньшего к большему.
func sortBySizeAsc(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size < files[j].Size
	})
}

//Функция принимает срез файлов и элементы сортируются в порядке убывания размера.
func sortBySizeDesc(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})
}

//Функция принимает строку root которая содержит информацию о размере каждой директории, начиная с корневой директории root.
func walk(root string) (map[string]int64, error) {

	//Создается пустая карта, которая будет содержать путь к директории и ее размер.
	directorySizes := make(map[string]int64)
	var mu sync.Mutex
	var wg sync.WaitGroup

	//Функция используется для рекурсивного прохода по директориям, начиная с корневой директории root.
	//Принимает путь к директории, функцию обратного вызова и возвращает ошибку, если что-то пошло не так.
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error { //
		if err != nil {
			fmt.Printf("Ошибка доступа %q: %v\n", path, err)
			return err
		}

		//Условие, если  на пути папка, то запускается горутина, которая обрабатывает эту папку
		//Для безопасности доступа к общей переменной directorySizes используется Lock и Unlock.
		if info.IsDir() {
			fmt.Printf("Проход директории: %s\n", path)
			wg.Add(1)
			go func(dirPath string) {
				defer wg.Done()
				var dirSize int64
				err := filepath.Walk(dirPath, func(subPath string, subInfo os.FileInfo, subErr error) error {
					if subErr != nil {
						fmt.Printf("Ошибка доступа %q: %v\n", subPath, subErr)
						return subErr
					}
					dirSize += subInfo.Size()
					return nil
				})
				if err != nil {
					fmt.Printf("Ошибка размера каталога %q: %v\n", dirPath, err)
				}
				mu.Lock()
				directorySizes[dirPath] = dirSize
				mu.Unlock()
			}(path)
		}

		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	wg.Wait()

	//Функция озвращает карту directorySizes, содержащую информацию о размере каждой директории и ошибку.
	return directorySizes, walkErr
}

func main() {
	//Запуск таймера выполнения программы
	start := time.Now()

	// Определение и парсинг флагов
	rootPtr := flag.String("root", "", "Укажите корневую папку")
	sortPtr := flag.String("sort", "asc", "Укажите порядок сортировки (asc или desc)")

	//Функция для помощи с вводом флагов
	flag.Usage = func() {
		fmt.Println("  --root string\tУкажите корневую папку\"")
		fmt.Println("  --sort string\tУкажите порядок сортировки (asc или desc)")
	}

	flag.Parse()

	if !flag.Parsed() {
		flag.Usage()
		os.Exit(1)
	}

	//Проверка правильности ввода флагов
	if *sortPtr != "asc" && *sortPtr != "desc" {
		fmt.Println("Ошибка: неверный порядок сортировки. Укажите asc или desc для флага --sort.")
		flag.Usage()
		os.Exit(1)
	}

	//Происходит получение списка файлов в указанной директории, используя filepath.Glob.
	//Эта функция возвращает список файлов в указанной директории, соответствующих шаблону * (все файлы).
	files, err := filepath.Glob(filepath.Join(*rootPtr, "*"))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	File_Info := make([]FileInfo, 0)

	//Вызывается функция walk, которая рекурсивно обходит все поддиректории начиная с указанной корневой директории.
	//Функция возвращает карту, где ключами являются пути к директориям, а значениями - их размеры. А так же возвращается ошибка и обрабатывается.
	directories, walkErr := walk(*rootPtr)
	if walkErr != nil {
		fmt.Println("Ошибка при обходе файловой системы:", walkErr)
		os.Exit(1)
	}

	//Происходит чтение информации о файлах
	//Для каждого файла вызывается os.Stat, чтобы получить информацию о файле (размер, тип и имя).
	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err != nil {
			fmt.Println("Ошибка чтения информации о файле:", err)
			continue
		}
		//Определяется тип файла (файл или директория) и размер файла.
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

		//Срез File_Info содержит информацию о всех файлах и директориях в указанной директории имя,тип и размеры.
		File_Info = append(File_Info, FileInfo{
			Name: filepath.Base(file),
			Type: fileType,
			Size: fileSize,
		})
	}

	//Сортировка файлов
	if *sortPtr == "asc" {
		sortBySizeAsc(File_Info)
	} else if *sortPtr == "desc" {
		sortBySizeDesc(File_Info)
	}

	//Вывод информации о файлах
	for _, fileInfo := range File_Info {
		var sizeStr string
		switch {
		case fileInfo.Size >= GB:
			sizeStr = fmt.Sprintf("%.2f GB", float64(fileInfo.Size)/(thousand*thousand*thousand))
		case fileInfo.Size >= MB:
			sizeStr = fmt.Sprintf("%.2f MB", float64(fileInfo.Size)/(thousand*thousand))
		case fileInfo.Size >= KB:
			sizeStr = fmt.Sprintf("%.2f KB", float64(fileInfo.Size)/thousand)
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
