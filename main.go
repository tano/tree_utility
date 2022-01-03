package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

type ByName []os.DirEntry

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return strings.Compare(a[i].Name(), a[j].Name()) < 0 }

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(writer io.Writer, path string, printFiles bool) error {
	return renderDir(writer, path, printFiles, "")
}

func renderDir(writer io.Writer, path string, printFiles bool, ident string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Error occured while closing file, %v", err)
		}
	}(file)
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		dirs, err := os.ReadDir(path)

		if !printFiles {
			var onlyDirs []os.DirEntry
			for _, dir := range dirs {
				if dir.IsDir() {
					onlyDirs = append(onlyDirs, dir)
				}
			}
			dirs = onlyDirs
		}

		sort.Sort(ByName(dirs))
		if err != nil {
			return err
		}
		for dirIdx, dirEntry := range dirs {

			_, err = fmt.Fprintf(writer, "%s", ident)
			if err != nil {
				return err
			}

			//fmt.Printf("Building for ident %q and path %s/%s\n", ident, path, dirEntry.Name())
			var currentLevelIdent, prefix string
			if dirIdx == len(dirs)-1 {
				//fmt.Printf("Dir entry name: %q, dir idx: %d", dirEntry.Name(), dirIdx)
				currentLevelIdent = "\t"
				prefix = "└───"
			} else {
				currentLevelIdent = "│\t"
				prefix = "├───"
			}

			var suffix string

			if dirEntry.IsDir() != true {
				info, err := dirEntry.Info()
				if err != nil {
					return err
				}
				size := info.Size()
				if size == 0 {
					suffix = " (empty)"
				} else {
					suffix = fmt.Sprintf(" (%db)", size)
				}
			}

			_, err := fmt.Fprintf(writer, "%s%s%s\n", prefix, dirEntry.Name(), suffix)
			if err != nil {
				return err
			}

			childDir := fmt.Sprintf("%s/%s", path, dirEntry.Name())
			err = renderDir(writer, childDir, printFiles, ident+currentLevelIdent)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
