package ui

import "strings"

// The keys that belong to the shell rather than to whatever has focus.
//
// The CHAT view hands every key it does not recognise to its input field, which
// is right for a letter and wrong for Alt+F: a view that eats the FILE menu and
// the help is a view you cannot get out of except by the keys it happens to
// remember. It remembered F2 to F6 and nothing else, because the list was
// written twice - once as the main handler's switch, once as a case in the CHAT
// handler - and only one of them was kept up.
//
// This is the list. The view keys come from the tabs themselves, so a view
// added to the tab line is reachable from inside CHAT without anyone having to
// remember this file exists.

// shellKeys is every key the shell answers wherever you are.
func shellKeys() []string {
	keys := []string{
		// F1 is the help, and Alt+H as well because Terminator, Konsole and
		// others take F1 before an application sees it.
		"f1", "alt+h",

		// The FILE menu. There is no F-key for it: F2 to F6 are the views, and
		// a menu is not one.
		"alt+f",

		// Reading what is on screen with the AI. It does nothing in CHAT,
		// which is already a conversation, but a key that is global is global.
		"alt+a",

		// Leaving. A view that takes the quit key is a view you are stuck in.
		"ctrl+q",
	}

	for _, tab := range modeTabs {
		keys = append(keys, strings.ToLower(tab.key))
	}
	for _, tab := range resultTabs {
		keys = append(keys, strings.ToLower(tab.key))
	}
	return keys
}

// isShellKey reports whether a key belongs to the shell, and so must reach the
// main handler rather than being typed into whatever has focus.
func isShellKey(key string) bool {
	for _, shell := range shellKeys() {
		if key == shell {
			return true
		}
	}
	return false
}
