package parser

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/polanski13/asyngo/spec"
)

type wsBinding struct {
	Method        string
	QueryProps    map[string]*spec.SchemaRef
	QueryRequired []string
	HeaderProps   map[string]*spec.SchemaRef
}

type operationBuilder struct {
	ChannelAddress     string
	ChannelDescription string
	ChannelParams      map[string]spec.Parameter
	ChannelServers     []string
	WsBinding          *wsBinding
	Action             spec.Action
	OperationID        string
	Summary            string
	Description        string
	Tags               []spec.Tag
	Messages           []messageEntry
	OneOfMessages      []oneOfMessageEntry
	HasReply           bool
	ReplyMessages      []messageEntry
	ReplyChannelAddr   string
	Security           []spec.Reference
	Warnings           []string
}

type messageEntry struct {
	Name        string
	PayloadType string
}

type oneOfMessageEntry struct {
	Name          string
	Discriminator string
	PayloadTypes  []string
}

func newOperationBuilder() *operationBuilder {
	return &operationBuilder{
		ChannelParams: make(map[string]spec.Parameter),
	}
}

func (b *operationBuilder) ensureWsBinding() *wsBinding {
	if b.WsBinding == nil {
		b.WsBinding = &wsBinding{
			QueryProps:  make(map[string]*spec.SchemaRef),
			HeaderProps: make(map[string]*spec.SchemaRef),
		}
	}
	return b.WsBinding
}

func (p *Parser) parseHandlers() error {
	for path, f := range p.packages.Files() {
		for _, decl := range f.Decls {
			funcDecl, ok := asFuncDecl(decl)
			if !ok || funcDecl.Doc == nil {
				continue
			}

			annotations := newAnnotationSet(funcDecl.Doc)
			if !annotations.Has("Channel") && !annotations.Has("Operation") {
				continue
			}

			if err := p.parseHandlerFunc(funcDecl, f, annotations); err != nil {
				pos := p.packages.FileSet().Position(funcDecl.Pos())
				return newParseError(path, pos.Line, funcDecl.Name.Name, err)
			}
		}
	}
	return nil
}

