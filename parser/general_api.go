package parser

import (
	"fmt"
	"strings"

	"github.com/polanski13/asyngo/spec"
)

func (p *Parser) parseGeneralAPI() error {
	mainFile := p.packages.GetFile(p.findMainFile())
	if mainFile == nil {
		return fmt.Errorf("main API file not found: %s", p.mainFile)
	}

	for _, decl := range mainFile.Decls {
		funcDecl, ok := asFuncDecl(decl)
		if !ok || funcDecl.Doc == nil {
			continue
		}

		annotations := newAnnotationSet(funcDecl.Doc)
		if !annotations.Has("AsyncAPI") && !annotations.Has("Title") {
			continue
		}

		for _, ann := range annotations.All() {
			if err := p.applyGeneralAnnotation(ann); err != nil {
				return err
			}
		}
		return nil
	}

	return fmt.Errorf("%w: expected @AsyncAPI or @Title on a function in %s", ErrMissingAnnotation, p.mainFile)
}

func (p *Parser) applyGeneralAnnotation(ann *annotation) error {
	switch strings.ToLower(ann.Name) {
	case "asyncapi":
		if ann.Raw != "" {
			p.spec.AsyncAPI = ann.Raw
		}
	case "title":
		p.spec.Info.Title = ann.Raw
	case "version":
		p.spec.Info.Version = ann.Raw
	case "description":
		p.spec.Info.Description = ann.Raw
	case "termsofservice":
		p.spec.Info.TermsOfService = ann.Raw
	case "defaultcontenttype":
		p.spec.DefaultContentType = ann.Raw
	case "id":
		p.spec.ID = ann.Raw
	case "contact.name":
		p.ensureContact()
		p.spec.Info.Contact.Name = ann.Raw
	case "contact.url":
		p.ensureContact()
		p.spec.Info.Contact.URL = ann.Raw
	case "contact.email":
		p.ensureContact()
		p.spec.Info.Contact.Email = ann.Raw
	case "license.name":
		p.ensureLicense()
		p.spec.Info.License.Name = ann.Raw
	case "license.url":
		p.ensureLicense()
		p.spec.Info.License.URL = ann.Raw
	case "externaldocs.description":
		p.ensureExternalDocs()
		p.spec.Info.ExternalDocs.Description = ann.Raw
	case "externaldocs.url":
		p.ensureExternalDocs()
		p.spec.Info.ExternalDocs.URL = ann.Raw
	case "server":
		return p.parseServerAnnotation(ann)
	}
	return nil
}

func (p *Parser) parseServerAnnotation(ann *annotation) error {
	if len(ann.Args) < 2 {
		return fmt.Errorf("@Server requires at least name and host: got %q: %w", ann.Raw, ErrInvalidAnnotation)
	}

	name := ann.Args[0]
	hostWithProtocol := ann.Args[1]

	host, protocol := parseHostProtocol(hostWithProtocol)

	var pathname string
	if len(ann.Args) >= 3 && ann.Args[2] != "-" {
		pathname = ann.Args[2]
	}

	var description string
	if len(ann.Args) >= 4 {
		description = strings.Join(ann.Args[3:], " ")
	}

	p.spec.Servers[name] = spec.Server{
		Host:        host,
		Protocol:    protocol,
		Pathname:    pathname,
		Description: description,
	}

	return nil
}

var knownProtocols = []string{"wss://", "ws://", "https://", "http://", "mqtt://", "amqp://", "kafka://"}

func parseHostProtocol(hostWithProtocol string) (host, protocol string) {
	for _, prefix := range knownProtocols {
		if after, ok := strings.CutPrefix(hostWithProtocol, prefix); ok {
			return after, strings.TrimSuffix(prefix, "://")
		}
	}
	return hostWithProtocol, "ws"
}

func (p *Parser) ensureContact() {
	if p.spec.Info.Contact == nil {
		p.spec.Info.Contact = &spec.Contact{}
	}
}

func (p *Parser) ensureLicense() {
	if p.spec.Info.License == nil {
		p.spec.Info.License = &spec.License{}
	}
}

func (p *Parser) ensureExternalDocs() {
	if p.spec.Info.ExternalDocs == nil {
		p.spec.Info.ExternalDocs = &spec.ExternalDocs{}
	}
}
