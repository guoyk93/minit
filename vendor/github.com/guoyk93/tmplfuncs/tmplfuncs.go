package tmplfuncs

import (
	"net"
	"strings"
)

var Funcs = map[string]interface{}{
	"stringsContains":   strings.Contains,
	"stringsFields":     strings.Fields,
	"stringsIndex":      strings.Index,
	"stringsLastIndex":  strings.LastIndex,
	"stringsHasPrefix":  strings.HasPrefix,
	"stringsHasSuffix":  strings.HasSuffix,
	"stringsRepeat":     strings.Repeat,
	"stringsReplaceAll": strings.ReplaceAll,
	"stringsSplit":      strings.Split,
	"stringsToLower":    strings.ToLower,
	"stringsToUpper":    strings.ToUpper,
	"stringsTrimPrefix": strings.TrimPrefix,
	"stringsTrimSpace":  strings.TrimSpace,
	"stringsTrimSuffix": strings.TrimSuffix,
	"netResolveIPAddr":  net.ResolveIPAddr,
}
