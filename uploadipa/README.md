# ipaupload
Upload IPA package with GoLang

## Note
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

## Lincense

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
