package msgpack

import (
	"bytes"
	"io"
	"math"
	"unicode/utf8"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/ctystrings"
)

type unknownType struct{}

var unknownVal = unknownType{}

// unknownValBytes is the raw bytes of the msgpack fixext1 value we
// write to represent a totally unknown value. It's an extension value of
// type zero whose value is irrelevant. Since it's irrelevant, we
// set it to a single byte whose value is also zero, since that's
// the most compact possible representation.
//
// The representation of a refined unknown value is different. See
// marshalUnknownValue for more details.
var unknownValBytes = []byte{0xd4, 0, 0}

func (uv unknownType) MarshalMsgpack() ([]byte, error) {
	return unknownValBytes, nil
}

const unknownWithRefinementsExt = 0x0c

type unknownValRefinementKey int64

const unknownValNullness unknownValRefinementKey = 1
const unknownValStringPrefix unknownValRefinementKey = 2
const unknownValNumberMin unknownValRefinementKey = 3
const unknownValNumberMax unknownValRefinementKey = 4
const unknownValLengthMin unknownValRefinementKey = 5
const unknownValLengthMax unknownValRefinementKey = 6

func marshalUnknownValue(rng cty.ValueRange, path cty.Path, enc *msgpack.Encoder) error {
	if rng.TypeConstraint() == cty.DynamicPseudoType {
		// cty.DynamicVal can never have refinements
		err := enc.Encode(unknownVal)
		if err != nil {
			return path.NewError(err)
		}
		return nil
	}

	var refnBuf bytes.Buffer
	refnEnc := msgpack.NewEncoder(&refnBuf)
	mapLen := 0

	if rng.DefinitelyNotNull() {
		mapLen++
		refnEnc.EncodeInt(int64(unknownValNullness))
		refnEnc.EncodeBool(false)
	}
	switch {
	case rng.TypeConstraint() == cty.Number:
		lower, lowerInc := rng.NumberLowerBound()
		upper, upperInc := rng.NumberUpperBound()
		boundTy := cty.Tuple([]cty.Type{cty.Number, cty.Bool})
		if lower.IsKnown() && lower != cty.NegativeInfinity {
			mapLen++
			refnEnc.EncodeInt(int64(unknownValNumberMin))
			marshal(
				cty.TupleVal([]cty.Value{lower, cty.BoolVal(lowerInc)}),
				boundTy,
				nil,
				refnEnc,
			)
		}
		if upper.IsKnown() && upper != cty.PositiveInfinity {
			mapLen++
			refnEnc.EncodeInt(int64(unknownValNumberMax))
			marshal(
				cty.TupleVal([]cty.Value{upper, cty.BoolVal(upperInc)}),
				boundTy,
				nil,
				refnEnc,
			)
		}
	case rng.TypeConstraint() == cty.String:
		if prefix := rng.StringPrefix(); prefix != "" {
			// To ensure the total size of the refinements blob does not exceed
			// the limit set by our decoder, truncate the prefix string.
			// We could allow up to 1018 bytes here if we assume that this
			// refinement will only ever be combined with NotNull(), but there
			// is no need for such long prefix refinements at the moment.
			maxPrefixLength := 256
			if len(prefix) > maxPrefixLength {
				prefix = prefix[:maxPrefixLength-1]
				prefix = ctystrings.SafeKnownPrefix(prefix)
			}
			mapLen++
			refnEnc.EncodeInt(int64(unknownValStringPrefix))
			refnEnc.EncodeString(prefix)
		}
	case rng.TypeConstraint().IsCollectionType():
		lower := rng.LengthLowerBound()
		upper := rng.LengthUpperBound()
		if lower != 0 {
			mapLen++
			refnEnc.EncodeInt(int64(unknownValLengthMin))
			refnEnc.EncodeInt(int64(lower))
		}
		if upper != math.MaxInt {
			mapLen++
			refnEnc.EncodeInt(int64(unknownValLengthMax))
			refnEnc.EncodeInt(int64(upper))
		}
	}

	if mapLen == 0 {
		// No refinements to encode, so we'll use the old compact representation.
		err := enc.Encode(unknownVal)
		if err != nil {
			return path.NewError(err)
		}
		return nil
	}

	// If we have at least one refinement to encode then we'll use the new
	// representation of unknown values where refinement information is in the
	// extension payload.
	var lenBuf bytes.Buffer
	lenEnc := msgpack.NewEncoder(&lenBuf)
	lenEnc.EncodeMapLen(mapLen)

	err := enc.EncodeExtHeader(unknownWithRefinementsExt, lenBuf.Len()+refnBuf.Len())
	if err != nil {
		return path.NewErrorf("failed to write unknown value: %s", err)
	}
	_, err = enc.Writer().Write(lenBuf.Bytes())
	if err != nil {
		return path.NewErrorf("failed to write unknown value: %s", err)
	}
	_, err = enc.Writer().Write(refnBuf.Bytes())
	if err != nil {
		return path.NewErrorf("failed to write unknown value: %s", err)
	}
	return nil
}

