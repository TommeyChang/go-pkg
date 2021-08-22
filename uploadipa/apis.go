package uploadipa

func RunUpload(userName, appPWD, filePath string) (result bool, response string, err  error) {
	proc, err := NewUploader(userName, appPWD, filePath)
	if err != nil {
		return false, "file error", err
	}
	tickets, err := proc.GetUploadTickets()
	if err != nil {
		return false, err.Error(), err
	}
	return proc.UploadWithTickets(tickets)
}