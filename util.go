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
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
)

const typeAttrName = "type"
const WorkspaceType = "workspace"
const DocumentType = "document"
const ChapterType = "chapter"
const AuthorType = "author"
const NewlineType = "newline"
const NewpageType = "newpage"
const ItalicType = "italic"
const BoldType = "bold"
const UnderlineType = "underline"
const CodeType = "code"
const ImageType = "image"
const TOCType = "toc"
const TitlepageType = "titlepage"
const TextType = "text"

func assertObjList(v interface{}) []map[string]interface{} {
	var res []map[string]interface{}
	if slice, ok := v.([]interface{}); ok {
		for _, o := range slice {
			if m, ok := o.(map[string]interface{}); ok {
				res = append(res, m)
			}
		}
	}
	return res
}

func toJson(genericSlice interface{}) []interface{} {
	if genericSlice == nil {
		return nil
	}
	slice := reflect.ValueOf(genericSlice)
	res := make([]interface{}, 0, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		item := slice.Index(i).Interface()
		res = append(res, item.(Discriminator).toJson())
	}
	return res
}

func fromJson(m map[string]interface{}) Discriminator {
	typeName := optString(m, typeAttrName)
	var obj Discriminator
	switch typeName {
	case WorkspaceType:
		obj = &Workspace{}
	case DocumentType:
		obj = &Document{}
	case AuthorType:
		obj = &Author{}
	case ChapterType:
		obj = &Chapter{}
	case TextType:
		obj = &Span{}
	case TOCType:
		obj = TOC()
	case NewlineType:
		obj = Newline()
	case ItalicType:
		obj = Italic()
	case BoldType:
		obj = Bold()
	case UnderlineType:
		obj = Underline()
	case CodeType:
		obj = &Code{}
	case ImageType:
		obj = &Image{}
	case TitlepageType:
		obj = TitlePage()
	case NewpageType:
		obj = Newpage()
	default:
		panic("unknown format type: " + typeName + " -> " + debugJson(m))
	}
	obj.fromJson(m)
	return obj
}
func optString(m map[string]interface{}, key string) string {
	if str, ok := m[key].(string); ok {
		return str
	}
	return ""
}

func optStringSlice(m map[string]interface{}, key string) []string {
	if str, ok := m[key].([]string); ok {
		return str
	}
	return nil
}

func optInt(m map[string]interface{}, key string) int {
	if i, ok := m[key].(int); ok {
		return i
	}
	if i, ok := m[key].(int64); ok {
		return int(i)
	}
	if i, ok := m[key].(float64); ok {
		return int(i)
	}
	return 0
}

type defaultType struct {
	name string
}

func (d defaultType) Type() string {
	return d.name
}

func (d defaultType) toJson() map[string]interface{} {
	m := make(map[string]interface{})
	m[typeAttrName] = d.Type()
	return m
}

func (d defaultType) fromJson(map[string]interface{}) {

}

type defaultBody struct {
	name string
	Body []Discriminator
}

func (d *defaultBody) Type() string {
	return d.name
}

func (d *defaultBody) toJson() map[string]interface{} {
	m := make(map[string]interface{})
	m[typeAttrName] = d.Type()
	m["body"] = toJson(d.Body)
	return m
}

func (d *defaultBody) fromJson(m map[string]interface{}) {
	d.Body = nil
	for _, obj := range assertObjList(m["body"]) {
		d.Body = append(d.Body, fromJson(obj))
	}
}

func optSet(m map[string]interface{}, key string, val interface{}) {
	if val == nil {
		delete(m, key)
	}
	if s, ok := val.(string); ok {
		if s == "" {
			delete(m, key)
		} else {
			m[key] = s
		}
	}
}

func debugJson(i interface{}) string {
	if i == nil {
		return "nil"
	}
	b, err := json.MarshalIndent(i, " ", " ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

// CopyDir copies a whole directory recursively
func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// CopyFile copies a single file from src to dst
func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func IsDir(p string) bool {
	if stat, err := os.Stat(p); err == nil {
		return stat.IsDir()
	}
	return false
}
