package tail

import (
	"io"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type tail struct {
	r       *io.PipeReader
	w       *io.PipeWriter
	watcher *fsnotify.Watcher
	offset  int64
	done    chan struct{}
}

func New(name string) (io.ReadCloser, error) {
	// open the file
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	t := tail{done: make(chan struct{})}

	// figure out where the end is
	if t.offset, err = f.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}

	// close the file, for now
	if err = f.Close(); err != nil {
		return nil, err
	}

	// start watching the file for changes
	if t.watcher, err = fsnotify.NewWatcher(); err != nil {
		log.Fatal(err)
	}

	if err = t.watcher.Add(name); err != nil {
		return nil, err
	}

	t.r, t.w = io.Pipe()

	go t.watch()

	return &t, nil
}

func (t *tail) watch() {
	for {
		select {
		case <-t.done:
			return
		case event, ok := <-t.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				f, err := os.Open(event.Name)
				if err != nil {
					continue
				}

				if _, err = f.Seek(t.offset, io.SeekStart); err != nil {
					f.Close() // #nosec
					continue
				}

				n, err := io.Copy(t.w, f)
				f.Close() // #nosec
				if err == nil {
					t.offset += n
				}
			}

			// if event.Op&fsnotify.Remove == fsnotify.Remove {
			// 	// TODO(jrubin) close? eof?
			// }

			// if event.Op&fsnotify.Rename == fsnotify.Rename {
			// 	// TODO(jrubin) close? eof?
			// }
		case err, ok := <-t.watcher.Errors:
			if !ok {
				// TODO(jrubin)
				panic(err)
			}
		}
	}
}

func (t *tail) Read(data []byte) (int, error) {
	return t.r.Read(data)
}

func (t *tail) Close() error {
	var err error

	if cerr := t.r.Close(); cerr != nil {
		err = cerr
	}

	if cerr := t.w.Close(); cerr != nil {
		err = cerr
	}

	if cerr := t.watcher.Close(); cerr != nil {
		err = cerr
	}

	close(t.done)

	return err
}
