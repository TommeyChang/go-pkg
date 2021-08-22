package uploadipa

import (
	"errors"
	"fmt"
	"time"
)

func authForSession(bus *dataBus) (bool, string, error) {

	if rp, err := makeProducerServiceReq(bus, cstAuthForSessioon, map[string]interface{}{
		"Username": bus.getString(dbsUsername),
		"Password": bus.getString(dbsPassword),
	}); err != nil {
		return false, "", err
	} else {
		sessionID, okID := (*rp)["SessionId"]
		sessionSec, okS := (*rp)["SharedSecret"]
		if okID && okS {
			bus.set(dbsSessionID, sessionID)
			bus.set(dbsSharedSecret, sessionSec)
			return true, "", nil
		} else {
			info := (*rp)["Errors"].([]interface{})[0]
			return false, fmt.Sprintln(info), nil
		}
	}

}

func lookupSoftwareForBundleID(bus *dataBus) (bool, string, error) {
	reqParams := map[string]interface{}{
		"Application":         "altlook",
		"ApplicationBundleId": "com.apple.itunes.altool",
		"BundleId":            bus.getString(dbsBundleid),
		"Version":             "4.0.1 (1182)",
	}
	if rp, err := makeSoftwareServiceReq(bus, cstLookupForBundleID, reqParams); err != nil {
		return false, "", err
	} else {
		_, okS := (*rp)["Success"]
		attr, okA := (*rp)["Attributes"]
		if okS && okA {
			attrMap := (attr.([]interface{})[0]).(map[string]interface{})
			bus.set(dbsAppleID, attrMap["AppleID"])
			bus.set(dbsAppName, attrMap["Application"])
			bus.set(dbsIconURL, attrMap["IconURL"])
			return true, "", nil

		} else {
			info := (*rp)["Errors"].([]interface{})[0]
			return false, fmt.Sprintln(info), nil
		}
	}

}

func validMetadata(bus *dataBus) (bool, string, error) {

	var param = map[string]interface{}{
		"Application":         "iTMSTransporter",
		"BaseVersion":         "2.0.0",
		"Files":               []string{bus.getString(dbsFilename), "metadata.xml"},
		"iTMSTransporterMode": "upload",
		"Metadata":            bus.getString(dbsMetadata),
		"MetadataChecksum":    bus.getString(dbsMetaChecksum),
		// "MetadataCompressed":  bus.getString(dbsMetaCompressed),
		"MetadataInfo": map[string]interface{}{
			"app_platform":                "ios",
			"apple_id":                    bus.getString(dbsAppleID),
			"asset_types":                 []string{"bundle"},
			"bundle_identifier":           bus.getString(dbsBundleid),
			"bundle_short_version_string": bus.getString(dbsBundleShortVersion),
			"bundle_version":              bus.getString(dbsBundleVersion),
			"device_id":                   "",
			"packageVersion":              "software5.4",
			"primary_bundle_identifier":   "",
		},
		"PackageName": bus.getString(dbsPackageName),
		"PackageSize": bus.getInt(dbsFileSize) + bus.getInt(dbsMetaSize),
		"Username":    bus.getString(dbsUsername),
		"Version":     "2.0.0",
	}
	if rp, err := makeProducerServiceReq(bus, cstValidMetadata, param); err != nil {
		return false, "", err
	} else {

		s, ok := (*rp)["Success"].(bool)
		if ok && s {
			return true, "", nil
		} else {
			info := (*rp)["Errors"].([]interface{})[0]
			return false, fmt.Sprintln(info), nil
		}

	}

}

func validAssets(bus *dataBus) (bool, string, error) {

	var param = map[string]interface{}{
		"Application":                 "iTMSTransporter",
		"BaseVersion":                 "2.0.0",
		"AssetDescriptionsCompressed": []interface{}{},
		"Files":                       []string{bus.getString(dbsFilename), "metadata.xml"},
		"iTMSTransporterMode":         "upload",
		"Metadata":                    bus.getString(dbsMetadata),
		"MetadataChecksum":            bus.getString(dbsMetaChecksum),
		// "MetadataCompressed":          bus.getString(dbsMetaCompressed),
		"MetadataInfo": map[string]interface{}{
			"app_platform":                "ios",
			"apple_id":                    bus.getString(dbsAppleID),
			"asset_types":                 []string{"bundle"},
			"bundle_identifier":           bus.getString(dbsBundleid),
			"bundle_short_version_string": bus.getString(dbsBundleShortVersion),
			"bundle_version":              bus.getString(dbsBundleVersion),
			"device_id":                   "",
			"packageVersion":              "software5.4",
			"primary_bundle_identifier":   "",
		},
		"PackageName":       bus.getString(dbsPackageName),
		"PackageSize":       bus.getInt(dbsFileSize) + bus.getInt(dbsMetaSize),
		"StreamingInfoList": []interface{}{},
		"Transport":         "HTTP",
		"Username":          bus.getString(dbsUsername),
		"Version":           "2.0.0",
	}
	if rp, err := makeProducerServiceReq(bus, cstValidAssets, param); err != nil {
		return false, "", err
	} else {

		s, ok := (*rp)["Success"].(bool)
		if ok && s {
			bus.set(dbsPackageName, (*rp)["NewPackageName"])
			return true, "", nil
		} else {
			info := (*rp)["Errors"].([]interface{})[0]
			return false, fmt.Sprintln(info), nil
		}

	}
}

