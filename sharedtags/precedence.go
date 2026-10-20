package sharedtags

// TitleKeys are the JSON keys a video's title is taken from, most preferred first.
var TitleKeys = []string{JFulltitle, "full_title", JTitle}

// DescriptionKeys are the JSON keys a video's description is taken from, most preferred first.
var DescriptionKeys = []string{JLongDesc, JLongUnderscoreDesc, JDescription, JSummary, JSynopsis}

// DateKeys are the JSON keys a video's date is taken from, most preferred first. Creation times are a last resort, as
// they may be when the file was made rather than the video.
var DateKeys = []string{
	JReleaseDate, "releasedate", "released_on",
	JOriginallyAvailable, "originally_available", "originallyavailable",
	JDate,
	JUploadDate, "uploaddate", "uploaded_on",
	JCreationTime, "created_at",
}

// FirstString returns the value of the first key in keys with a non-empty string value in m, or "" if none has one.
func FirstString(m map[string]any, keys []string) string {
	for _, k := range keys {
		if s, ok := m[k].(string); ok && s != "" {
			return s
		}
	}
	return ""
}
