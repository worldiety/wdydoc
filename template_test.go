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
	"testing"
)

func TestOpen(t *testing.T) {
	build, err := NewBuild(createModel(t), ".build/doc")
	if err != nil {
		t.Fatal(err)
	}
	build.AddRule(&BuildRule{
		Id:       "1234",
		Template: "/Users/tschinke/tmp/muondoc-wdy-book-01-latex",
		Name:     "mybook",
	})

	err = build.Apply()
	if err != nil {
		t.Fatal(err)
	}
}

func createModel(t *testing.T) *Workspace {
	t.Helper()

	ws := &Workspace{}
	ws.Title = "my workspace"
	ws.Version = "1.0.1"
	ws.Format = 1
	doc := ws.NewDocument()
	doc.Id = "1234"
	doc.Add(TitlePage(Text("my technical book"), Text("a subtitle")), TOC())
	chap := doc.NewChapter("my first chapter")

	chap.Text(
		`
		The inventory system consists of a login server, an inventory service and a web application.
		Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt 
		ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo 
		dolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. 
		Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut 
		labore et dolore magna aliquyam erat, sed diam voluptua.
		
		Span is aligned to chars at the left side. Empty lines are ignored.`)
	chap.Text("no space between")
	chap.Add(Newline())
	chap.Add(Text("hello "), Italic(Text("worl"), Bold(Underline(Text("d")))), Newline())
	chap.Add(Bold(Italic(Text(`ugly chars: & % $ # _ { } ~ ^ \`))), Newline())

	sub := chap.NewChapter("a section")
	sub.Text("This is a section within a chapter.")

	subsub := sub.NewChapter("a subsection")
	subsub.Text("This is another text but in a subsubsection.")

	chap = doc.NewChapter("another main chapter")
	chap.Text("typesetting test.")

	b, err := Marshal(ws)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))

	rereadWs, err := Unmarshal(b)
	if err != nil {
		t.Fatal(err)
	}

	return rereadWs
}
