// Package sharedtags holds tag information for things like JSON fields, and container metatags.
package sharedtags

// ASF container metatag keys. FFmpeg writes any key, these are the standard Windows Media ones.
const (
	ASFAlbumArtist         = "WM/AlbumArtist"
	ASFAlbumTitle          = "WM/AlbumTitle"
	ASFAuthor              = "Author"
	ASFComposer            = "WM/Composer"
	ASFCopyright           = "Copyright"
	ASFDescription         = "Description"
	ASFDirector            = "WM/Director"
	ASFGenre               = "WM/Genre"
	ASFLanguage            = "WM/Language"
	ASFOriginalReleaseTime = "WM/OriginalReleaseTime"
	ASFPartOfSet           = "WM/PartOfSet"
	ASFProducer            = "WM/Producer"
	ASFPublisher           = "WM/Publisher"
	ASFSubtitle            = "WM/SubTitle"
	ASFSubTitleDescription = "WM/SubTitleDescription"
	ASFTitle               = "Title"
	ASFTrackNumber         = "WM/TrackNumber"
	ASFWriter              = "WM/Writer"
	ASFYear                = "WM/Year"
)

// AVI container metatag keys: every RIFF INFO tag FFmpeg writes.
const (
	AVIArchivalLocation = "IARL"
	AVIArtist           = "IART"
	AVIFirstLanguage    = "IAS1"
	AVISecondLanguage   = "IAS2"
	AVIThirdLanguage    = "IAS3"
	AVIFourthLanguage   = "IAS4"
	AVIFifthLanguage    = "IAS5"
	AVISixthLanguage    = "IAS6"
	AVISeventhLanguage  = "IAS7"
	AVIEighthLanguage   = "IAS8"
	AVINinthLanguage    = "IAS9"
	AVICommissioned     = "ICMS"
	AVIComment          = "ICMT"
	AVICopyright        = "ICOP"
	AVIDateCreated      = "ICRD"
	AVICropped          = "ICRP"
	AVIDimensions       = "IDIM"
	AVIDotsPerInch      = "IDPI"
	AVIEngineer         = "IENG"
	AVIGenre            = "IGNR"
	AVIKeywords         = "IKEY"
	AVILightness        = "ILGT"
	AVILanguage         = "ILNG"
	AVIMedium           = "IMED"
	AVITitle            = "INAM"
	AVIPaletteColors    = "IPLT"
	AVIProduct          = "IPRD"
	AVIPart             = "IPRT"
	AVITrackNumber      = "ITRK"
	AVISubject          = "ISBJ"
	AVISoftware         = "ISFT"
	AVISharpness        = "ISHP"
	AVITimeCode         = "ISMP"
	AVISource           = "ISRC"
	AVISourceForm       = "ISRF"
	AVITechnician       = "ITCH"
)

// FLV container metatag keys. FFmpeg writes any key.
const (
	FLVCreationDate = "creationdate"
)

// MP4, M4V and F4V container metatag keys: every iTunes tag FFmpeg writes.
const (
	MP4Album           = "album"
	MP4AlbumArtist     = "album_artist"
	MP4Artist          = "artist"
	MP4Comment         = "comment"
	MP4Compilation     = "compilation"
	MP4Composer        = "composer"
	MP4Copyright       = "copyright"
	MP4CreationTime    = "creation_time"
	MP4Date            = "date"
	MP4Description     = "description"
	MP4Disc            = "disc"
	MP4DiscSubtitle    = "disc_subtitle"
	MP4EncodingTool    = "encoding_tool"
	MP4EpisodeID       = "episode_id"
	MP4EpisodeSort     = "episode_sort"
	MP4GaplessPlayback = "gapless_playback"
	MP4Genre           = "genre"
	MP4Grouping        = "grouping"
	MP4HDVideo         = "hd_video"
	MP4Keywords        = "keywords"
	MP4Location        = "location"
	MP4Lyrics          = "lyrics"
	MP4MediaType       = "media_type"
	MP4Network         = "network"
	MP4SeasonNumber    = "season_number"
	MP4Show            = "show"
	MP4Synopsis        = "synopsis"
	MP4Tempo           = "tmpo"
	MP4Title           = "title"
	MP4Track           = "track"
)

