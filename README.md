# wdydoc ![wip](https://img.shields.io/badge/-work%20in%20progress-red) ![draft](https://img.shields.io/badge/-draft-red)
wdydoc is like hugo, but for technical documentation and a markup made for machines.
It uses the go stdlib template engine to create outputs in arbitrary text based formats
like Latex. Everything is organized in a single workspace, but you can create multiple
outputs with different templates from arbitrary subtrees using custom Ids.

## Why should I use it?
It is so easy and so much fun to create nice PDFs with a clean and proper typesetting using 
(existing) Latex templates. You need to work hard with the following 
substitutes, to get the same result (if ever):

* docbook
* markdown
* asciidoc
* html
* Latex
* DITA
* Bikeshed
* and many more...

Take a look at the [example file](example.pdf).

## usage
Normally you use the API, however you can also use it from the commandline:
```bash
# update the pkg in your gopath
# do not execute from within an existing go module
go get -u github.com/worldiety/wdydoc/cmd/wdydoc

# install with go install into ~/go/bin
go install github.com/worldiety/wdydoc/cmd/wdydoc

# finally
wdydoc -id=1234 -in=example.json -out=.build -template=https://github.com/worldiety/tmpl-doc-latex-book-01.git    
```

## API
The main use case is to generate documents by source code:

```go
ws := &Workspace{}
ws.Title = "my workspace"
ws.Version = "1.0.1"
ws.Format = 1
doc := ws.NewDocument()
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
```


## Markup interchange format
The default serialization format is currently JSON. For the above sample, it looks like
this (never try to write it by hand, either use the API or e.g. an importer for Markdown):

```json
{
  "format": 1,
  "resources": [
    {
      "authors": [],
      "body": [
        {
          "body": [
            {
              "type": "text",
              "value": "my technical book"
            },
            {
              "type": "text",
              "value": "a subtitle"
            }
          ],
          "type": "titlepage"
        },
        {
          "type": "toc"
        },
        {
          "body": [
            {
              "type": "text",
              "value": "\n\t\tThe inventory system consists of a login server, an inventory service and a web application.\n\t\tLorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt \n\t\tut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo \n\t\tdolores et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. \n\t\tLorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut \n\t\tlabore et dolore magna aliquyam erat, sed diam voluptua.\n\t\t\n\t\tSpan is aligned to chars at the left side. Empty lines are ignored."
            },
            {
              "type": "text",
              "value": "no space between"
            },
            {
              "type": "newline"
            },
            {
              "type": "text",
              "value": "hello "
            },
            {
              "body": [
                {
                  "type": "text",
                  "value": "worl"
                },
                {
                  "body": [
                    {
                      "body": [
                        {
                          "type": "text",
                          "value": "d"
                        }
                      ],
                      "type": "underline"
                    }
                  ],
                  "type": "bold"
                }
              ],
              "type": "italic"
            },
            {
              "type": "newline"
            },
            {
              "body": [
                {
                  "body": [
                    {
                      "type": "text",
                      "value": "ugly chars: & % $ # _ { } ~ ^ \\"
                    }
                  ],
                  "type": "italic"
                }
              ],
              "type": "bold"
            },
            {
              "type": "newline"
            },
            {
              "body": [
                {
                  "type": "text",
                  "value": "This is a section within a chapter."
                },
                {
                  "body": [
                    {
                      "type": "text",
                      "value": "This is another text but in a subsubsection."
                    }
                  ],
                  "level": 2,
                  "title": "a subsection",
                  "type": "chapter"
                }
              ],
              "level": 1,
              "title": "a section",
              "type": "chapter"
            }
          ],
          "level": 0,
          "title": "my first chapter",
          "type": "chapter"
        },
        {
          "body": [
            {
              "type": "text",
              "value": "typesetting test."
            }
          ],
          "level": 0,
          "title": "another main chapter",
          "type": "chapter"
        }
      ],
      "title": "",
      "type": "document"
    }
  ],
  "title": "my workspace",
  "type": "workspace",
  "version": "1.0.1"
}
```