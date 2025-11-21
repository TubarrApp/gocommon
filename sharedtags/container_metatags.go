// Package sharedtags holds tag information for things like JSON fields, and container metatags.
package sharedtags

// ASF container metatag keys.
const (
	ASFArtist              = "WM/AlbumArtist"
	ASFComposer            = "WM/Composer"
	ASFDirector            = "WM/Director"
	ASFEncodingTime        = "WM/EncodingTime"
	ASFProducer            = "WM/Producer"
	ASFSubtitle            = "WM/SubTitle"
	ASFSubTitleDescription = "WM/SubTitleDescription"
	ASFTitle               = "Title"
	ASFYear                = "WM/Year"
)

// AVI container metatag keys.
const (
	AVIComments    = "COMM"
	AVIArtist      = "IART"
	AVIComment     = "ICMT"
	AVIDateCreated = "ICRD"
	AVIProducer    = "IENG"
	AVITitle       = "INAM"
	AVISubject     = "ISBJ"
	AVIYear        = "YEAR"
)

// FLV container metatags keys.
const (
	FLVCreationDate = "creationdate"
)

// ISOBMFF container metatag keys.
const (
	ISOArtist       = "artist"
	ISOComment      = "comment"
	ISOComposer     = "composer"
	ISOCreationTime = "creation_time"
	ISODate         = "date"
	ISODescription  = "description"
	ISOSynopsis     = "synopsis"
	ISOTitle        = "title"
)

// Matroska container metatag keys.
const (
	MatroskaArtist        = "ARTIST"
	MatroskaComposer      = "COMPOSER"
	MatroskaDateEncoded   = "DATE_ENCODED"
	MatroskaDateReleased  = "DATE_RELEASED"
	MatroskaDescription   = "DESCRIPTION"
	MatroskaDirector      = "DIRECTOR"
	MatroskaLeadPerformer = "LEAD_PERFORMER"
	MatroskaPerformer     = "PERFORMER"
	MatroskaProducer      = "PRODUCER"
	MatroskaSubject       = "SUBJECT"
	MatroskaSummary       = "SUMMARY"
	MatroskaSynopsis      = "SYNOPSIS"
	MatroskaTitle         = "TITLE"
)

// MPEG-TS container metatag keys.
const (
	TSServiceName     = "service_name"
	TSServiceProvider = "service_provider"
)

// Ogg container metatag keys.
const (
	OggArtist      = "ARTIST"
	OggComposer    = "COMPOSER"
	OggDate        = "DATE"
	OggDescription = "DESCRIPTION"
	OggPerformer   = "PERFORMER"
	OggSummary     = "SUMMARY"
	OggTitle       = "TITLE"
)

// RealMedia container metatag keys.
const (
	RMAuthor  = "Author"
	RMComment = "Comment"
	RMTitle   = "Title"
)