func (p *Parser) parseHandlerFunc(funcDecl *ast.FuncDecl, file *ast.File, annotations *annotationSet) error {
	builder := newOperationBuilder()

	for _, ann := range annotations.All() {
		if err := p.applyHandlerAnnotation(builder, ann); err != nil {
			return err
		}
	}

	if len(builder.Warnings) > 0 {
		if p.strict {
			return fmt.Errorf("%s: %w", builder.Warnings[0], ErrUnknownType)
		}
		p.warnings = append(p.warnings, builder.Warnings...)
	}

	if builder.ChannelAddress == "" {
		return fmt.Errorf("@Channel is required when @Operation is present: %w", ErrMissingChannel)
	}

	channelKey := addressToKey(builder.ChannelAddress)

	if err := p.registerChannel(channelKey, builder); err != nil {
		return err
	}

	if err := p.registerMessages(channelKey, file, builder); err != nil {
		return err
	}

	if builder.OperationID != "" {
		if err := p.registerOperation(channelKey, builder); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) applyHandlerAnnotation(b *operationBuilder, ann *annotation) error {
	switch strings.ToLower(ann.Name) {
	case "channel":
		b.ChannelAddress = ann.Raw
	case "channeldescription":
		b.ChannelDescription = ann.Raw
	case "channelparam":
		return parseChannelParam(b, ann)
	case "channelserver":
		if len(ann.Args) > 0 {
			b.ChannelServers = append(b.ChannelServers, ann.Args[0])
		}
	case "wsbinding.method":
		if len(ann.Args) > 0 {
			b.ensureWsBinding().Method = strings.ToUpper(ann.Args[0])
		}
	case "wsbinding.query":
		return parseWsQueryParam(b, ann)
	case "wsbinding.header":
		return parseWsHeaderParam(b, ann)
	case "operation":
		if len(ann.Args) == 0 {
			return fmt.Errorf("@Operation requires action (send or receive): %w", ErrInvalidAnnotation)
		}
		action := strings.ToLower(ann.Args[0])
		switch action {
		case "send":
			b.Action = spec.ActionSend
		case "receive":
			b.Action = spec.ActionReceive
		default:
			return fmt.Errorf("@Operation: invalid action %q (expected send or receive): %w", action, ErrInvalidAction)
		}
	case "operationid":
		if len(ann.Args) > 0 {
			b.OperationID = ann.Args[0]
		}
	case "summary":
		b.Summary = ann.Raw
	case "description":
		b.Description = ann.Raw
	case "tags":
		b.Tags = parseTags(ann.Raw)
	case "message":
		if len(ann.Args) < 2 {
			return fmt.Errorf("@Message requires name and payload type: %w", ErrInvalidAnnotation)
		}
		b.Messages = append(b.Messages, messageEntry{
			Name:        ann.Args[0],
			PayloadType: ann.Args[1],
		})
	case "messageoneof":
		if len(ann.Args) < 2 {
			return fmt.Errorf("@MessageOneOf requires name and payload types (pipe-separated): %w", ErrInvalidAnnotation)
		}
		entry := oneOfMessageEntry{
			Name:         ann.Args[0],
			PayloadTypes: strings.Split(ann.Args[1], "|"),
		}
		for _, arg := range ann.Args[2:] {
			if strings.HasPrefix(arg, "discriminator(") && strings.HasSuffix(arg, ")") {
				entry.Discriminator = arg[len("discriminator(") : len(arg)-1]
			}
		}
		b.OneOfMessages = append(b.OneOfMessages, entry)
	case "reply":
		b.HasReply = true
	case "replymessage":
		if len(ann.Args) < 2 {
			return fmt.Errorf("@ReplyMessage requires name and payload type: %w", ErrInvalidAnnotation)
		}
		b.ReplyMessages = append(b.ReplyMessages, messageEntry{
			Name:        ann.Args[0],
			PayloadType: ann.Args[1],
		})
	case "replychannel":
		b.ReplyChannelAddr = ann.Raw
	case "security":
		if len(ann.Args) > 0 {
			b.Security = append(b.Security, spec.NewRef("#/components/securitySchemes/"+ann.Args[0]))
		}
	default:
		b.Warnings = append(b.Warnings, fmt.Sprintf("unknown handler annotation @%s", ann.Name))
	}
	return nil
}

func parseChannelParam(b *operationBuilder, ann *annotation) error {
	if len(ann.Args) < 3 {
		return fmt.Errorf("@ChannelParam requires name, type, and required: %w", ErrInvalidAnnotation)
	}
	name := ann.Args[0]
	typeName := ann.Args[1]

	if _, known := mapSimpleType(typeName); !known {
		b.Warnings = append(b.Warnings, fmt.Sprintf("@ChannelParam %s: unknown type %q, defaulting to string", name, typeName))
	}

	param := spec.Parameter{
		Location: "$message.payload#/" + name,
	}

	if len(ann.Args) >= 4 {
		param.Description = ann.Args[3]
	}

	if len(ann.Args) > 4 {
		for _, arg := range ann.Args[4:] {
			if strings.HasPrefix(arg, "enum(") && strings.HasSuffix(arg, ")") {
				inner := arg[5 : len(arg)-1]
				param.Enum = strings.Split(inner, ",")
			}
			if strings.HasPrefix(arg, "example(") && strings.HasSuffix(arg, ")") {
				inner := arg[8 : len(arg)-1]
				param.Examples = strings.Split(inner, ",")
			}
		}
	}

	b.ChannelParams[name] = param
	return nil
}

func parseWsQueryParam(b *operationBuilder, ann *annotation) error {
	if len(ann.Args) < 3 {
		return fmt.Errorf("@WsBinding.Query requires name, type, and required: %w", ErrInvalidAnnotation)
	}
	name := ann.Args[0]
	typeName := ann.Args[1]
	required := strings.ToLower(ann.Args[2]) == "true"

	mapped, known := mapSimpleType(typeName)
	if !known {
		b.Warnings = append(b.Warnings, fmt.Sprintf("@WsBinding.Query %s: unknown type %q, defaulting to string", name, typeName))
	}

	prop := spec.NewInlineSchema(&spec.Schema{
		Type: mapped,
	})

	if len(ann.Args) >= 4 {
		prop.Schema.Description = ann.Args[3]
	}

	if len(ann.Args) > 4 {
		for _, arg := range ann.Args[4:] {
			if strings.HasPrefix(arg, "enum(") && strings.HasSuffix(arg, ")") {
				inner := arg[5 : len(arg)-1]
				vals := strings.Split(inner, ",")
				enums := make([]any, len(vals))
				for i, v := range vals {
					enums[i] = v
				}
				prop.Schema.Enum = enums
			}
		}
	}

	ws := b.ensureWsBinding()
	ws.QueryProps[name] = prop
	if required {
		ws.QueryRequired = append(ws.QueryRequired, name)
	}

	return nil
}

func parseWsHeaderParam(b *operationBuilder, ann *annotation) error {
	if len(ann.Args) < 3 {
		return fmt.Errorf("@WsBinding.Header requires name, type, and required: %w", ErrInvalidAnnotation)
	}
	name := ann.Args[0]
	typeName := ann.Args[1]

	mapped, known := mapSimpleType(typeName)
	if !known {
		b.Warnings = append(b.Warnings, fmt.Sprintf("@WsBinding.Header %s: unknown type %q, defaulting to string", name, typeName))
	}

	prop := spec.NewInlineSchema(&spec.Schema{
		Type: mapped,
	})

	if len(ann.Args) >= 4 {
		prop.Schema.Description = ann.Args[3]
	}

	b.ensureWsBinding().HeaderProps[name] = prop
	return nil
}

func (p *Parser) registerChannel(key string, b *operationBuilder) error {
	if existing, ok := p.spec.Channels[key]; ok {
		for _, msg := range b.Messages {
			if existing.Messages == nil {
				existing.Messages = make(map[string]spec.MessageRef)
			}
			existing.Messages[msg.Name] = spec.MessageRef{
				Ref: spec.ComponentMessageRef(msg.Name),
			}
		}
		for _, msg := range b.ReplyMessages {
			if existing.Messages == nil {
				existing.Messages = make(map[string]spec.MessageRef)
			}
			existing.Messages[msg.Name] = spec.MessageRef{
				Ref: spec.ComponentMessageRef(msg.Name),
			}
		}
		for _, msg := range b.OneOfMessages {
			if existing.Messages == nil {
				existing.Messages = make(map[string]spec.MessageRef)
			}
			existing.Messages[msg.Name] = spec.MessageRef{
				Ref: spec.ComponentMessageRef(msg.Name),
			}
		}
		p.spec.Channels[key] = existing
		return nil
	}

	channel := spec.Channel{
		Address:     b.ChannelAddress,
		Description: b.ChannelDescription,
		Messages:    make(map[string]spec.MessageRef),
		Parameters:  make(map[string]spec.Parameter),
	}

	for name, param := range b.ChannelParams {
		channel.Parameters[name] = param
	}

	for _, serverName := range b.ChannelServers {
		channel.Servers = append(channel.Servers, spec.NewRef(spec.ServerRef(serverName)))
	}

	for _, msg := range b.Messages {
		channel.Messages[msg.Name] = spec.MessageRef{
			Ref: spec.ComponentMessageRef(msg.Name),
		}
	}
	for _, msg := range b.ReplyMessages {
		channel.Messages[msg.Name] = spec.MessageRef{
			Ref: spec.ComponentMessageRef(msg.Name),
		}
	}
	for _, msg := range b.OneOfMessages {
		channel.Messages[msg.Name] = spec.MessageRef{
			Ref: spec.ComponentMessageRef(msg.Name),
		}
	}

	if ws := b.WsBinding; ws != nil {
		channel.Bindings = &spec.ChannelBindings{
			WS: &spec.WebSocketChannelBinding{
				Method:         ws.Method,
				BindingVersion: "0.1.0",
			},
		}
		if len(ws.QueryProps) > 0 {
			channel.Bindings.WS.Query = &spec.Schema{
				Type:       "object",
				Properties: ws.QueryProps,
				Required:   ws.QueryRequired,
			}
		}
		if len(ws.HeaderProps) > 0 {
			channel.Bindings.WS.Headers = &spec.Schema{
				Type:       "object",
				Properties: ws.HeaderProps,
			}
		}
	}

	if len(channel.Parameters) == 0 {
		channel.Parameters = nil
	}

	p.spec.Channels[key] = channel
	return nil
}

func (p *Parser) registerMessages(channelKey string, file *ast.File, b *operationBuilder) error {
	allMessages := make([]messageEntry, 0, len(b.Messages)+len(b.ReplyMessages))
	allMessages = append(allMessages, b.Messages...)
	allMessages = append(allMessages, b.ReplyMessages...)
	for _, msg := range allMessages {
		if _, exists := p.spec.Components.Messages[msg.Name]; exists {
			continue
		}
		schemaName := msg.PayloadType
		if idx := strings.LastIndex(schemaName, "."); idx >= 0 {
			schemaName = schemaName[idx+1:]
		}
		p.spec.Components.Messages[msg.Name] = &spec.Message{
			Name:    msg.Name,
			Payload: spec.NewSchemaRef(spec.ComponentSchemaRef(schemaName)),
		}
		p.referencedTypes = append(p.referencedTypes, typeReference{
			typeName: msg.PayloadType,
			file:     file,
		})
	}
	for _, msg := range b.OneOfMessages {
		if _, exists := p.spec.Components.Messages[msg.Name]; exists {
			continue
		}
		var refs []*spec.SchemaRef
		for _, pt := range msg.PayloadTypes {
			schemaName := pt
			if idx := strings.LastIndex(schemaName, "."); idx >= 0 {
				schemaName = schemaName[idx+1:]
			}
			refs = append(refs, spec.NewSchemaRef(spec.ComponentSchemaRef(schemaName)))
			p.referencedTypes = append(p.referencedTypes, typeReference{
				typeName: pt,
				file:     file,
			})
		}
		p.spec.Components.Messages[msg.Name] = &spec.Message{
			Name: msg.Name,
			Payload: spec.NewInlineSchema(&spec.Schema{
				OneOf:         refs,
				Discriminator: msg.Discriminator,
			}),
		}
	}
	return nil
}

func (p *Parser) registerOperation(channelKey string, b *operationBuilder) error {
	if _, exists := p.operationIDs[b.OperationID]; exists {
		return fmt.Errorf("operation ID %q: %w", b.OperationID, ErrDuplicateOperationID)
	}
	p.operationIDs[b.OperationID] = true

	op := spec.Operation{
		Action:      b.Action,
		Channel:     spec.NewRef(spec.ChannelRef(channelKey)),
		Summary:     b.Summary,
		Description: b.Description,
		Tags:        b.Tags,
		Security:    b.Security,
	}

	for _, msg := range b.Messages {
		op.Messages = append(op.Messages, spec.NewRef(spec.ChannelMessageRef(channelKey, msg.Name)))
	}
	for _, msg := range b.OneOfMessages {
		op.Messages = append(op.Messages, spec.NewRef(spec.ChannelMessageRef(channelKey, msg.Name)))
	}

	if b.HasReply && len(b.ReplyMessages) > 0 {
		replyChannelKey := channelKey
		if b.ReplyChannelAddr != "" {
			replyChannelKey = addressToKey(b.ReplyChannelAddr)
		}

		reply := &spec.OperationReply{
			Channel: &spec.Reference{Ref: spec.ChannelRef(replyChannelKey)},
		}
		for _, msg := range b.ReplyMessages {
			reply.Messages = append(reply.Messages, spec.NewRef(spec.ChannelMessageRef(replyChannelKey, msg.Name)))
		}
		op.Reply = reply
	}

	p.spec.Operations[b.OperationID] = op
	return nil
}

func addressToKey(address string) string {
	address = strings.Trim(address, "/")

	var parts []string
	for _, segment := range strings.Split(address, "/") {
		segment = strings.TrimPrefix(segment, "{")
		segment = strings.TrimSuffix(segment, "}")
		if segment != "" {
			parts = append(parts, segment)
		}
	}

	if len(parts) == 0 {
		return "root"
	}

	result := parts[0]
	for _, p := range parts[1:] {
		if len(p) > 0 {
			result += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return result
}

func parseTags(raw string) []spec.Tag {
	parts := strings.Split(raw, ",")
	tags := make([]spec.Tag, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tags = append(tags, spec.Tag{Name: p})
		}
	}
	return tags
}

func mapSimpleType(t string) (string, bool) {
	switch strings.ToLower(t) {
	case "string":
		return "string", true
	case "int", "integer", "int32", "int64":
		return "integer", true
	case "float", "float32", "float64", "number":
		return "number", true
	case "bool", "boolean":
		return "boolean", true
	default:
		return "string", false
	}
}
