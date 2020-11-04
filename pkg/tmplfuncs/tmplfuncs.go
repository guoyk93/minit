package tmplfuncs

import (
	"errors"
	"net"
	"os"
	"os/user"
	"strconv"
	"strings"
)

var Funcs = map[string]interface{}{
	// built-in functions
	"netResolveIPAddr":    net.ResolveIPAddr,
	"osHostname":          os.Hostname,
	"osUserCacheDir":      os.UserCacheDir,
	"osUserConfigDir":     os.UserConfigDir,
	"osUserHomeDir":       os.UserHomeDir,
	"osGetegid":           os.Getegid,
	"osGetenv":            os.Getenv,
	"osGeteuid":           os.Geteuid,
	"osGetgid":            os.Getgid,
	"osGetgroups":         os.Getgroups,
	"osGetpagesize":       os.Getpagesize,
	"osGetpid":            os.Getpid,
	"osGetppid":           os.Getppid,
	"osGetuid":            os.Getuid,
	"osGetwd":             os.Getwd,
	"osTempDir":           os.TempDir,
	"osUserLookupGroup":   user.LookupGroup,
	"osUserLookupGroupId": user.LookupGroupId,
	"osUserCurrent":       user.Current,
	"osUserLookup":        user.Lookup,
	"osUserLookupId":      user.LookupId,
	"stringsContains":     strings.Contains,
	"stringsFields":       strings.Fields,
	"stringsIndex":        strings.Index,
	"stringsLastIndex":    strings.LastIndex,
	"stringsHasPrefix":    strings.HasPrefix,
	"stringsHasSuffix":    strings.HasSuffix,
	"stringsRepeat":       strings.Repeat,
	"stringsReplaceAll":   strings.ReplaceAll,
	"stringsSplit":        strings.Split,
	"stringsToLower":      strings.ToLower,
	"stringsToUpper":      strings.ToUpper,
	"stringsTrimPrefix":   strings.TrimPrefix,
	"stringsTrimSpace":    strings.TrimSpace,
	"stringsTrimSuffix":   strings.TrimSuffix,
	"strconvQuote":        strconv.Quote,
	"strconvUnquote":      strconv.Unquote,
	"strconvParseBool":    strconv.ParseBool,
	"strconvParseInt":     strconv.ParseInt,
	"strconvParseUint":    strconv.ParseUint,
	"strconvParseFloat":   strconv.ParseFloat,
	"strconvFormatBool":   strconv.FormatBool,
	"strconvFormatInt":    strconv.FormatInt,
	"strconvFormatUint":   strconv.FormatUint,
	"strconvFormatFloat":  strconv.FormatFloat,
	"strconvAoti":         strconv.Atoi,
	"strconvItoa":         strconv.Itoa,

	// extra functions
	"k8sStatefulSetID": k8sStatefulSetID,
}

func k8sStatefulSetID() (id int64, err error) {
	var hostname string
	if hostname = os.Getenv("HOSTNAME"); hostname == "" {
		if hostname, err = os.Hostname(); err != nil {
			return
		}
	}
	splits := strings.Split(hostname, "-")
	if len(splits) < 2 {
		err = errors.New("invalid stateful-set hostname")
		return
	}
	id, err = strconv.ParseInt(splits[len(splits)-1], 10, 64)
	return
}
