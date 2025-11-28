// Package sharedtemplates holds templating elements used across Tubarr and Metarr.
package sharedtemplates

// Template tags for channel elements.
const (
	ChannelID   = "channel_id"
	ChannelName = "channel_name"
	ChannelURL  = "channel_url"
)

// Template tags for individual video elements.
const (
	MetVideoTitle = "video_title"
	MetVideoURL   = "video_url"
)

// Template tags for date metadata.
const (
	MetDay   = "day"
	MetMonth = "month"
	MetYear  = "year"
)

// Template tags for cast metadata.
const (
	MetAuthor   = "author"
	MetDirector = "director"
)

// Template tags for URL metadata.
const (
	MetDomain = "domain"
)

// AllTemplatesMap contains the different template tags available across both Tubarr and Metarr.
var AllTemplatesMap = map[string]struct{}{
	ChannelID:     {},
	ChannelName:   {},
	ChannelURL:    {},
	MetVideoTitle: {},
	MetVideoURL:   {},
	MetDay:        {},
	MetMonth:      {},
	MetYear:       {},
	MetAuthor:     {},
	MetDirector:   {},
	MetDomain:     {},
}

// TubarrTemplateTags contains templating tags which are fillable by Tubarr channels.
var TubarrTemplateTags = map[string]struct{}{
	ChannelID:   {},
	ChannelName: {},
	ChannelURL:  {},
}

// MetarrTemplateTags contains templating tags which are fillable by Metarr.
var MetarrTemplateTags = map[string]struct{}{
	MetVideoTitle: {},
	MetVideoURL:   {},
	MetYear:       {},
	MetMonth:      {},
	MetDay:        {},
	MetAuthor:     {},
	MetDirector:   {},
	MetDomain:     {},
}
