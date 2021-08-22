package uploadipa

const (
	xRequestID      = "x-request-id"
	xSessionDigest  = "x-session-digest"
	xSessionID      = "x-session-id"
	xSessionVersion = "x-session-version"

	cstAuthForSessioon   = "authenticateForSession"
	cstLookupForBundleID = "lookupSoftwareForBundleId"
	cstValidMetadata     = "validateMetadata"
	cstValidAssets       = "validateAssets"
	cstCleintChecksum    = "clientChecksumCompleted"
	cstCreateReservation = "createReservation"
	cstCommitRev         = "commitReservation"
	cstDone              = "uploadDoneWithArguments"

	cstSoftURL    = "https://contentdelivery.itunes.apple.com/WebObjects/MZLabelService.woa/json/MZITunesSoftwareService"
	cstProduceURL = "https://contentdelivery.itunes.apple.com/WebObjects/MZLabelService.woa/json/MZITunesProducerService"

	headKeyUserAgent    = "User-Agent"
	headValueUserAgent  = "iTMSTransporter/2.0.0"
	headKeyContentType  = "Content-Type"
	headValueContenType = "application/json"

	cstMetadata = `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://apple.com/itunes/importer" version="software5.4">
  <software_assets apple_id="APPLE_ID" bundle_short_version_string="BUNDLE_SHORT_VERSION" bundle_version="BUNDLE_VERSION" bundle_identifier="BUNDLE_IDENTIFIER" app_platform="ios">
    <asset type="bundle">
      <data_file>
        <size>FILE_SIZE</size>
        <file_name>FILE_NAME</file_name>
        <checksum type="md5">MD5</checksum>
      </data_file>
    </asset>
  </software_assets>
</package>`

	cstServerError = `Apple's web service error!
Unable to process createReservation request at this time due to a general error (1015).
`

)
