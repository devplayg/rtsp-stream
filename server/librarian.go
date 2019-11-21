package server

import "time"

type Librarian struct {
	interval time.Duration // 1 min
	manager  *Manager
}

func (l *Librarian) work() {
	// streams

	//
	for {

	}
}

func (l *Librarian) organize(stream *Stream) error {
	// Read *.ts files in the live directory

	// Sort *.ts files

	// Merge *.ts files and rename to timestamp.mp4 format

	return nil
}

// Boltdb

/*
	streams
		key: id
		val: stream information

	records
		key: id

*/
