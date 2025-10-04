package progress

import "io"

type Stage string

const (
	StageDownload  Stage = "download"
	StageExtract   Stage = "extract"
	StageSearchBin Stage = "search_bin"
)

type Event struct {
	Stage   Stage
	Current int64
	Total   int64
	Done    bool
}

type Hooks struct {
	Callback  Callback
	Decorator Decorator
}

type Callback func(Event)
type Decorator func(stage Stage, r io.Reader, total int64) io.Reader

type Reader struct {
	reader   io.Reader
	callback Callback
	stage    Stage
	total    int64
	curr     int64
}

var Nop Callback = func(Event) {}

func NewReader(r io.Reader, total int64, st Stage, cb Callback) *Reader {
	if cb == nil {
		cb = Nop
	}
	return &Reader{
		reader:   r,
		callback: cb,
		stage:    st,
		total:    total,
		curr:     0,
	}
}

func (pr *Reader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.curr += int64(n)
		pr.callback(Event{
			Stage:   pr.stage,
			Current: pr.curr,
			Total:   pr.total,
		})
	}
	return n, err
}

func GetAsyncCallback(cb Callback, buf int) (wrapped Callback, stop func()) {
	ch := make(chan Event, buf)

	go func() {
		for ev := range ch {
			cb(ev)
		}
	}()

	wrapped = func(ev Event) {
		select {
		case ch <- ev:
		default:
		}
	}

	stop = func() { close(ch) }
	return
}