func unmarshalUnknownValue(dec *msgpack.Decoder, ty cty.Type, path cty.Path) (cty.Value, error) {
	// The next item in the stream should be a msgpack extension value,
	// which might be zero-length for a totally unknown value, or it might
	// contain a mapping describing some type-specific refinements.
	typeCode, extLen, err := dec.DecodeExtHeader()
	if err != nil {
		return cty.DynamicVal, path.NewErrorf("extension code is required for unknown value")
	}

	if extLen <= 1 {
		// Zero-length or one-length extension represents an unknown value with
		// no refinements. (msgpack's serialization of a zero-length extension
		// is one byte longer than a one-byte extension, so the encoder uses
		// one nul byte as its "totally unknown" encoding.

		if extLen > 0 {
			// We need to skip the body, then.
			body := make([]byte, extLen)
			_, err = io.ReadAtLeast(dec.Buffered(), body, len(body))
			if err != nil {
				return cty.DynamicVal, path.NewErrorf("failed to read msgpack extension body: %s", err)
			}
		}
		return cty.UnknownVal(ty), nil
	}

	if typeCode != unknownWithRefinementsExt {
		// If there's a non-zero length then we require a specific type code
		// as an additional signal that the body is intended to be a refinement map.
		return cty.DynamicVal, path.NewErrorf("unsupported extension type 0x%02x with len %d", typeCode, extLen)
	}

	if extLen > 1024 {
		// A refinement description greater than 1 kiB is unreasonable and
		// might be an abusive attempt to allocate large amounts of memory
		// in a system consuming this input.
		return cty.DynamicVal, path.NewErrorf("oversize unknown value refinement")
	}

	// If we get here then typeCode == 0xc and we have a non-zero length.
	// We expect to find a msgpack-encoded map in the payload which describes
	// any refinements to add to the result.
	body := make([]byte, extLen)
	_, err = io.ReadAtLeast(dec.Buffered(), body, len(body))
	if err != nil {
		return cty.DynamicVal, path.NewErrorf("failed to read msgpack extension body: %s", err)
	}

	rfnDec := msgpack.NewDecoder(bytes.NewReader(body))
	entryCount, err := rfnDec.DecodeMapLen()
	if err != nil {
		return cty.DynamicVal, path.NewErrorf("failed to decode msgpack extension body: not a map")
	}

	if ty == cty.DynamicPseudoType {
		// We'll silently ignore all refinements for DynamicPseudoType for now,
		// since we know that's invalid today but we might find a way to
		// support it in the future and if so will want to introduce that
		// in a backward-compatible way.
		return cty.UnknownVal(ty), nil
	}

	builder := cty.UnknownVal(ty).Refine()
	for i := 0; i < entryCount; i++ {
		// Our refinement encoding format uses compact msgpack primitives to
		// minimize the encoding size of refinements, which could otherwise
		// add up to be quite large for a payload containing lots of unknown
		// values. The keys are small integers to fit in the positive fixint
		// encoding scheme. The values are encoded differently depending on
		// the key but also aim for compactness.
		// The smallest possible non-empty refinement map is three bytes:
		// - one byte to encode that it's a one-element map
		// - one byte to encode the key
		// - at least one byte to encode the value associated with that key
		// Encoders should avoid encoding zero-length maps and prefer to
		// leave the payload zero bytes long in that case.

		keyCode, err := rfnDec.DecodeInt64()
		if err != nil {
			return cty.DynamicVal, path.NewErrorf("failed to decode msgpack extension body: non-integer key in map")
		}

		// Exactly which keys are supported depends on the destination type.
		// We'll reject keys that we know can't possibly apply to the given
		// type, but we'll ignore keys we haven't seen before to allow for
		// future expansion of the possible refinements.
		// These keys all have intentionally-short names
		switch keyCode := unknownValRefinementKey(keyCode); keyCode {
		case unknownValNullness:
			isNull, err := rfnDec.DecodeBool()
			if err != nil {
				return cty.DynamicVal, path.NewErrorf("failed to decode msgpack extension body: null refinement is not boolean")
			}
			// The presence of this key means we're refining the null-ness one
			// way or another. If nullness is unknown then this key should not
			// be present at all.
			if isNull {
				// it'd be weird to actually serialize a refinement like
				// this because trying to apply this refinement in the first
				// place should've collapsed into a known null value. But we'll
				// allow it anyway just for complete encoding of the current
				// refinement model.
				builder = builder.Null()
			} else {
				builder = builder.NotNull()
			}
		case unknownValStringPrefix:
			if ty != cty.String {
				return cty.DynamicVal, path.NewErrorf("failed to decode msgpack extension body: string prefix refinement for non-string type")
			}
			prefixStr, err := rfnDec.DecodeString()
			if err != nil {
				return cty.DynamicVal, path.NewErrorf("failed to decode msgpack extension body: string prefix refinement is not string")
			}
			if !utf8.ValidString(prefixStr) {
				return cty.DynamicVal, path.NewErrorf("failed to decode msgpack extension body: string prefix refinement is not valid UTF-8")
			}
			// We assume that the original creator of this value already took
			// care of making sure the prefix is safe, so we don't need to
			// constrain it any further.
			builder = builder.StringPrefixFull(prefixStr)
		case unknownValLengthMin, unknownValLengthMax:
			if !ty.IsCollectionType() {
				return cty.DynamicVal, path.NewErrorf("failed to decode msgpack extension body: length lower bound refinement for non-collection type")
			}

			bound, err := rfnDec.DecodeInt()
			if err != nil {
				return cty.DynamicVal, path.NewErrorf("failed to decode msgpack extension body: length bound refinement must be integer or [integer, bool] array")
			}
			switch keyCode {
			case unknownValLengthMin:
				builder = builder.CollectionLengthLowerBound(bound)
			case unknownValLengthMax:
				builder = builder.CollectionLengthUpperBound(bound)
			default:
				panic("unsupported keyCode") // should not get here
			}
		case unknownValNumberMin, unknownValNumberMax:
			if ty != cty.Number {
				return cty.DynamicVal, path.NewErrorf("failed to decode msgpack extension body: numeric bound refinement for non-number type")
			}
			// We want to support all of the same various number encodings we
			// support for normal numbers, so here we'll cheat a bit and decode
			// using our own value unmarshal function.
			rawBound, err := unmarshal(rfnDec, cty.Tuple([]cty.Type{cty.Number, cty.Bool}), nil)
			if err != nil || rawBound.IsNull() || !rawBound.IsKnown() {
				return cty.DynamicVal, path.NewErrorf("failed to decode msgpack extension body: length bound refinement must be [number, bool] array")
			}
			boundVal := rawBound.Index(cty.Zero)
			isIncVal := rawBound.Index(cty.NumberIntVal(1))
			if boundVal.Type() != cty.Number || !boundVal.IsKnown() || boundVal.IsNull() {
				return cty.DynamicVal, path.NewErrorf("failed to decode msgpack extension body: length bound refinement must be [number, bool] array")
			}
			if isIncVal.Type() != cty.Bool || !isIncVal.IsKnown() || isIncVal.IsNull() {
				return cty.DynamicVal, path.NewErrorf("failed to decode msgpack extension body: length bound refinement must be [number, bool] array")
			}
			isInc := isIncVal.True()
			switch keyCode {
			case unknownValNumberMin:
				builder = builder.NumberRangeLowerBound(boundVal, isInc)
			case unknownValNumberMax:
				builder = builder.NumberRangeUpperBound(boundVal, isInc)
			default:
				panic("unsupported keyCode") // should not get here
			}
		}
	}

	// NOTE: We intentionally ignore any trailing bytes after the extension
	// map in case we want to pack something else in there later or in case
	// a future version wants to use padding to optimize storage. Current
	// encoders should not add any extra content there, though.

	return builder.NewValue(), nil
}
