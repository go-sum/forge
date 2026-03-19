package core

import (
	g "maragu.dev/gomponents"
	h "maragu.dev/gomponents/html"
)

type avatarNS struct{}

// Avatar groups avatar sub-components under a namespace: Avatar.Root, Avatar.Image, Avatar.Fallback.
var Avatar avatarNS

// Root renders a circular avatar wrapper <span>.
func (avatarNS) Root(children ...g.Node) g.Node {
	return h.Span(
		h.Class("relative flex h-10 w-10 shrink-0 overflow-hidden rounded-full"),
		g.Group(children),
	)
}

// Image renders the avatar <img>.
func (avatarNS) Image(src, alt string, extra ...g.Node) g.Node {
	nodes := []g.Node{
		h.Class("aspect-square h-full w-full"),
		h.Src(src),
		h.Alt(alt),
	}
	nodes = append(nodes, extra...)
	return h.Img(nodes...)
}

// Fallback renders the fallback content shown when the image is unavailable.
func (avatarNS) Fallback(children ...g.Node) g.Node {
	return h.Span(
		h.Class("flex h-full w-full items-center justify-center rounded-full bg-muted"),
		g.Group(children),
	)
}
