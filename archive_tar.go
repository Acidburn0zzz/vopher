package main

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// TarArchive handles tar archives
type TarArchive struct{}

func init() {

	suffixes := []string{".tar", ".tgz", ".tar.gz", ".tar.bz2"}
	supportedArchives = append(supportedArchives, suffixes...)

	archiveGuesser = append(archiveGuesser, func(n string) PluginArchive {

		switch path.Ext(n) {
		case ".tar":
			return &TarArchive{}
		case ".tar.gz", ".tgz":
			return &GzArchive{&TarArchive{}}
		case ".tar.bz2", ".tar.bzip2":
			return &BzipArchive{&TarArchive{}}
		}
		return nil
	})

}

func (ta *TarArchive) Extract(folder string, r io.Reader, stripDirs int) error {
	_, err := ta.handle(folder, r, stripDirs, tarExtractEntry)
	return err
}

func (ta *TarArchive) Entries(r io.Reader, stripDirs int) ([]string, error) {
	return ta.handle("", r, stripDirs, tarIgnoreEntry)
}

// small helper to operate on a tar-entry. reader r points directly
// to the data for 'name' in the tar file.
type tarExtractFunc func(name string, r io.Reader) error

// handle all file-like entries in the tar represented by 'r' due the 'extract'
// function.
// TODO: make sure "file-like" is the correct criteria.
func (ta *TarArchive) handle(folder string, r io.Reader, stripDirs int, extract tarExtractFunc) ([]string, error) {

	var (
		entries = make([]string, 0)
		reader  = tar.NewReader(r)
	)

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			continue
		}

		fi := header.FileInfo()
		if fi.IsDir() {
			// TODO: decide, if we skip dirs or not
			continue
		} else if strings.HasPrefix(header.Name, "/") {
			return nil, fmt.Errorf("entry with absolute filename %q", header.Name)
		}

		oname, isRoot := stripArchiveEntry(header.Name, stripDirs)
		if isRoot {
			continue
		}
		entries = append(entries, oname)
		if err := extract(filepath.Join(folder, oname), reader); err != nil {
			return nil, err
		}
	}
	return entries, nil
}

func tarExtractEntry(name string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(name), 0777); err != nil {
		return err
	}
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, r)
	return err
}
func tarIgnoreEntry(name string, r io.Reader) error {
	_, err := io.Copy(ioutil.Discard, r)
	return err
}
