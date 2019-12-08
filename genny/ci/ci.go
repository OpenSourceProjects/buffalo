package ci

import (
	"fmt"
	"html/template"
	"path"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/genny/gogen"
	"github.com/gobuffalo/packr/v2"
)

// New generator for adding travis or gitlab
func New(opts *Options) (*genny.Generator, error) {
	g := genny.New()

	if err := opts.Validate(); err != nil {
		return g, err
	}

	box := packr.New("buffalo:genny:ci", "../ci/templates")

	var fname string
	switch opts.Provider {
	case "travis", "travis-ci":
		fname = "-dot-travis.yml.tmpl"
	case "gitlab", "gitlab-ci":
		if opts.App.WithPop {
			fname = "-dot-gitlab-ci.yml.tmpl"
		} else {
			fname = "-dot-gitlab-ci-no-pop.yml.tmpl"
		}
	case "github":
		fname = "github-ci.yml.tmpl"
	default:
		return g, fmt.Errorf("could not find a template for %s", opts.Provider)
	}

	g.Transformer(genny.Replace("-no-pop", ""))
	if opts.Provider == "github" {
		g.Transformer(genny.Replace("github-ci.yml", path.Join(".github", "workflows", "tests.yml")))
	} else {
		g.Transformer(genny.Dot())
	}

	f, err := box.FindString(fname)
	if err != nil {
		return g, err
	}

	g.File(genny.NewFileS(fname, f))

	data := map[string]interface{}{
		"opts": opts,
	}

	if opts.DBType == "postgres" {
		data["testDbUrl"] = "postgres://postgres:postgres@postgres:5432/" + opts.App.Name.File().String() + "_test?sslmode=disable"
	} else if opts.DBType == "mysql" {
		data["testDbUrl"] = "mysql://root:root@(mysql:3306)/" + opts.App.Name.File().String() + "_test?parseTime=true&multiStatements=true&readTimeout=1s"
	} else {
		data["testDbUrl"] = ""
	}

	helpers := template.FuncMap{}

	t := gogen.TemplateTransformer(data, helpers)
	g.Transformer(t)

	return g, nil
}
