package api

const (
	PrimaryVersion = "0"
	SubVersion     = "1"
	DevVersion     = "46"
)

const (
	BaseVersion = "v" + PrimaryVersion + "." + SubVersion + "." + DevVersion
	// GetNowTimeStringWithHyphens
	DateTime = "2024-11-25 11:22:55"
)

func apiVersionWithMaker(maker string) string {
	if maker == "" {
		return BaseVersion + "-" + DateTime
	} else {
		return BaseVersion + "-" + maker + "-" + DateTime
	}
}

// GetApiVersion
// @Description: Get api version
func GetApiVersion() string {
	return apiVersionWithMaker("")
}