// MOV container metatag keys: every QuickTime tag FFmpeg writes.
const (
	MOVAlbum        = "album"
	MOVArtist       = "artist"
	MOVAuthor       = "author"
	MOVComment      = "comment"
	MOVCopyright    = "copyright"
	MOVCreationTime = "creation_time"
	MOVDate         = "date"
	MOVEncoder      = "encoder"
	MOVGenre        = "genre"
	MOVKeywords     = "keywords"
	MOVLocation     = "location"
	MOVMake         = "make"
	MOVModel        = "model"
	MOVTitle        = "title"
	MOVXMP          = "xmp"
)

// 3GP and 3G2 container metatag keys: every 3GPP tag FFmpeg writes.
const (
	ThreeGPAlbum        = "album"
	ThreeGPArtist       = "artist"
	ThreeGPAuthor       = "author"
	ThreeGPComment      = "comment"
	ThreeGPCopyright    = "copyright"
	ThreeGPCreationTime = "creation_time"
	ThreeGPDate         = "date" // Written as the year only.
	ThreeGPGenre        = "genre"
	ThreeGPLocation     = "location"
	ThreeGPTitle        = "title"
)

// Matroska container metatag keys. FFmpeg writes any key, in capitals.
const (
	MatroskaActor            = "ACTOR"
	MatroskaArtist           = "ARTIST"
	MatroskaComment          = "COMMENT"
	MatroskaComposer         = "COMPOSER"
	MatroskaCopyright        = "COPYRIGHT"
	MatroskaDateEncoded      = "DATE_ENCODED"
	MatroskaDateReleased     = "DATE_RELEASED"
	MatroskaDescription      = "DESCRIPTION"
	MatroskaDirector         = "DIRECTOR"
	MatroskaGenre            = "GENRE"
	MatroskaKeywords         = "KEYWORDS"
	MatroskaLeadPerformer    = "LEAD_PERFORMER"
	MatroskaPartNumber       = "PART_NUMBER"
	MatroskaProducer         = "PRODUCER"
	MatroskaProductionStudio = "PRODUCTION_STUDIO"
	MatroskaPublisher        = "PUBLISHER"
	MatroskaSubtitle         = "SUBTITLE"
	MatroskaSummary          = "SUMMARY"
	MatroskaSynopsis         = "SYNOPSIS"
	MatroskaTitle            = "TITLE"
	MatroskaURL              = "URL"
	MatroskaWrittenBy        = "WRITTEN_BY"
)

// MPEG-TS container metatag keys: every tag FFmpeg writes.
const (
	TSServiceName     = "service_name"
	TSServiceProvider = "service_provider"
)

// Ogg container metatag keys. FFmpeg writes any key, these are the standard Vorbis comment ones.
const (
	OggAlbum        = "ALBUM"
	OggAlbumArtist  = "ALBUMARTIST"
	OggArtist       = "ARTIST"
	OggComposer     = "COMPOSER"
	OggContact      = "CONTACT"
	OggCopyright    = "COPYRIGHT"
	OggDate         = "DATE"
	OggDescription  = "DESCRIPTION"
	OggDiscNumber   = "DISCNUMBER"
	OggGenre        = "GENRE"
	OggISRC         = "ISRC"
	OggLicense      = "LICENSE"
	OggLocation     = "LOCATION"
	OggOrganization = "ORGANIZATION"
	OggPerformer    = "PERFORMER"
	OggSummary      = "SUMMARY"
	OggTitle        = "TITLE"
	OggTrackNumber  = "TRACKNUMBER"
	OggVersion      = "VERSION"
)

// RealMedia container metatag keys: every tag FFmpeg writes.
const (
	RMAuthor    = "Author"
	RMComment   = "Comment"
	RMCopyright = "Copyright"
	RMTitle     = "Title"
)

// FFprobe's names for container tags FFmpeg converts on reading, e.g. AVI's "IART" is reported as "artist".
const (
	FFProbeAlbum       = "album"
	FFProbeAlbumArtist = "album_artist"
	FFProbeArtist      = "artist"
	FFProbeComment     = "comment"
	FFProbeComposer    = "composer"
	FFProbeCopyright   = "copyright"
	FFProbeDate        = "date"
	FFProbeDisc        = "disc"
	FFProbeGenre       = "genre"
	FFProbeLanguage    = "language"
	FFProbePerformer   = "performer"
	FFProbeProduct     = "product"
	FFProbePublisher   = "publisher"
	FFProbeSubject     = "subject"
	FFProbeTitle       = "title"
	FFProbeTrack       = "track"
)
