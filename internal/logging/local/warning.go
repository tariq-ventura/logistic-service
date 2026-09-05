package logging_local

func (lc *LocalClient) LogWarning(msg string, fields map[string]any) {
	lc.logger.WithFields(fields).Warn(msg)
}
