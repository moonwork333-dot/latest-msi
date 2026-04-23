package main

import (
	"fmt"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	procSetCursorPos = user32.NewProc("SetCursorPos")
	procSendInput    = user32.NewProc("SendInput")
)

const (
	inputMouse    = 0
	inputKeyboard = 1

	mouseeventfMove      = 0x0001
	mouseeventfLeftdown  = 0x0002
	mouseeventfLeftup    = 0x0004
	mouseeventfRightdown = 0x0008
	mouseeventfRightup   = 0x0010
	mouseeventfAbsolute  = 0x8000

	keyeventfKeyup   = 0x0002
	keyeventfUnicode = 0x0004

	vkReturn = 0x0D
	vkBack   = 0x08
	vkDelete = 0x2E
	vkTab    = 0x09
	vkEscape = 0x1B
	vkLeft   = 0x25
	vkUp     = 0x26
	vkRight  = 0x27
	vkDown   = 0x28
	vkHome   = 0x24
	vkEnd    = 0x23
	vkPgUp   = 0x21
	vkPgDn   = 0x22
)

type mouseInput struct {
	Type  uint32
	_     [4]byte
	Dx    int32
	Dy    int32
	Data  uint32
	Flags uint32
	Time  uint32
	Extra uintptr
}

type keybdInput struct {
	Type  uint32
	_     [4]byte
	Vk    uint16
	Scan  uint16
	Flags uint32
	Time  uint32
	Extra uintptr
}

func MouseMove(x, y int) error {
	ret, _, err := procSetCursorPos.Call(uintptr(x), uintptr(y))
	if ret == 0 {
		return fmt.Errorf("SetCursorPos failed: %w", err)
	}
	return nil
}

func MouseClick(button string) error {
	var downFlag, upFlag uint32
	switch button {
	case "right":
		downFlag = mouseeventfRightdown
		upFlag = mouseeventfRightup
	default:
		downFlag = mouseeventfLeftdown
		upFlag = mouseeventfLeftup
	}

	inputs := [2]mouseInput{
		{Type: inputMouse, Flags: downFlag},
		{Type: inputMouse, Flags: upFlag},
	}

	ret, _, err := procSendInput.Call(
		2,
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(unsafe.Sizeof(inputs[0])),
	)
	if ret == 0 {
		return fmt.Errorf("SendInput (click) failed: %w", err)
	}
	return nil
}

func KeyPress(key string) error {
	vk, ok := resolveVK(key)
	if !ok {
		return KeyType(key)
	}

	inputs := [2]keybdInput{
		{Type: inputKeyboard, Vk: uint16(vk)},
		{Type: inputKeyboard, Vk: uint16(vk), Flags: keyeventfKeyup},
	}

	ret, _, err := procSendInput.Call(
		2,
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(unsafe.Sizeof(inputs[0])),
	)
	if ret == 0 {
		return fmt.Errorf("SendInput (key) failed: %w", err)
	}
	return nil
}

func KeyType(text string) error {
	utf16Chars := utf16.Encode([]rune(text))
	inputs := make([]keybdInput, 0, len(utf16Chars)*2)
	for _, ch := range utf16Chars {
		inputs = append(inputs,
			keybdInput{Type: inputKeyboard, Scan: ch, Flags: keyeventfUnicode},
			keybdInput{Type: inputKeyboard, Scan: ch, Flags: keyeventfUnicode | keyeventfKeyup},
		)
	}
	if len(inputs) == 0 {
		return nil
	}
	ret, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(unsafe.Sizeof(inputs[0])),
	)
	if ret == 0 {
		return fmt.Errorf("SendInput (type) failed: %w", err)
	}
	return nil
}

func resolveVK(key string) (int, bool) {
	vkMap := map[string]int{
		"enter": vkReturn, "return": vkReturn,
		"backspace": vkBack, "delete": vkDelete,
		"tab": vkTab, "escape": vkEscape, "esc": vkEscape,
		"left": vkLeft, "up": vkUp, "right": vkRight, "down": vkDown,
		"home": vkHome, "end": vkEnd,
		"pageup": vkPgUp, "pagedown": vkPgDn,
	}
	vk, ok := vkMap[key]
	return vk, ok
}

var _ = syscall.NewLazyDLL
