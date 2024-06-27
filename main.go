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

/*
sortBySizeAsc принимает слайс файлов files типа []FileInfo. Внутри функции используется функция sort.Slice,
которая принимает слайс и функцию сравнения для сортировки элементов слайса.
Функция сравнения задается анонимной функцией, которая сравнивает размеры файлов files[i].Size и files[j].Size.
Если размер файла i меньше размера файла j,
то возвращается true, что указывает на то, что элементы должны быть поменяны местами.
*/

func sortBySizeDesc(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})
}

/*
Функция sortBySizeDesc имеет аналогичную структуру, но в этой функции элементы сортируются в порядке убывания размера.
То есть, если размер файла i больше размера файла j, то возвращается true, чтобы элементы были поменяны местами.
Обе функции изменяют исходный слайс files, поэтому после вызова одной из этих функций слайс будет отсортирован в соответствии
с заданным порядком размеров файлов.
*/

/*
func walk(root string) map[string]int64
Этот код содержит функцию walk, которая принимает путь к корневой директории root и возвращает карту,
где ключами являются пути к директориям, а значениями - размеры соответствующих директорий.

Внутри функции создается пустая карта directorySizes, которая будет использоваться для хранения путей к директориям и их размеров.

Функция filepath.Walk используется для рекурсивного обхода всех файлов и поддиректорий, начиная с указанной корневой директории.
Внутри этой функции определяется анонимная функция, которая будет вызываться для каждого элемента в директории.
Внутри этой анонимной функции:
    Проверяется наличие ошибок при доступе к файлу или директории. Если ошибка есть, выводится сообщение об ошибке и возвращается эта ошибка.
    Если текущий элемент является директорией, то происходит вложенный вызов filepath.Walk для этой директории. Внутри вложенной
      функции определяется еще одна анонимная функция, которая суммирует размеры всех файлов внутри этой директории.
    Размер суммируется и сохраняется в переменной dirSize.
    Если во время обхода директории произошла ошибка, выводится сообщение об ошибке.
    Затем размер директории и путь к ней добавляются в карту directorySizes.
    При завершении обхода всех файлов и директорий функция возвращает карту directorySizes, содержащую пути к директориям и их размеры.
    В случае возникновения ошибки при обходе корневой директории также выводится сообщение об ошибке.
*/
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

}
