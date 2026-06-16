package resources

// CodecRegistration pairs a codec with a zero value of its state struct, so the
// global contract test can validate the mapping without reaching into each
// resource's (unexported) codec var.
type CodecRegistration struct {
	Name  string
	Codec Codec
	State any
}

var codecRegistry []CodecRegistration

// RegisterCodec records a codec and its state type for the global contract test.
// Call it from an init() in the resource package, next to the codec var:
//
//	func init() { resources.RegisterCodec("mysql", mysqlFeaturesCodec, &MySQL{}) }
//
// A single test in pkg/registry (which imports every resource package) then runs
// Validate on all registrations, so a codec/struct mismatch is caught at CI time.
func RegisterCodec(name string, c Codec, state any) {
	codecRegistry = append(codecRegistry, CodecRegistration{Name: name, Codec: c, State: state})
}

// RegisteredCodecs returns every codec registered via RegisterCodec.
func RegisteredCodecs() []CodecRegistration { return codecRegistry }
