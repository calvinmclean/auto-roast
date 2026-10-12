package ui

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const replayQueueRowHeight = 36

type dragHandle struct {
	widget.BaseWidget
	icon     *canvas.Image
	itemID   int
	index    int
	deltaY   float32
	original fyne.Position
	dragging bool
	onDrop   func(itemID, from, to int)
}

func newDragHandle(onDrop func(itemID, from, to int)) *dragHandle {
	resource := fyne.NewStaticResource("drag-handle.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20"><g fill="#000"><circle cx="7" cy="5" r="1.5"/><circle cx="13" cy="5" r="1.5"/><circle cx="7" cy="10" r="1.5"/><circle cx="13" cy="10" r="1.5"/><circle cx="7" cy="15" r="1.5"/><circle cx="13" cy="15" r="1.5"/></g></svg>`))
	icon := canvas.NewImageFromResource(theme.NewColoredResource(resource, theme.ColorNamePlaceHolder))
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(20, 20))
	handle := &dragHandle{
		icon:   icon,
		onDrop: onDrop,
	}
	handle.ExtendBaseWidget(handle)
	return handle
}

func (h *dragHandle) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.icon)
}

func (h *dragHandle) Dragged(event *fyne.DragEvent) {
	if !h.dragging {
		h.original = h.icon.Position()
		h.dragging = true
	}
	h.deltaY += event.Dragged.DY
	h.icon.Move(h.original.Add(fyne.NewDelta(0, h.deltaY)))
	canvas.Refresh(h.icon)
}

func (h *dragHandle) DragEnd() {
	rowHeight := h.Size().Height
	if rowHeight <= 0 {
		rowHeight = replayQueueRowHeight
	}
	offset := int(math.Round(float64(h.deltaY / rowHeight)))
	h.icon.Move(h.original)
	canvas.Refresh(h.icon)
	if offset != 0 {
		h.onDrop(h.itemID, h.index, h.index+offset)
	}
	h.deltaY = 0
	h.dragging = false
}
