package fastjson

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"log"
	"reflect"
	"strings"

	"github.com/langbeck/bfjson/pkg/engine/fastjson/internal/basictypes"
	"github.com/langbeck/bfjson/pkg/goparser"
	"github.com/langbeck/bfjson/pkg/internal"
)

// Type annotations
const (
	AnnotationRawMessage = "rawmessage"
)

type Analyzer struct {
	qf  types.Qualifier
	ctx *goparser.Context

	defaultRawMessage types.Type

	PackageName string
}

func NewAnalyzer(ctx *goparser.Context, qf types.Qualifier) (*Analyzer, error) {
	analyzer := &Analyzer{
		qf:  qf,
		ctx: ctx,
	}

	err := analyzer.loadResources()
	if err != nil {
		return nil, fmt.Errorf("could not load basic resources: %w", err)
	}

	return analyzer, nil
}

func (a *Analyzer) loadResources() error {
	jsonPkg, err := a.ctx.Import("encoding/json")
	if err != nil {
		return err
	}

	jsonRawMessage := jsonPkg.ObjectForName("RawMessage")
	if jsonRawMessage == nil {
		return fmt.Errorf("could not find encoding/json.RawMessage")
	}

	a.defaultRawMessage = jsonRawMessage.Type()
	return nil
}

func (a *Analyzer) ProcessPath(path string) (*Package, error) {
	pkg, err := a.ctx.Import(path)
	if err != nil {
		return nil, err
	}

	dotImport := pkg.Path()
	p := &Package{
		structMap: make(map[*goparser.Struct]*StructInfo),
		structs:   make([]*StructInfo, 0),
		imports:   make(map[string]struct{}),
		dotImport: &dotImport,

		analyzer: a,
		pkg:      pkg,
	}
	p.processTypes()

	return p, nil
}

type Package struct {
	structs   []*StructInfo
	structMap map[*goparser.Struct]*StructInfo
	imports   map[string]struct{}
	dotImport *string
	analyzer  *Analyzer
	pkg       *goparser.Package
}

func (p *Package) commonStructField(field *goparser.StructField) *StructFieldInfo {
	sf := &StructFieldInfo{
		Name:     field.Name,
		NameJSON: field.Name,

		TypeName: internal.TypeString(field.Type, p.analyzer.qf),
	}

	tags := reflect.StructTag(field.Tag)
	tag, ok := tags.Lookup("json")
	if ok {
		segs := strings.SplitN(tag, ",", 2)
		if len(segs) > 0 {
			sf.NameJSON = segs[0]
		}
	}

	defvalue, ok := tags.Lookup("default")
	if ok {
		sf.Default = &defvalue
	}

	bftag, ok := tags.Lookup("bfjson")
	if ok {
		for _, opt := range strings.Split(bftag, ",") {
			switch opt {
			default:
				log.Printf("[WARN] unknow bfjson tag option %q", opt)
			}
		}
	}

	return sf
}

func (p *Package) decodeInfoForPointer(typ *types.Pointer) *DecodeInfo {
	etype := typ.Elem()

	// Check for pointers of basic types (e.g. *int)
	basic, _ := etype.(*types.Basic)
	if basic != nil {
		info := decodeInfoForBasicPtr(basic)
		return &info
	}

	// Check for custom types
	o := p.pkg.ObjectForType(etype)
	s, isStruct := o.(*goparser.Struct)
	if !isStruct {
		return nil
	}

	si := p.processStruct(s)
	if si == nil {
		return nil
	}

	return &DecodeInfo{
		DecoderRef: si.ObjectPtrDecoder,
		IsObject:   true,
		IsBasic:    false,
	}
}

func (p *Package) decodeInfoForSlice(typ *types.Slice) *DecodeInfo {
	etype := typ.Elem()
	basic, _ := etype.(*types.Basic)
	if basic != nil {
		info := decodeInfoForBasicSlice(basic)
		return &info
	}

	o := p.pkg.ObjectForType(etype)
	s, isStruct := o.(*goparser.Struct)
	if !isStruct {
		return nil
	}

	si := p.processStruct(s)
	if si == nil {
		return nil
	}

	return &DecodeInfo{
		DecoderRef: si.ObjectSliceDecoder,
		IsObject:   false,
		IsBasic:    false,
	}
}

