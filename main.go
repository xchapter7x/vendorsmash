package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	pathlib "path"
	"path/filepath"
	"strings"
)

func main() {
	flag.Parse()
	root := flag.Arg(0)
	target := flag.Arg(1)
	fmt.Println(fmt.Sprintf("smashing vendor @%s into %s", root, target))
	err := new(smasher).Run(root, target)
	fmt.Printf("errors returned %v\n", err)
}

type smasher struct {
	root   string
	target string
}

func (s *smasher) Run(root, target string) (err error) {
	s.root = root
	s.target = target
	if err = filepath.Walk(root, s.visit); err == nil {
		err = filepath.Walk(root, s.destroy)
	}
	return err
}

func (s *smasher) destroy(path string, f os.FileInfo, err error) error {
	if filepath.Base(path) == "vendor" {
		fmt.Printf("Hulk Crush: %s\n", path)
		err = os.RemoveAll(path)
	}
	return err
}

func (s *smasher) visit(path string, f os.FileInfo, err error) error {
	if filepath.Base(path) == "vendor" {
		fmt.Printf("Hulk Smash: %s\n", path)
		err = filepath.Walk(path, s.smash)
	}
	return err
}

func (s *smasher) smash(path string, f os.FileInfo, err error) error {
	p := strings.Split(path, "/vendor/")
	ref, _ := os.Stat(path)
	if !ref.IsDir() {
		err = Copy(pathlib.Join(s.target, p[len(p)-1]), path)
	} else {
		os.RemoveAll(path)
		os.MkdirAll(path, 0777)
	}
	return err
}

func Copy(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := SafeCreate(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

// SafeCreate creates a file, creating parent directories if needed
func SafeCreate(name ...string) (file *os.File, err error) {
	p, e := ensurePath(pathlib.Join(name...))

	if e != nil {
		return nil, e
	}
	return os.Create(p)
}

func ensurePath(path string) (string, error) {
	base := filepath.Dir(path)
	e, _ := Exists(base)
	if e {
		return path, nil
	}

	// Create missing directory recursively
	err := os.MkdirAll(base, 0777)
	return path, err
}

//Exists - check if the given path exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
