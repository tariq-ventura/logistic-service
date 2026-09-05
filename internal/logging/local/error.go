package logging_local

func (lc *LocalClient) LogError(msg string, fields map[string]any) {
	lc.logger.WithFields(fields).Error(msg)
}
