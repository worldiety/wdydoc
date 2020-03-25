/*
 * Copyright 2020 Torben Schinke
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package wdydoc

import (
	"fmt"
	html "html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	text "text/template"
)

const htmlTemplate = ".gohtml"
const textTemplate = ".tmpl"

type Template struct {
	dir      string
	buildDir string
	html     *html.Template
	text     *text.Template
	files    []*File
}

// ReadTemplate creates a project based on an existing and parsable template folder structure. Empty and hidden folders
// are ignored.
func ReadTemplate(dir string, buildDir string) (*Template, error) {
	prj := &Template{
		dir:      dir,
		html:     html.New("/html/"),
		text:     text.New("/text/"),
		buildDir: buildDir,
	}
	prj.text.Funcs(text.FuncMap{
		"escapeLatex": EscapeLatex,
		"typeOf":      typeOfName,
		"isType":      is,
		"str":         strOf,
	})

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk path %s: %w", path, err)
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") || path == buildDir {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			if info.Name() == ".DS_Store" {
				return nil
			}
			file, err := NewFile(prj, path)
			if err != nil {
				return fmt.Errorf("failed to scan file: %w", err)
			}
			prj.files = append(prj.files, file)
		}
		return nil
	})
	if err != nil {
		return prj, fmt.Errorf("failed to list template files: %w", err)
	}
	return prj, nil
}

// Build applies the model to the template project. In general, all files are just copied over, however *.gohtml
// and *.tmpl files are applied as html or text template definitions with the actual model. The resulting filename
// is without the template extension, e.g. myfile.tex.tmpl will result in a file named myfile.tex.
// The generated files from the template are returned.
func (p *Template) Build(model interface{}) ([]string, error) {
	dstDir := p.buildDir
	err := os.RemoveAll(dstDir)
	if err != nil {
		return nil, fmt.Errorf("failed to remove build dir %s: %w", dstDir, err)
	}
	err = os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create build dir %s: %w", dstDir, err)
	}
	for _, file := range p.files {
		err := file.Apply(model)
		if err != nil {
			return nil, fmt.Errorf("failed to build: %w", err)
		}
	}
	return p.autobuild()
}

func (p *Template) autobuild() ([]string, error) {
	if _, err := os.Stat(filepath.Join(p.buildDir, "latexmkrc")); err == nil {
		fmt.Println("latexmkrc")
		cmd := exec.Command("latexmk")
		cmd.Dir = p.buildDir
		cmd.Env = os.Environ()
		res, err := cmd.CombinedOutput()
		fmt.Println(string(res))
		if err != nil {
			return nil, fmt.Errorf("failed to build latex project in %s: %w", p.buildDir, err)
		}
		files, err := listRootFiles(p.buildDir)
		if err != nil {
			return nil, err
		}
		var paths []string
		for _, f := range files {
			if strings.HasSuffix(f, ".pdf") {
				paths = append(paths, f)
			}
		}
		return paths, nil
	} else {
		fmt.Println("autobuild not supported")
	}

	return listRootFiles(p.buildDir)
}

func listRootFiles(dir string) ([]string, error) {
	var res []string
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to list file from %s: %w", dir, err)
	}
	for _, f := range files {
		res = append(res, filepath.Join(dir, f.Name()))
	}
	return res, nil
}
