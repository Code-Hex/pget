package pget

// Pget structs
type Pget struct {
	ARGV  []string
	Trace bool
	procs int
	args  []string
	url   string
	Utils
}

// Options struct for parse command line arguments
type Options struct {
	Help    bool   `short:"h" long:"help" description:"print usage and exit"`
	Version bool   `short:"v" long:"version" description:"display the version of pget and exit"`
	Procs   int    `short:"p" long:"procs" description:"split ratio to download file"`
	Output  string `short:"o" long:"output" description:"output file to FILENAME"`
	Trace   bool   `long:"trace" description:"display detail error messages"`
	// File    string `long:"file" description:"urls has same hash in a file to download"`
}

// Data struct has file of relational data
type Data struct {
	filename string
	filesize uint64
	dirname  string
}

// Utils interface indicate function
type Utils interface {
	ProgressBar() error
	BindwithFiles(int) error
	IsFree(uint64) error

	// like setter
	SetFileName(string)
	URLFileName(string)
	SetDirName(string)
	SetFileSize(uint64)

	// like getter
	FileName() string
	FileSize() uint64
	DirName() string
}

type cause interface {
	Cause() error
}
