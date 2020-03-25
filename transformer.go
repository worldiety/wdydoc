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
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	text "text/template"
)

// A File maps between an original src file and
type File struct {
	parent      *Template
	srcFile     string
	dstFilename string
	transformer Transformer
}

func NewFile(parent *Template, fname string) (*File, error) {
	f := &File{}
	f.srcFile = fname
	f.parent = parent
	basePath := filepath.Base(fname)
	ext := filepath.Ext(basePath)
	switch strings.ToLower(ext) {
	case htmlTemplate:
		f.dstFilename = basePath[:len(basePath)-len(htmlTemplate)]
		tpl, err := parent.html.New(basePath).ParseFiles(f.srcFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse html template %s: %w", f.srcFile, err)
		}
		f.transformer = &HtmlTransformer{
			Name:     basePath,
			Template: tpl,
		}
	case textTemplate:
		f.dstFilename = basePath[:len(basePath)-len(textTemplate)]
		tpl, err := parent.text.New(basePath).ParseFiles(f.srcFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse text template %s: %w", f.srcFile, err)
		}
		f.transformer = &TextTransformer{
			Name:     basePath,
			Template: tpl,
		}
	default:
		f.dstFilename = basePath
		f.transformer = &CopyTransformer{SrcFilename: f.srcFile}
	}
	return f, nil
}

func (f *File) Apply(model interface{}) error {
	relativePath := f.srcFile[len(f.parent.dir):]
	dstFile := filepath.Join(f.parent.buildDir, filepath.Dir(relativePath), f.dstFilename)
	_ = os.MkdirAll(filepath.Dir(dstFile), os.ModePerm)
	out, err := os.OpenFile(dstFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create file %s: %w", dstFile, err)
	}
	defer func() {
		err := out.Close()
		if err != nil {
			fmt.Printf("failed to close %s: %v", dstFile, err)
		}
	}()
	return f.transformer.Transform(model, out)
}

// A Transformer takes the model as input and a writer as output and applies a content transformation on it.
type Transformer interface {
	Transform(model interface{}, out io.Writer) error
}

// A HtmlTransformer applies an html template on the current model
type HtmlTransformer struct {
	Name     string
	Template *html.Template
}

func (h *HtmlTransformer) Transform(model interface{}, out io.Writer) error {
	return h.Template.ExecuteTemplate(out, h.Name, model)
}

// A TextTransformer applies a text template on the current model
type TextTransformer struct {
	Name     string
	Template *text.Template
}

func (h *TextTransformer) Transform(model interface{}, out io.Writer) error {
	err := h.Template.ExecuteTemplate(out, h.Name, model)
	if err != nil {
		return fmt.Errorf("failed to apply text template for %s: %w", h.Name, err)
	}
	return nil
}

// A CopyTransformer just pipes an existing file through
type CopyTransformer struct {
	SrcFilename string
}

func (h *CopyTransformer) Transform(model interface{}, out io.Writer) error {
	in, err := os.OpenFile(h.SrcFilename, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("unable to open %s: %w", h.SrcFilename, err)
	}
	defer func() {
		err := in.Close()
		if err != nil {
			fmt.Printf("failed to close %s: %v", h.SrcFilename, err)
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("failed to copy: %s: %w", h.SrcFilename, err)
	}
	return nil
}

func Marshal(w *Workspace) ([]byte, error) {
	return json.Marshal(w.toJson())
}

func Unmarshal(b []byte) (*Workspace, error) {
	tmp := make(map[string]interface{})
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return nil, err
	}
	w := &Workspace{}
	w.fromJson(tmp)
	return w, nil
}

// UnmarshalFile decodes a json markup file
func UnmarshalFile(fname string) (*Workspace, error) {
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", fname, err)
	}
	return Unmarshal(b)
}
