package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func buildTree(prefix, path string, printFiles bool) ([]string, error) {
	var res, temp []string
	var s, pre, suffix string

	fileInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return res, errors.New("Can't read directory " + path)
	}

	var files []os.FileInfo
	for _, file := range fileInfo {
		if printFiles || file.IsDir() {
			files = append(files, file)
		}
	}

	length := len(files)
	for i := 0; i < length; i++ {
		if i < length-1 {
			pre = "├"
			suffix = "│\t"
		} else {
			pre = "└"
			suffix = "\t"
		}
		s = fmt.Sprintf("%s%s───%s", prefix, pre, files[i].Name())

		if files[i].IsDir() == false {
			if files[i].Size() > 0 {
				s += fmt.Sprintf(" (%db)", files[i].Size())
			} else {
				s += " (empty)"
			}
		}

		res = append(res, s+"\n")

		if files[i].IsDir() {
			temp, err = buildTree(prefix+suffix, path+string(os.PathSeparator)+files[i].Name(), printFiles)
			if err != nil {
				return res, err
			}
			res = append(res, temp...)
		}
	}
	return res, nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	res, err := buildTree("", path, printFiles)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, strings.Join(res, ""))
	return nil
}

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
