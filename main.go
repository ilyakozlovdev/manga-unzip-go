package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/signintech/gopdf"
)

func main() {
	archiveName := flag.String("file", "", "file name")
	flag.Parse()
	outName := strings.ReplaceAll(*archiveName, ".zip", ".pdf")
	outName = strings.ReplaceAll(outName, ".\\", "")

	filenames, err := unzipArchive(*archiveName, "out")
	if err != nil {
		fmt.Println(err)
		return
	}
	generatePdf(filenames, outName)
	if err = os.RemoveAll("./out"); err != nil {
		panic(err)
	}
}

func unzipArchive(src string, destination string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}

	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destination, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s is an illegal filepath", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			err = os.Mkdir(fpath, os.ModePerm)
			if err != nil {
				return filenames, err
			}
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return filenames, err
		}

		if err = outFile.Close(); err != nil {
			return filenames, err
		}

		if err = rc.Close(); err != nil {
			return filenames, err
		}
	}

	return filenames, nil
}

func generatePdf(filenames []string, outName string) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	for _, filename := range filenames {
		pdf.AddPage()
		err := pdf.Image(filename, 0, 0, gopdf.PageSizeA4)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Generated file:", outName)
	err := pdf.WritePdf(outName)
	if err != nil {
		panic(err)
	}
}
