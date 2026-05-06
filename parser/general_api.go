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
	name := strings.ToLower(ann.Name)

	if sub, ok := strings.CutPrefix(name, "contact."); ok {
		p.applyContactAnnotation(sub, ann.Raw)
		return nil
	}
	if sub, ok := strings.CutPrefix(name, "license."); ok {
		p.applyLicenseAnnotation(sub, ann.Raw)
		return nil
	}
	if sub, ok := strings.CutPrefix(name, "externaldocs."); ok {
		p.applyExternalDocsAnnotation(sub, ann.Raw)
		return nil
	}

	switch name {
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
	case "server":
		return p.parseServerAnnotation(ann)
	}
	return nil
}

func (p *Parser) applyContactAnnotation(field, raw string) {
	if p.spec.Info.Contact == nil {
		p.spec.Info.Contact = &spec.Contact{}
	}
	switch field {
	case "name":
		p.spec.Info.Contact.Name = raw
	case "url":
		p.spec.Info.Contact.URL = raw
	case "email":
		p.spec.Info.Contact.Email = raw
	}
}

func (p *Parser) applyLicenseAnnotation(field, raw string) {
	if p.spec.Info.License == nil {
		p.spec.Info.License = &spec.License{}
	}
	switch field {
	case "name":
		p.spec.Info.License.Name = raw
	case "url":
		p.spec.Info.License.URL = raw
	}
}

func (p *Parser) applyExternalDocsAnnotation(field, raw string) {
	if p.spec.Info.ExternalDocs == nil {
		p.spec.Info.ExternalDocs = &spec.ExternalDocs{}
	}
	switch field {
	case "description":
		p.spec.Info.ExternalDocs.Description = raw
	case "url":
		p.spec.Info.ExternalDocs.URL = raw
	}
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

