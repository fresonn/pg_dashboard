package postgres

const PG_VERSION_QUERY = "SELECT version()"

func (s *Storage) Version() (string, error) {

	db, err := s.pgManager.SQLX()
	if err != nil {
		return "", err
	}

	var version string

	err = db.Get(&version, PG_VERSION_QUERY)
	if err != nil {
		return "", err
	}

	return version, nil
}
