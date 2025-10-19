You are an expert Golang developer. Your task is to create a complete, locally runnable web application that functions as a translator for a fictional language called "Pejelagarto". The project will involve a Go backend and a dynamic frontend powered by HTMX and simple, vanilla JavaScript (no npm or external libraries allowed).

Core Task: Create a Go application in a single main.go file that starts a local web server. The server will provide a simple, interactive UI to translate text to and from "Pejelagarto". The translation logic must be perfectly reversible.

Backend Requirements (Go):

Translation Maps: The core logic relies on three distinct replacement maps (map[string]string).
wordMap: Replaces words or syllables (e.g., {"hello": "jetzo", "world": "vorlag"}).
conjunctionMap: Replaces letter pairs (e.g., {"ch": "jj", "sh": "xx"}).
letterMap: Replaces single, invertible letters (e.g., {"a": "i", "i": "a"}).
Core Replacement Logic:
applyMapReplacementsToPejelagarto(input string) string: This function contains the primary replacement logic. It must process the input by applying wordMap, then conjunctionMap, then letterMap. To prevent re-replacing, every successful replacement must be wrapped in temporary tags (<rep> and </rep>). These tags are removed in a final step before returning.
applyMapReplacementsFromPejelagarto(input string) string: This function reverses the process. It uses inverted maps and applies them in the reverse order: inverted letterMap, inverted conjunctionMap, then inverted wordMap, also using the tagging strategy.
Main Translation Functions:
TranslateToPejelagarto(input string) string: This function acts as a wrapper and, for now, must simply call applyMapReplacementsToPejelagarto.
TranslateFromPejelagarto(input string) string: This wrapper must call applyMapReplacementsFromPejelagarto.
Web Server:
The main function should start a web server on http://localhost:8080.
It must serve the main HTML UI from the root path (/).
It must handle POST requests from the frontend at /to and /from, using the translation functions to process the request body and return the translated text.
Frontend Requirements (HTML/JS/HTMX):

Single File UI: The complete HTML, CSS, and JavaScript must be stored in a Go string constant and served by the web server.
UI Elements:
A main title: "Pejelagarto Translator".
An input textarea with a label (e.g., "English").
An output textarea (read-only) with a label (e.g., "Pejelagarto").
A single main translation button (e.g., "Translate to Pejelagarto").
An "Invert" button (e.g., with "⇅" or similar icon).
A checkbox with the label "Live Translation".
Interactivity (JavaScript & HTMX):
Invert Button: When clicked, the JavaScript should:
Swap the text content of the input and output textarea elements.
Swap the labels for the textarea elements.
Toggle the state of the main translation button: change its text (e.g., "Translate to Pejelagarto" <-> "Translate from Pejelagarto") and update its hx-post attribute to point to the correct endpoint (/to or /from).
Live Translation Checkbox: When checked, the JavaScript should:
Hide the main translation button.
Add an event listener to the input textarea (e.g., on input or keyup).
On input, automatically trigger the HTMX request that the (now hidden) button would have, to perform the translation in real-time.
When unchecked, the JavaScript should show the main translation button again and remove the live-translation event listener.
HTMX Core: The main translation button (and the live translation) will use HTMX attributes (hx-post, hx-trigger, hx-target, hx-include) to send the input textarea content to the backend and place the response in the output textarea.
Testing Requirements:

Reversibility Test: Create a main_test.go file.
Inside it, create a function TestReversibility that verifies the translation is perfectly reversible by checking both:
TranslateToPejelagarto(TranslateFromPejelagarto(originalText)) == originalText
TranslateFromPejelagarto(TranslateToPejelagarto(originalText)) == originalText
The test should use a comprehensive default string that utilizes replacements from all three maps.
Final Deliverable: A single, well-commented main.go file and a main_test.go file. The comments should include the build command needed to compile the server.