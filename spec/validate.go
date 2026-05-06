package spec

import (
	"fmt"
	"strings"
)

func (doc *AsyncAPI) ValidateBasic() []error {
	var errs []error

	if doc.AsyncAPI == "" {
		errs = append(errs, fmt.Errorf("asyncapi version is required"))
	}
	if doc.Info.Title == "" {
		errs = append(errs, fmt.Errorf("info.title is required"))
	}
	if doc.Info.Version == "" {
		errs = append(errs, fmt.Errorf("info.version is required"))
	}

	return errs
}

func (doc *AsyncAPI) Validate() []error {
	errs := doc.ValidateBasic()

	for name, ch := range doc.Channels {
		if ch.Address == "" {
			errs = append(errs, fmt.Errorf("channel %q: address is required", name))
		}
		for msgName, msgRef := range ch.Messages {
			if msgRef.Ref != "" {
				if err := doc.validateRef(msgRef.Ref, "messages"); err != nil {
					errs = append(errs, fmt.Errorf("channel %q message %q: %w", name, msgName, err))
				}
			}
		}
	}

	for name, op := range doc.Operations {
		if op.Action == "" {
			errs = append(errs, fmt.Errorf("operation %q: action is required", name))
		}
		if op.Channel.Ref != "" {
			channelName := refTarget(op.Channel.Ref)
			if _, ok := doc.Channels[channelName]; !ok {
				errs = append(errs, fmt.Errorf("operation %q: channel ref %q not found", name, op.Channel.Ref))
			}
		}
		for i, msgRef := range op.Messages {
			if err := doc.validateChannelMessageRef(msgRef.Ref); err != nil {
				errs = append(errs, fmt.Errorf("operation %q messages[%d]: %w", name, i, err))
			}
		}
		if op.Reply != nil {
			if op.Reply.Channel != nil && op.Reply.Channel.Ref != "" {
				replyChannelName := refTarget(op.Reply.Channel.Ref)
				if _, ok := doc.Channels[replyChannelName]; !ok {
					errs = append(errs, fmt.Errorf("operation %q reply channel ref %q not found", name, op.Reply.Channel.Ref))
				}
			}
			for i, msgRef := range op.Reply.Messages {
				if err := doc.validateChannelMessageRef(msgRef.Ref); err != nil {
					errs = append(errs, fmt.Errorf("operation %q reply.messages[%d]: %w", name, i, err))
				}
			}
		}
	}

	if doc.Components != nil {
		for name, msg := range doc.Components.Messages {
			if msg.Payload != nil && msg.Payload.Ref != "" {
				if err := doc.validateRef(msg.Payload.Ref, "schemas"); err != nil {
					errs = append(errs, fmt.Errorf("message %q payload: %w", name, err))
				}
			}
			if msg.Payload != nil && msg.Payload.Schema != nil {
				for i, ref := range msg.Payload.Schema.OneOf {
					if ref.Ref != "" {
						if err := doc.validateRef(ref.Ref, "schemas"); err != nil {
							errs = append(errs, fmt.Errorf("message %q payload.oneOf[%d]: %w", name, i, err))
						}
					}
				}
			}
		}
	}

	return errs
}

func (doc *AsyncAPI) validateRef(ref string, componentType string) error {
	prefix := "#/components/" + componentType + "/"
	if !strings.HasPrefix(ref, prefix) {
		return nil
	}
	name := strings.TrimPrefix(ref, prefix)
	if doc.Components == nil {
		return fmt.Errorf("$ref %q: components is nil", ref)
	}
	switch componentType {
	case "schemas":
		if _, ok := doc.Components.Schemas[name]; !ok {
			return fmt.Errorf("$ref %q: schema not found", ref)
		}
	case "messages":
		if _, ok := doc.Components.Messages[name]; !ok {
			return fmt.Errorf("$ref %q: message not found", ref)
		}
	}
	return nil
}

func (doc *AsyncAPI) validateChannelMessageRef(ref string) error {
	const prefix = "#/channels/"
	if !strings.HasPrefix(ref, prefix) {
		return nil
	}
	rest := strings.TrimPrefix(ref, prefix)
	parts := strings.SplitN(rest, "/messages/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("$ref %q: not a valid channel message reference", ref)
	}
	channelName, msgName := parts[0], parts[1]
	ch, ok := doc.Channels[channelName]
	if !ok {
		return fmt.Errorf("$ref %q: channel %q not found", ref, channelName)
	}
	if _, ok := ch.Messages[msgName]; !ok {
		return fmt.Errorf("$ref %q: message %q not found in channel %q", ref, msgName, channelName)
	}
	return nil
}

func refTarget(ref string) string {
	idx := strings.LastIndex(ref, "/")
	if idx < 0 {
		return ref
	}
	return ref[idx+1:]
}