func (p *Package) processStructField(field *goparser.StructField) *StructFieldInfo {
	sf := p.commonStructField(field)
	o := p.pkg.ObjectForType(field.Type)
	if o != nil {
		if o.HasAnnotation(AnnotationRawMessage) {
			sf.IsRawMessage = true
			return sf
		}

		if o.Implements(basictypes.JSONUnmarshaler) {
			sf.IsUnmarshaler = true
			return sf
		}

		s, isStruct := o.(*goparser.Struct)
		if isStruct {
			sf.DecodeInfo = decodeInfoForStruct(p.processStruct(s))
			return sf
		}

		// NOTE: would we ever reach this point?
	}

	switch field.Type.String() {
	case "encoding/json.RawMessage":
		sf.IsRawMessage = true
		return sf
	}

	switch gftype := field.Type.(type) {
	case *types.Basic:
		sf.DecodeInfo = decodeInfoForBasic(gftype)
		return sf

	case *types.Named:
		log.Printf("N?\t%-20s\t%-50s\ttype=%T", field.Name, gftype, gftype)
		return nil

	case *types.Pointer:
		info := p.decodeInfoForPointer(gftype)
		if info == nil {
			log.Printf("P?\t%-20s\t%-50s\tobj=%v", field.Name, gftype, p.pkg.ObjectForType(gftype.Elem()))
			return nil
		}

		sf.DecodeInfo = *info
		return sf

	case *types.Slice:
		info := p.decodeInfoForSlice(gftype)
		if info == nil {
			log.Printf("S?\t%-20s\t%-50s\tobj=%v", field.Name, gftype, p.pkg.ObjectForType(gftype.Elem()))
			return nil
		}

		sf.DecodeInfo = *info
		return sf

	default:
		log.Printf("?\t%-20s\t%-50s\ttype=%T", field.Name, gftype, gftype)
		return nil
	}
}

func (p *Package) processStructInto(s *goparser.Struct, si *StructInfo) {
	for _, field := range s.Fields() {
		if field.Embedded {
			o := p.pkg.ObjectForType(field.Type)
			if o == nil {
				panic("missing object definition (external object maybe?)")
			}

			ss := o.(*goparser.Struct)
			p.processStructInto(ss, si)
			continue
		}

		sf := p.processStructField(&field)
		if sf == nil {
			continue
		}

		si.Fields = append(si.Fields, sf)
	}
}

func (p *Package) processStruct(s *goparser.Struct) *StructInfo {
	si, found := p.structMap[s]
	if found {
		return si
	}

	log.Printf("[%s]", s.Name())
	name := s.Name()
	si = &StructInfo{
		Name:   name,
		Type:   s.QualifiedName(p.analyzer.qf),
		Fields: []*StructFieldInfo{},

		ObjectDecoder:         fmt.Sprintf("Decode_%s", name),
		ObjectPtrDecoder:      fmt.Sprintf("DecodePtr_%s", name),
		ObjectSliceDecoder:    fmt.Sprintf("DecodeSlice_%s", name),
		ObjectSlicePtrDecoder: fmt.Sprintf("DecodePtrSlice_%s", name),
		ObjectReleaser:        fmt.Sprintf("Release_%s", name),
		ObjectPool:            fmt.Sprintf("poolOf_%s", name),
	}

	p.processStructInto(s, si)

	pkgpath := s.Package().Path()
	p.imports[pkgpath] = struct{}{}
	p.structs = append(p.structs, si)
	p.structMap[s] = si
	return si
}

func (p *Package) processTypes() {
	for _, s := range p.pkg.Structs() {
		p.processStruct(s)
	}
}

func (p *Package) WriteGenerated(out io.Writer) error {
	type templateData struct {
		Imports     map[string]struct{}
		DotImport   *string
		PackageName string
	}

	err := templates.ExecuteTemplate(out, "header.gotmpl", templateData{
		Imports: p.imports,
		// DotImport:   p.dotImport,
		PackageName: p.analyzer.PackageName,
	})
	if err != nil {
		return err
	}

	for _, ref := range p.structs {
		text, err := ref.MarshalText()
		if err != nil {
			return err
		}

		_, err = out.Write(text)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Package) WriteGeneratedFormatted(out io.Writer) error {
	var b bytes.Buffer
	err := p.WriteGenerated(&b)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "generated", b.String(), parser.AllErrors|parser.ParseComments)
	if err != nil {
		return err
	}

	err = printer.Fprint(out, fset, f)
	if err != nil {
		return err
	}

	return nil
}

func decodeInfoForStruct(s *StructInfo) DecodeInfo {
	return DecodeInfo{
		DecoderRef: s.ObjectDecoder,
		IsObject:   true,
		IsBasic:    false,
	}
}

func decodeInfoForBasicSlice(typ *types.Basic) DecodeInfo {
	return DecodeInfo{
		DecoderRef: fmt.Sprintf("DecodeSliceOf%s", strings.Title(typ.Name())),
		IsObject:   false,
		IsBasic:    true,
	}
}

func decodeInfoForBasicPtr(typ *types.Basic) DecodeInfo {
	return DecodeInfo{
		DecoderRef: fmt.Sprintf("DecodePtr%s", strings.Title(typ.Name())),
		IsObject:   false,
		IsBasic:    true,
	}
}

func decodeInfoForBasic(typ *types.Basic) DecodeInfo {
	return DecodeInfo{
		DecoderRef: fmt.Sprintf("Decode%s", strings.Title(typ.Name())),
		IsObject:   false,
		IsBasic:    true,
	}
}
