package internal

import "jrubin.io/httpmon/clf"

type Handler interface {
	Handle(clf.Entry)
}

type Renderer interface {
	Render()
}
