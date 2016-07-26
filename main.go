package main

import (
	"container/list"
	"fmt"
	"github.com/itpkg/epub"
	"io/ioutil"
	"os"
	// "os/exec"
	"bufio"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {

	var dir string
	// fmt.Println(len(os.Args))
	if len(os.Args) <= 1 {
		dir = getCurrentDirectory()
	} else {
		dir = os.Args[1]
	}

	oFile, oErr := os.OpenFile("result.txt", os.O_CREATE|os.O_RDWR, 0666)
	if oErr != nil {
		fmt.Println("写结果文件出错，程序退出")
		os.Exit(1)
	}

	defer oFile.Close()
	writer := bufio.NewWriter(oFile)
	defer writer.Flush()
	writer.WriteString("检查的目录：" + dir + "\n")
	files, _ := WalkDir(dir, "epub")
	total := 0
	var content []string

	for i := 0; i < len(files); i++ {
		r1, l2 := validateEpubLink(files[i])
		if !r1 {
			total++
			writer.WriteString(files[i] + "\n")
			content = append(content, strconv.Itoa(total)+"."+files[i]+"\n")
			for e := l2.Front(); e != nil; e = e.Next() {
				msg := e.Value.(string)
				//writer.WriteString("\t" + msg + "\n")
				content = append(content, "\t"+msg+"\n")
			}
		}
	}
	writer.WriteString("\n有问题的文档数量:" + strconv.Itoa(total) + "\n")
	for i := 0; i < len(content); i++ {
		writer.WriteString(content[i])
	}
	fmt.Println("finished")
}

func validateEpubLink(path string) (bool, *list.List) {
	bk, err := epub.Open(path)
	if err != nil {
		return false, list.New()
	}
	defer bk.Close()
	manifest := bk.Opf.Manifest
	errors := list.New()
	var result bool = true
	for i := 0; i < len(manifest); i++ {
		_, e1 := bk.Open(manifest[i].Href)
		if e1 != nil {
			result = false
			errors.PushBack("\tmissing resource " + manifest[i].Href)
		}
	}
	ncx := bk.Ncx
	ps := ncx.Points
	files := bk.Files()

	for i := 0; i < len(ps); i++ {
		ps = append(ps, ps[i].Points...)
	}
	//fmt.Println(len(ps))
	for i := 0; i < len(ps); i++ {
		src := ps[i].Content.Src

		index := strings.Index(src, "#")
		if index < 0 {
			continue
		}
		anchor := Substr(ps[i].Content.Src, index+1, len(ps[i].Content.Src)+1)
		targetResource := Substr(ps[i].Content.Src, 0, index)
		chptExistd := false
		for j := 0; j < len(files); j++ {
			if strings.Index(files[j], targetResource) >= 0 {

				chptExistd = true
				r, e2 := bk.Open(targetResource)
				if e2 != nil {
					result = false
					errors.PushBack("\tcan not open resource: " + files[j] + "  err : " + e2.Error())
					continue
				}
				defer r.Close()
				resourceContent, e3 := ioutil.ReadAll(r)
				if e3 != nil {
					result = false
					errors.PushBack("\t " + e3.Error())
					continue
				}

				if strings.Index(string(resourceContent), anchor) >= 0 {
					continue
				} else {

					result = false
					errors.PushBack("\tcan not find link: " + src)
				}
				//fmt.Println(string(resourceContent))
			}
		}
		if !chptExistd {
			result = false
			//errors.PushBack("\tmissing html resource: " + targetResource)
		}
	}
	return result, errors
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

//获取指定目录及所有子目录下的所有文件，可以匹配后缀过滤。
func WalkDir(dirPth, suffix string) (files []string, err error) {
	files = make([]string, 0, 30)
	suffix = strings.ToUpper(suffix)                                                     //忽略后缀匹配的大小写
	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		//if err != nil { //忽略错误
		// return err
		//}
		if fi.IsDir() { // 忽略目录
			return nil
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, filename)
		}
		return nil
	})
	return files, err
}

func Substr(str string, start int, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}

	return string(rs[start:end])
}
