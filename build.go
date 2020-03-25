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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// A Build describes which workspace to build and how.
// It uses build rules to generate specific outputs from sub trees of the workspace.
type Build struct {
	workspace *Workspace   // actual model
	dir       string       // dir to generate the output into
	rules     []*BuildRule // the rules to apply the transformation on
	tmpDir    string       // downloaded resources are put here
}

func NewBuild(w *Workspace, dir string) (*Build, error) {
	tmp, err := ioutil.TempDir("", "wdydoc")
	if err != nil {
		return nil, fmt.Errorf("tmp dir required: %w", err)
	}
	return &Build{
		workspace: w,
		dir:       dir,
		tmpDir:    tmp,
	}, nil
}

func (b *Build) AddRule(r *BuildRule) {
	b.rules = append(b.rules, r)
}

func (b *Build) Apply() error {
	for _, r := range b.rules {
		template, err := b.provideTemplate(r.Template)
		if err != nil {
			return fmt.Errorf("unable to provide template: %w", err)
		}
		objRoot := b.workspace.ById(r.Id)
		if objRoot == nil {
			return fmt.Errorf("workspace does not contain '%s'", r.Id)
		}

		tmp := sha256.Sum224([]byte(r.Id + r.Template))
		transformTmpDir := filepath.Join(b.tmpDir, "transform", hex.EncodeToString(tmp[:]))

		tpl, err := ReadTemplate(template, transformTmpDir)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", template, err)
		}
		files, err := tpl.Build(objRoot)
		targetDir := filepath.Join(b.dir, r.Name)

		err = os.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("mkdir %s failed: %w", targetDir, err)
		}

		for _, f := range files {
			dst := filepath.Join(targetDir, filepath.Base(f))
			if IsDir(f) {
				err := CopyDir(f, dst)
				if err != nil {
					return fmt.Errorf("failed to copy result folder: %w", err)
				}
			} else {
				err := CopyFile(f, dst)
				if err != nil {
					return fmt.Errorf("failed to copy result file: %w", err)
				}
			}
		}
	}
	return nil
}

// provideTemplate either clones a repository (or pulls from it) or just returns a local path
func (b *Build) provideTemplate(urlOrDir string) (string, error) {
	if isUrl(urlOrDir) {
		tmp := sha256.Sum224([]byte(urlOrDir))
		dstDir := filepath.Join(b.tmpDir, "template", hex.EncodeToString(tmp[:]))
		if _, err := os.Stat(dstDir); err == nil {
			err := b.exec(dstDir, "git", "pull")
			if err != nil {
				return "", err
			}
			return dstDir, nil
		}
		err := os.MkdirAll(dstDir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to create template clone folder %s: %w", dstDir, err)
		}

		err = b.exec(dstDir, "git", "clone", urlOrDir, ".")
		if err != nil {
			return "", err
		}
		return dstDir, nil
	}
	if _, err := os.Stat(urlOrDir); err != nil {
		return "", fmt.Errorf("cannot find template %s: %w", urlOrDir, err)
	}
	return urlOrDir, nil
}

func (b *Build) exec(dir string, name string, args ...string) error {
	str := "cd " + dir + " && " + name + " " + strings.Join(args, " ")
	fmt.Println(str)
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	res, err := cmd.CombinedOutput()
	fmt.Println(string(res))
	if err != nil {
		return fmt.Errorf("'%s' failed: %w", str, err)
	}
	return nil
}

func isUrl(str string) bool {
	str = strings.ToLower(str)
	return strings.HasPrefix(str, "http")
}

// A BuildRules describes a (sub) tree of a workspace, which should be processed.
type BuildRule struct {
	Id       string // Id of the root to apply
	Template string // Template, either a local directory or an http/https git repository
	Name     string // Name of the target folder in the build directory. The entire template result just copied over.
}
