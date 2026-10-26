package sharedtags

// JSON credits constants.
const (
	JActor       = "actor"
	JAuthor      = "author"
	JArtist      = "artist"
	JAlbumArtist = "album_artist"
	JChannel     = "channel"
	JComposer    = "composer"
	JCreator     = "creator"
	JDirector    = "director"
	JPerformer   = "performer"
	JProducer    = "producer"
	JPublisher   = "publisher"
	JStudio      = "studio"
	JUploader    = "uploader"
	JWriter      = "writer"
)

// JSON title and description constants.
const (
	// Titles.
	JFulltitle = "fulltitle"
	JTitle     = "title"

	// Subtitle (keep separate).
	JSubtitle = "subtitle"

	// Descriptions.
	JComment            = "comment"
	JDescription        = "description"
	JLongDesc           = "longdescription"
	JLongUnderscoreDesc = "long_description"
	JSummary            = "summary"
	JSynopsis           = "synopsis"
)

// JSON date/time constants.
const (
	JCreationTime        = "creation_time"           // Ends "Txx:xx:xxZ"
	JDate                = "date"                    // YYYY-MM-DD
	JOriginallyAvailable = "originally_available_at" // YYYY-MM-DD
	JReleaseDate         = "release_date"            // YYYY-MM-DD
	JUploadDate          = "upload_date"             // YYYY-MM-DD
	JYear                = "year"                    // YYYY
	JReleaseYear         = "release_year"
	JFormattedDate       = "formatted_date" // YYYY-MM-DD
	JDuration            = "duration"
	JDurationString      = "duration_string"
)

// JSON show constants.
const (
	JEpisodeID     = "episode_id"
	JEpisodeNumber = "episode_number"
	JEpisodeSort   = "episode_sort"
	JNetwork       = "network"
	JSeason        = "season"
	JSeasonNumber  = "season_number"
	JSeasonTitle   = "season_title"
	JSeries        = "series"
	JShow          = "show"
)

// JSON other constants.
const (
	JAlbum       = "album"
	JCategories  = "categories"
	JCopyright   = "copyright"
	JDiscNumber  = "disc_number"
	JGenre       = "genre"
	JGenres      = "genres"
	JHDVideo     = "hd_video"
	JHeight      = "height"
	JKeywords    = "keywords"
	JLanguage    = "language"
	JLicense     = "license"
	JTags        = "tags"
	JTrackNumber = "track_number"
)

// JSON web constants.
const (
	JChannelURL     = "channel_url"
	JDirectVideoURL = "direct_video_url"
	JDomain         = "domain"
	JReferer        = "referer"
	JUploaderURL    = "uploader_url"
	JURL            = "url"
	JWebpageDomain  = "webpage_url_domain"
	JWebpageURL     = "webpage_url"
	JThumbnail      = "thumbnail"
	JThumbnails     = "thumbnails"
	JPreference     = "preference"
)
