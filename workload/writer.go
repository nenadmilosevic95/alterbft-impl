package workload

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Writer is a helper to log data to a file.
type Writer struct {
	file      *os.File
	buffer    *bufio.Writer
	startTime int64
}

// NewWriter creates a new writer for a load generator.
func NewWriter(config *Config, id int) *Writer {
	filename := config.LogDirectory
	err := checkDirectory(filename)
	// FIXME: anything better than panicing here?
	if err != nil {
		panic(err)
	}

	filename += fmt.Sprint("/deliveries.", id)
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	return &Writer{
		file:   file,
		buffer: bufio.NewWriterSize(file, config.LogBufferSize),
	}
}

// Close writes all buffered bytes and closes the output file.
func (w *Writer) Close() {
	w.buffer.WriteString(fmt.Sprint("# DONE ",
		time.Now().Format(time.UnixDate), "\n"))
	err := w.buffer.Flush()
	if err == nil {
		err = w.file.Close()
	}
	if err != nil {
		panic(err)
	}
}

// LogDelivery logs the information of a delivery.
func (w *Writer) LogDelivery(d *Delivery) {
	var out strings.Builder
	out.WriteString(formatTime(d.Time))
	out.WriteString("\t")
	out.WriteString(fmt.Sprint(d.Height))
	if d.Height == 0 {
		w.startTime = time.Now().UnixMilli() - d.Latency().Milliseconds()
	}
	//	out.WriteString("\t")
	//	out.WriteString(fmt.Sprint(d.Epoch))
	//	out.WriteString("\t")
	//	out.WriteString(base64.RawStdEncoding.EncodeToString(d.BlockID))
	if d.Submission != nil && d.Height > 0 && d.Height%64 == 63 {
		latency := time.Now().UnixMilli() - w.startTime
		w.startTime = time.Now().UnixMilli()
		out.WriteString("\t")
		out.WriteString(fmt.Sprintf("%.6f", float64(latency/1000)))
	}
	out.WriteString("\n")
	w.buffer.WriteString(out.String())
}

// String returns string information about this writer.
func (w *Writer) String() string {
	fname, _ := filepath.Abs(w.file.Name())
	return fmt.Sprintf("'%s' buffer size %d", fname, w.buffer.Size())
}

//// Helpers

func formatTime(t time.Time) string {
	return t.Format("15:04:05.00000")
}

func checkDirectory(dir string) error {
	_, err := os.Stat(dir)
	return err
}

func createDirectory(dir string) error {
	return os.Mkdir(dir, 0755)
}