func clientCheckSum(bus *dataBus) (bool, string, error) {

	para := map[string]interface{}{
		"Application":         "iTMSTransporter",
		"BaseVersion":         "2.0.0",
		"iTMSTransporterMode": "upload",
		"NewPackageName":      bus.getString(dbsPackageName),
		"Username":            bus.getString(dbsUsername),
		"Version":             "2.0.0",
	}
	if rp, err := makeProducerServiceReq(bus, cstCleintChecksum, para); err != nil {
		return false, "", err
	} else {
		s, ok := (*rp)["Success"].(bool)
		if ok && s {
			return true, "", nil
		} else {
			info := (*rp)["Errors"].([]interface{})[0]
			return false, fmt.Sprintln(info), nil
		}

	}

}

func createReservation(bus *dataBus) ([]interface{}, error) {
	para := map[string]interface{}{
		"Application": "iTMSTransporter",
		"BaseVersion": "2.0.0",
		"fileDescriptions": []map[string]interface{}{
			{
				"checksum":          bus.getString(dbsMetaChecksum),
				"checksumAlgorithm": "MD5",
				"contentType":       "application/xml",
				"fileName":          "metadata.xml",
				"fileSize":          bus.getInt(dbsMetaSize),
			},
			{
				"checksum":          bus.getString(dbsFileCheckSum),
				"checksumAlgorithm": "MD5",
				"contentType":       "application/octet-stream",
				"fileName":          bus.getString(dbsFilename),
				"fileSize":          bus.getInt(dbsFileSize),
				"uti":               "com.apple.ipa",
			},
		},

		"iTMSTransporterMode": "upload",
		"NewPackageName":      bus.getString(dbsPackageName),
		"Username":            bus.getString(dbsUsername),
		"Version":             "2.0.0",
	}
	if rp, err := makeProducerServiceReq(bus, cstCreateReservation, para); err != nil {
		return nil, err
	} else {
		s, ok := (*rp)["Success"].(bool)
		if ok && s {
			reservations := (*rp)["Reservations"].([]interface{})
			return reservations, nil
		} else {
			info := (*rp)["Errors"].([]interface{})[0]
			return nil, errors.New(fmt.Sprintln(info))
		}

	}

}

func commitReservation(bus *dataBus, reservation *map[string]interface{}) (bool, string, error) {
	para := map[string]interface{}{
		"Application":         "iTMSTransporter",
		"BaseVersion":         "2.0.0",
		"iTMSTransporterMode": "upload",
		"NewPackageName":      bus.getString(dbsPackageName),
		"reservations":        []interface{}{(*reservation)["id"]},
		"Username":            bus.getString(dbsUsername),
		"Version":             "2.0.0",
	}
	if rp, err := makeProducerServiceReq(bus, cstCommitRev, para); err != nil {
		return false, "", err
	} else {

		s, ok := (*rp)["Success"].(bool)
		if ok && s {
			return true, "", nil
		} else {
			info := (*rp)["Errors"].([]interface{})[0]
			return false, fmt.Sprintln(info), nil
		}
	}

}

func uploadDone(bus *dataBus, lastTime time.Duration) (bool, string, error) {

	para := map[string]interface{}{
		"Application": "iTMSTransporter",
		"BaseVersion": "2.0.0",
		"FileSizeInfo": map[string]int{
			bus.getString(dbsFilename): bus.getInt(dbsFileSize),
			"metadata.xml":             bus.getInt(dbsMetaSize),
		},
		"ClientChecksumInfo": []map[string]interface{}{
			{
				"CalculatedChecksum": bus.getString(dbsFileCheckSum),
				"CalculationTime":    100,
				"FileLastModified":   bus.getInt(dbsModifedTime),
				"Filename":           bus.getString(dbsFilename),
				"fileSize":           bus.getInt(dbsFileSize),
			},
		},

		"StatisticsArray":        []string{},
		"StreamingInfoList":      []string{},
		"iTMSTransporterMode":    "upload",
		"PackagePathWithoutBase": nil,
		"NewPackageName":         bus.getString(dbsPackageName),
		"Transport":              "HTTP",
		"TransferTime":           -int(lastTime.Milliseconds()),
		"NumberBytesTransferred": bus.getInt(dbsFileSize) + bus.getInt(dbsMetaSize),
		"Username":               bus.getString(dbsUsername),
		"Version":                "2.0.0",
	}

	if rp, err := makeProducerServiceReq(bus, cstDone, para); err != nil {
		return false, "", err
	} else {
		s, ok := (*rp)["Success"].(bool)
		if ok && s {
			return true, "", nil
		} else {
			info := (*rp)["Errors"].([]interface{})[0]
			return false, fmt.Sprintln(info), nil
		}
	}
}
