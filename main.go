package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Структура для хранения информации о файлах

type FileInfo struct {
	Name string
	Type string
	Size int64
}

/*
sortBySizeAsc принимает слайс файлов files типа []FileInfo. Внутри функции используется функция sort.Slice,
которая принимает слайс и функцию сравнения для сортировки элементов слайса.
Функция сравнения задается анонимной функцией, которая сравнивает размеры файлов files[i].Size и files[j].Size.
Если размер файла i меньше размера файла j,
то возвращается true, что указывает на то, что элементы должны быть поменяны местами.
*/

func sortBySizeAsc(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size < files[j].Size
	})
}

/*
Функция sortBySizeDesc имеет аналогичную структуру, но в этой функции элементы сортируются в порядке убывания размера.
То есть, если размер файла i больше размера файла j, то возвращается true, чтобы элементы были поменяны местами.
Обе функции изменяют исходный слайс files, поэтому после вызова одной из этих функций слайс будет отсортирован в соответствии
с заданным порядком размеров файлов.
*/

func sortBySizeDesc(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})
}

/*Функция walk принимает строку root в качестве аргумента и возвращает карту map[string]int64,
которая содержит информацию о размере каждой директории, начиная с корневой директории root.*/

/*Как работает walk
Сначала в функции создается пустая карта directorySizes, которая будет содержать путь к директории и ее размер.
Затем создаются две переменные: mu типа sync.Mutex для синхронизации доступа к общим данным и wg типа sync.WaitGroup
для ожидания завершения всех горутин.

Функция filepath.Walk используется для рекурсивного прохода по директориям, начиная с корневой директории root.
Внутри этой функции передается анонимная функция, которая будет вызываться для каждого элемента в директории.
Если происходит ошибка доступа к файлу или директории, она обрабатывается и выводится сообщение об ошибке.
Если текущий элемент является директорией, то увеличивается счетчик wg и запускается новая горутина с анонимной функцией.
В этой функции происходит рекурсивный проход по всем элементам внутри директории и вычисление общего размера.
После завершения прохода по директории считанный размер сохраняется в карте directorySizes.

Для безопасности доступа к общим данным используется методы Lock и Unlock мьютекса mu.
Если происходит ошибка при вычислении размера директории, выводится соответствующее сообщение.
По завершении прохода по всем директориям ожидается завершение всех горутин с помощью метода Wait у WaitGroup.
В конце функция возвращает карту directorySizes, содержащую информацию о размере каждой директории.
*/

func walk(root string) map[string]int64 {
	directorySizes := make(map[string]int64)
	var mu sync.Mutex
	var wg sync.WaitGroup

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Ошибка доступа %q: %v\n", path, err)
			return err
		}

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

	if err != nil {
		fmt.Printf("Ошибка прохода %q: %v\n", root, err)
	}

	wg.Wait()

	return directorySizes
}

func main() {
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
		Если происходит ошибка при получении списка файлов, выводится сообщение об ошибке и программа завершается.
	*/
	files, err := filepath.Glob(filepath.Join(*rootPtr, "*"))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fileInfos := make([]FileInfo, 0)
	directories := walk(*rootPtr)

	/*
		Вызывается функция walk, которая рекурсивно обходит все поддиректории начиная с указанной корневой директории
		и возвращает карту, где ключами являются пути к директориям, а значениями - их размеры.
	*/

	//--------------------------------------------------------------------------------------------------------------

	// Чтение информации о файлах
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

}
