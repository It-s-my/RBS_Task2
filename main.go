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

const thousand float64 = 1000
const GB int64 = 1000 * 1000 * 1000
const MB int64 = 1000 * 1000
const KB int64 = 1000

// Структура для хранения информации о файлах

type FileInfo struct {
	Name string
	Type string
	Size int64
}

/*
sortBySizeAsc принимает срез файлов. Внутри используется функция sort.Slice,
которая принимает срез и функцию сравнения для сортировки элементов среза.
*/

func sortBySizeAsc(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size < files[j].Size
	})
}

/*
Функция sortBySizeDesc имеет аналогичную структуру, но в этой функции элементы сортируются в порядке убывания размера.
*/

func sortBySizeDesc(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})
}

/*Функция walk принимает строку root в качестве аргумента и возвращает карту map[string]int64,
которая содержит информацию о размере каждой директории, начиная с корневой директории root.*/
func walk(root string) (map[string]int64, error) {
	directorySizes := make(map[string]int64) // создается пустая карта directorySizes, которая будет содержать путь к директории и ее размер.
	var mu sync.Mutex
	var wg sync.WaitGroup

	//используется для рекурсивного прохода по директориям, начиная с корневой директории root.
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error { //
		if err != nil {
			fmt.Printf("Ошибка доступа %q: %v\n", path, err)
			return err
		} //принимает путь к директории, функцию обратного вызова и возвращает ошибку, если что-то пошло не так.

		//Проверка, если папка, то запускается горутина
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
				mu.Unlock() //Для безопасности доступа к общей переменной используется Lock и Unlock.
			}(path)
		}

		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	wg.Wait()

	return directorySizes, walkErr //возвращает карту directorySizes, содержащую информацию о размере каждой директории.
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
	/*
		происходит получение списка файлов в указанной директории, используя filepath.Glob.
		Эта функция возвращает список файлов в указанной директории, соответствующих шаблону * (все файлы).
	*/
	files, err := filepath.Glob(filepath.Join(*rootPtr, "*"))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	File_Info := make([]FileInfo, 0)
	directories, walkErr := walk(*rootPtr)
	if walkErr != nil {
		fmt.Println("Ошибка при обходе файловой системы:", walkErr)
		os.Exit(1)
	}

	/*
		Вызывается функция walk, которая рекурсивно обходит все поддиректории начиная с указанной корневой директории
		и возвращает карту, где ключами являются пути к директориям, а значениями - их размеры.
	*/

	// Чтение информации о файлах
	for _, file := range files {
		fileInfo, err := os.Stat(file) //Для каждого файла вызывается os.Stat, чтобы получить информацию о файле (размер, тип и имя).
		if err != nil {
			fmt.Println("Ошибка чтения информации о файле:", err)
			continue
		}

		fileType := "file" //Определяется тип файла (файл или директория) и размер файла.
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

		File_Info = append(File_Info, FileInfo{
			Name: filepath.Base(file),
			Type: fileType,
			Size: fileSize,
		}) /*срез filesInfo содержит информацию о всех файлах и директориях
		в указанной директории, включая их имена, типы (файл или директория) и размеры.*/
	}

	// Сортировка файлов
	if *sortPtr == "asc" {
		sortBySizeAsc(File_Info)
	} else if *sortPtr == "desc" {
		sortBySizeDesc(File_Info)
	}

	// Вывод информации о файлах
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
	elapsed := time.Since(start) //остановка счётчика и вывод
	fmt.Printf("Время выполнения программы: %s\n", elapsed)

}
