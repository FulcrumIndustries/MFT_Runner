module MFT_Runner

go 1.21

require (
	github.com/jlaffaye/ftp v0.2.0
	github.com/pkg/sftp v1.13.5
	golang.org/x/crypto v0.31.0
)

require (
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/kr/fs v0.1.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)

replace (
	MFT_Runner => ./
	MFT_Runner/internal/events => ./internal/events
	MFT_Runner/internal/runner => ./internal/runner
)
