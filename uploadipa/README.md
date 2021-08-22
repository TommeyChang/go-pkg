# ipaupload
Upload iOS apps to App Store Connect with GoLang


Our code is based on the codes [ios-uploader](<https://github.com/simonnilsson/ios-uploader>) , a Node.js for uploading iOS apps to App Store Connect.

**Note**
* The appPWD is not the password for login, it is generated for uploade the specific app in you account.

## Simple Usage
```
import github.com/TommeyChang/go-pkg/uploadipa
result, response, err := uploadipa.RunUpload(user, appPWD, filePath)
```

**NOTE**:
* The upload is successful only when the result is true. Otherwise failed.
* Most of the responses were given by the Apple server. 
* The Apple AppStore supports block transfer. We set the default number of blocks is 4.
* The retry times is 3.

## Step by step usage
```
import github.com/TommeyChang/go-pkg/uploadipa

# You can set debug mode to print the details
uploadipa.SetDebuggerMode(True)

proc, err := uploadipa.NewUploader(userName, appPWD, filePath)
if err != nil {
  return false, "file error", err
}

proc.SetBlockNum = 4
proc.SetRetryTimes = 5

tickets, err := proc.GetUploadTickets()
if err != nil {
  return false, err.Error(), err
}

resutl, response, err := proc.UploadWithTickets(tickets)


```
