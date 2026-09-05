package logging_local

func (lc *LocalClient) LogInfo(msg string, fields map[string]any) {
	lc.logger.WithFields(fields).Info(msg)
}
