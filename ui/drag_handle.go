package ui

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const replayQueueRowHeight = 36

type dragHandle struct {
	widget.BaseWidget
	text     *canvas.Text
	itemID   int
	index    int
	deltaY   float32
	original fyne.Position
	dragging bool
	onDrop   func(itemID, from, to int)
}

func newDragHandle(onDrop func(itemID, from, to int)) *dragHandle {
	handle := &dragHandle{
		text:   canvas.NewText("≡", nil),
		onDrop: onDrop,
	}
	handle.text.TextSize = 20
	handle.ExtendBaseWidget(handle)
	return handle
}

func (h *dragHandle) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.text)
}

func (h *dragHandle) Dragged(event *fyne.DragEvent) {
	if !h.dragging {
		h.original = h.text.Position()
		h.dragging = true
	}
	h.deltaY += event.Dragged.DY
	h.text.Move(h.original.Add(fyne.NewDelta(0, h.deltaY)))
	canvas.Refresh(h.text)
}

func (h *dragHandle) DragEnd() {
	offset := int(math.Round(float64(h.deltaY) / replayQueueRowHeight))
	h.text.Move(h.original)
	canvas.Refresh(h.text)
	if offset != 0 {
		h.onDrop(h.itemID, h.index, h.index+offset)
	}
	h.deltaY = 0
	h.dragging = false
}
