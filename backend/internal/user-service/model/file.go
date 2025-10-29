package model

import "io"

type File struct {
	Filename    string
	Data        io.ReadCloser
	Size        int64
	ContentType string
}
