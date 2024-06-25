package format

import (
	"errors"
	"fmt"
	"strings"

	"github.com/protobom/protobom/pkg/formats"
)

func DefaultSPDXJSONVersion() formats.Format {
	return formats.SPDX23JSON
}

func DefaultSPDXTVVersion() formats.Format {
	return formats.SPDX23TV
}

func DefaultCycloneDXVersion() formats.Format {
	return formats.CDX15JSON
}

func DefaultEncoding() string {
	return formats.JSON
}

func DefaultFormatString() string {
	return "cyclonedx-1.5"
}

func JSONFormatMap() map[string]formats.Format {
	return map[string]formats.Format{
		"spdx":     formats.SPDXFORMAT,
		"spdx-2.2": formats.SPDX22JSON,
		"spdx-2.3": formats.SPDX23JSON,

		"cyclonedx":     formats.CDXFORMAT,
		"cyclonedx-1.0": formats.CDX10JSON,
		"cyclonedx-1.1": formats.CDX11JSON,
		"cyclonedx-1.2": formats.CDX12JSON,
		"cyclonedx-1.3": formats.CDX13JSON,
		"cyclonedx-1.4": formats.CDX14JSON,
		"cyclonedx-1.5": formats.CDX15JSON,
	}
}

func TVFormatMap() map[string]formats.Format {
	return map[string]formats.Format{
		"spdx":     formats.SPDXFORMAT,
		"spdx-2.2": formats.SPDX22TV,
		"spdx-2.3": formats.SPDX23TV,
	}
}

func EncodingMap() map[string]string {
	return map[string]string{
		"json": formats.JSON,
		"xml":  formats.XML,
		"text": formats.TEXT,
	}
}

func XMLFormatMap() map[string]formats.Format {
	return map[string]formats.Format{}
}

type Format struct {
	formats.Format
}

// Parse parses the format string into a formats.Format
func Parse(fs string, encoding string) (formats.Format, error) {
	if fs == "" {
		return formats.EmptyFormat, errors.New("no format specified")
	}
	s := strings.ToLower(fs)
	var fm map[string]formats.Format

	switch encoding {
	case formats.JSON:
		fm = JSONFormatMap()
	case formats.TEXT:
		fm = TVFormatMap()
	case formats.XML:
		fm = XMLFormatMap()
	default:
		return formats.EmptyFormat,
			fmt.Errorf("unknown encoding: %s", encoding)
	}
	switch f, ok := fm[s]; {
	case !ok:
		return formats.EmptyFormat, fmt.Errorf("unknown format: %s", fs)
	case f == formats.SPDXFORMAT && encoding == formats.JSON:
		return DefaultSPDXJSONVersion(), nil
	case f == formats.SPDXFORMAT && encoding == formats.TEXT:
		return DefaultSPDXTVVersion(), nil
	case f == formats.CDXFORMAT:
		return DefaultCycloneDXVersion(), nil
	default:
		return f, nil
	}
}
