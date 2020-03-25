package main

import (
	"flag"
	"fmt"
	"github.com/worldiety/wdydoc"
	"os"
)

func main() {
	fmt.Printf("wdydoc version '%s'\n", wdydoc.BuildGitCommit)
	help := flag.Bool("help", false, "shows this help")
	format := flag.String("format", "json", "the input format type for the file of 'in'")
	in := flag.String("in", "", "the input markup file, as defined by 'format'")
	out := flag.String("out", "", "the folder to place the generated files")
	id := flag.String("id", "", "the id of the subtree to use for generation")
	template := flag.String("template", "", "the local folder or remote git repository containing the template")
	name := flag.String("name", "", "the subfolder name in 'out', to place the generated output")

	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}

	if len(*in) == 0 || len(*template) == 0 {
		fmt.Printf("invalid parameters\nusage:\n\n")
		flag.PrintDefaults()
		os.Exit(-5)
	}

	if *format != "json" {
		fmt.Printf("only json is currently supported\n")
		os.Exit(-1)
	}

	w, err := wdydoc.UnmarshalFile(*in)
	if err != nil {
		fmt.Printf("cannot parse markup of '%s': %v\n", *in, err)
		os.Exit(-2)
	}

	build, err := wdydoc.NewBuild(w, *out)
	if err != nil {
		fmt.Printf("cannot create build: %v\n", err)
		os.Exit(-3)
	}
	build.AddRule(&wdydoc.BuildRule{
		Id:       *id,
		Template: *template,
		Name:     *name,
	})

	err = build.Apply()
	if err != nil {
		fmt.Printf("cannot apply build transformation: %v\n", err)
		os.Exit(-4)
	}
}
