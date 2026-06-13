package main

// Method is a single MTProto RPC method exposed by tg.Client.
type Method struct {
	GoName     string  // e.g. MessagesSendMessage
	TLName     string  // e.g. messages.sendMessage
	Namespace  string  // e.g. messages ("general" when the method has no namespace)
	Short      string  // e.g. sendMessage (TL name without the namespace)
	Slug       string  // e.g. send-message (kebab-case of Short)
	Hash       string  // e.g. fef48f62
	Summary    string  // one-paragraph description
	DocURL     string  // https://core.telegram.org/method/<TLName>
	Signature  string  // full Go signature as users call it
	Usage      string  // runnable "calling this method" snippet
	Params     []Param // request parameters
	ReturnType string  // Go return type (excluding error), e.g. UpdatesClass
	Errors     []ErrorRow
}

// Param is one field of a method request or a constructor.
type Param struct {
	Name        string
	GoType      string
	Optional    bool
	Description string
}

// ErrorRow is one documented error a method may return.
type ErrorRow struct {
	Code        string // e.g. 400
	Name        string // e.g. CHANNEL_INVALID
	Description string
}

// Constructor is a concrete TL type (a constructor of a class, or a bare type).
type Constructor struct {
	GoName     string // e.g. InputPeerUser
	TLName     string // e.g. inputPeerUser
	Namespace  string
	Short      string
	Slug       string
	Hash       string
	Summary    string
	DocURL     string // https://core.telegram.org/constructor/<TLName>
	Fields     []Param
	Implements *Ref // the class it constructs, or nil for bare types
}

// Type is a TL class: an interface implemented by one or more constructors.
type Type struct {
	GoName       string // e.g. InputPeerClass
	TLName       string // e.g. InputPeer (or auth.Authorization)
	Namespace    string
	Short        string
	Slug         string
	Summary      string
	DocURL       string // https://core.telegram.org/type/<TLName>
	Constructors []*Ref
}

// Ref is a cross-reference to another generated page.
type Ref struct {
	GoName    string
	TLName    string
	Namespace string
	Slug      string
}

// Namespace groups methods for the overview.
type Namespace struct {
	Name        string
	Description string
	Methods     []*Method
}

// Result is the full parsed model.
type Result struct {
	Methods      []*Method
	Types        []*Type
	Constructors []*Constructor
}
