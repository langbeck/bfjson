package main

import (
	"flag"
	"fmt"
	"go/types"
	"io"
	"log"
	"os"
	"strings"

	"github.com/langbeck/bfjson/pkg/engine/custom"
	"github.com/langbeck/bfjson/pkg/engine/fastjson"
	"github.com/langbeck/bfjson/pkg/goparser"
)

func defaultQualifier(pkg *types.Package) string {
	return pkg.Name()
}

type Engine func(w io.Writer, path, pkgname string, noformat bool) error

func engineCustom(w io.Writer, path, pkgname string, noformat bool) error {
	analyzer, err := custom.NewAnalyzer(goparser.DefaultContext, defaultQualifier)
	if err != nil {
		return fmt.Errorf("NewAnalyzer failed: %w", err)
	}

	// NOTE: A config structure passed to the code writter would probably be a
	// better place to put this.
	analyzer.PackageName = pkgname

	p, err := analyzer.ProcessPath(path)
	if err != nil {
		return fmt.Errorf("could not process path %q: %w", path, err)
	}

	if noformat {
		return p.WriteGenerated(w)
	}

	return p.WriteGeneratedFormatted(w)
}

func engineFastJSON(w io.Writer, path, pkgname string, noformat bool) error {
	analyzer, err := fastjson.NewAnalyzer(goparser.DefaultContext, defaultQualifier)
	if err != nil {
		return fmt.Errorf("NewAnalyzer failed: %w", err)
	}

	// NOTE: A config structure passed to the code writter would probably be a
	// better place to put this.
	analyzer.PackageName = pkgname

	p, err := analyzer.ProcessPath(path)
	if err != nil {
		return fmt.Errorf("could not process path %q: %w", path, err)
	}

	if noformat {
		return p.WriteGenerated(w)
	}

	return p.WriteGeneratedFormatted(w)
}

var (
	defaultEngine = "custom"
	engines       = map[string]Engine{
		defaultEngine: engineCustom,
		"fastjson":    engineFastJSON,
	}
)

func run() error {
	availableEnginesList := make([]string, 0, len(engines))
	for name := range engines {
		availableEnginesList = append(availableEnginesList, name)
	}

	availableEngines := strings.Join(availableEnginesList, ", ")

	var (
		flagEngine      = flag.String("engine", defaultEngine, "Engine used by the generated code. Available engines are: "+availableEngines)
		flagPackageName = flag.String("pkgname", "generated", "Destination package name where generated code will live.")
		flagPackage     = flag.String("pkg", ".", "Source package to be analyzed.")
		flagWritePath   = flag.String("write", "-", `Path to write the generated code. "-" writes to stdout.`)
		flagNoFormat    = flag.Bool("noformat", false, "Skip formatting of the generated code. It can be useful for troubleshooting.")
	)
	flag.Parse()

	engine, found := engines[*flagEngine]
	if !found {
		return fmt.Errorf("invalid engine: %s", *flagEngine)
	}

	var w io.Writer = os.Stdout
	if *flagWritePath != "-" {
		fp, err := os.OpenFile(*flagWritePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
		if err != nil {
			return err
		}

		defer fp.Close()

		w = fp
	}

	err := engine(w, *flagPackage, *flagPackageName, *flagNoFormat)
	if err != nil {
		return fmt.Errorf("processTypes failed: %w", err)
	}

	return nil
}

func main() {
	log.SetFlags(0)
	err := run()
	if err != nil {
		log.Fatalf("[ERROR] %v", err)
		return
	}
}
