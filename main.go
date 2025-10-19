// Pejelagarto Translator - A reversible fictional language translator
// Build command: go build -o pejelagarto-translator main.go
// Run command: go run main.go
// The server will start on http://localhost:8080

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"unicode"
)

// Translation maps for the Pejelagarto language
var (
	// wordMap replaces whole words or syllables
	wordMap = map[string]string{
		"hello":   "jetzo",
		"world":   "vorlag",
		"the":     "ze",
		"is":      "ez",
		"and":     "ung",
		"you":     "yux",
		"are":     "irr",
		"this":    "zez",
		"that":    "zit",
		"will":    "vell",
		"can":     "kin",
		"have":    "jiv",
		"with":    "veez",
		"from":    "frux",
		"they":    "zey",
		"what":    "vit",
		"when":    "ven",
		"where":   "verr",
		"how":     "jov",
		"why":     "vey",
		"good":    "gux",
		"great":   "grit",
		"thank":   "zink",
		"please":  "plix",
		"sorry":   "surry",
		"yes":     "yiz",
		"no":      "nux",
		"maybe":   "mibby",
		"friend":  "frund",
		"love":    "luv",
		"time":    "tym",
		"day":     "diy",
		"night":   "nyt",
		"morning": "murneng",
		"evening": "ivneng",
	}

	// conjunctionMap replaces letter pairs
	conjunctionMap = map[string]string{
		"ch": "jj",
		"sh": "xx",
		"th": "zz",
		"ph": "ff",
		"ck": "kk",
		"ng": "gg",
		"qu": "kv",
	}

	// letterMap replaces single letters (must be invertible)
	letterMap = map[string]string{
		"a": "i",
		"e": "o",
		"i": "a",
		"o": "e",
		"u": "y",
	}
)

// invertMap creates a reverse mapping from the original map
func invertMap(m map[string]string) map[string]string {
	inverted := make(map[string]string)
	for k, v := range m {
		inverted[v] = k
	}
	return inverted
}

// makeBidirectionalMap creates a bidirectional map by adding inverse mappings
// This prevents Pejelagarto sequences in the input from being mis-translated
func makeBidirectionalMap(m map[string]string) map[string]string {
	bidirectional := make(map[string]string)
	// Add original mappings
	for k, v := range m {
		bidirectional[k] = v
	}
	// Add inverse mappings (value maps to itself to prevent re-translation)
	for _, v := range m {
		if _, exists := bidirectional[v]; !exists {
			bidirectional[v] = v // Identity mapping
		}
	}
	return bidirectional
}

// applyMapReplacementsToPejelagarto applies the translation maps in order: word, conjunction, letter
// Uses a greedy approach: at each position, tries to match the longest pattern first
func applyMapReplacementsToPejelagarto(input string) string {
	// Make maps bidirectional to prevent Pejelagarto sequences in input from being mis-translated
	bidirectionalWordMap := makeBidirectionalMap(wordMap)
	bidirectionalConjunctionMap := makeBidirectionalMap(conjunctionMap)

	runes := []rune(input)
	result := []rune{}

	for i := 0; i < len(runes); {
		// Try wordMap first (case-insensitive) - find longest match
		longestWordValue := ""
		longestWordLen := 0
		for key, value := range bidirectionalWordMap {
			keyRunes := []rune(strings.ToLower(key))
			if len(keyRunes) > longestWordLen && i+len(keyRunes) <= len(runes) {
				// Check if we have a match (case-insensitive)
				match := true
				for j := 0; j < len(keyRunes); j++ {
					if unicode.ToLower(runes[i+j]) != keyRunes[j] {
						match = false
						break
					}
				}
				if match {
					longestWordValue = value
					longestWordLen = len(keyRunes)
				}
			}
		}
		if longestWordLen > 0 {
			// Preserve capitalization of first letter
			valueRunes := []rune(longestWordValue)
			if unicode.IsUpper(runes[i]) && len(valueRunes) > 0 {
				valueRunes[0] = unicode.ToUpper(valueRunes[0])
			}
			result = append(result, valueRunes...)
			i += longestWordLen
			continue
		}

		// Try conjunctionMap - find longest match
		longestConjValue := ""
		longestConjLen := 0
		for key, value := range bidirectionalConjunctionMap {
			keyRunes := []rune(key)
			if len(keyRunes) > longestConjLen && i+len(keyRunes) <= len(runes) {
				match := true
				for j := 0; j < len(keyRunes); j++ {
					if runes[i+j] != keyRunes[j] {
						match = false
						break
					}
				}
				if match {
					longestConjValue = value
					longestConjLen = len(keyRunes)
				}
			}
		}
		if longestConjLen > 0 {
			result = append(result, []rune(longestConjValue)...)
			i += longestConjLen
			continue
		}

		// Try letterMap
		charMatched := false
		for key, value := range letterMap {
			if string(runes[i]) == key {
				result = append(result, []rune(value)...)
				charMatched = true
				break
			}
		}
		if charMatched {
			i++
		} else {
			// No match, keep original character
			result = append(result, runes[i])
			i++
		}
	}

	return string(result)
}

// applyMapReplacementsFromPejelagarto reverses the translation in reverse order: word, conjunction, letter
// Uses a greedy approach: at each position, tries to match the longest pattern first
func applyMapReplacementsFromPejelagarto(input string) string {
	// Invert the maps
	invertedLetterMap := invertMap(letterMap)
	invertedConjunctionMap := invertMap(conjunctionMap)
	invertedWordMap := invertMap(wordMap)

	runes := []rune(input)
	result := []rune{}

	for i := 0; i < len(runes); {
		// Try inverted wordMap first (case-insensitive) - find longest match
		longestWordValue := ""
		longestWordLen := 0
		for key, value := range invertedWordMap {
			keyRunes := []rune(strings.ToLower(key))
			if len(keyRunes) > longestWordLen && i+len(keyRunes) <= len(runes) {
				// Check if we have a match (case-insensitive)
				match := true
				for j := 0; j < len(keyRunes); j++ {
					if unicode.ToLower(runes[i+j]) != keyRunes[j] {
						match = false
						break
					}
				}
				if match {
					longestWordValue = value
					longestWordLen = len(keyRunes)
				}
			}
		}
		if longestWordLen > 0 {
			// Preserve capitalization of first letter
			valueRunes := []rune(longestWordValue)
			if unicode.IsUpper(runes[i]) && len(valueRunes) > 0 {
				valueRunes[0] = unicode.ToUpper(valueRunes[0])
			}
			result = append(result, valueRunes...)
			i += longestWordLen
			continue
		}

		// Try inverted conjunctionMap - find longest match
		longestConjValue := ""
		longestConjLen := 0
		for key, value := range invertedConjunctionMap {
			keyRunes := []rune(key)
			if len(keyRunes) > longestConjLen && i+len(keyRunes) <= len(runes) {
				match := true
				for j := 0; j < len(keyRunes); j++ {
					if runes[i+j] != keyRunes[j] {
						match = false
						break
					}
				}
				if match {
					longestConjValue = value
					longestConjLen = len(keyRunes)
				}
			}
		}
		if longestConjLen > 0 {
			result = append(result, []rune(longestConjValue)...)
			i += longestConjLen
			continue
		}

		// Try inverted letterMap
		charMatched := false
		for key, value := range invertedLetterMap {
			if string(runes[i]) == key {
				result = append(result, []rune(value)...)
				charMatched = true
				break
			}
		}
		if charMatched {
			i++
		} else {
			// No match, keep original character
			result = append(result, runes[i])
			i++
		}
	}

	return string(result)
}

// TranslateToPejelagarto translates Human text to Pejelagarto
func TranslateToPejelagarto(input string) string {
	return applyMapReplacementsToPejelagarto(input)
}

// TranslateFromPejelagarto translates Pejelagarto text back to Human
func TranslateFromPejelagarto(input string) string {
	return applyMapReplacementsFromPejelagarto(input)
}

// HTML template for the web interface
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pejelagarto Translator</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            padding: 40px;
            max-width: 900px;
            width: 100%;
        }
        h1 {
            text-align: center;
            color: #667eea;
            margin-bottom: 30px;
            font-size: 2.5em;
        }
        .translator-box {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin-bottom: 20px;
        }
        .text-group {
            display: flex;
            flex-direction: column;
        }
        label {
            font-weight: 600;
            color: #333;
            margin-bottom: 8px;
            font-size: 1.1em;
        }
        textarea {
            width: 100%;
            min-height: 200px;
            padding: 15px;
            border: 2px solid #e0e0e0;
            border-radius: 10px;
            font-size: 16px;
            font-family: inherit;
            resize: vertical;
            transition: border-color 0.3s;
        }
        textarea:focus {
            outline: none;
            border-color: #667eea;
        }
        textarea[readonly] {
            background-color: #f5f5f5;
            cursor: default;
        }
        .controls {
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 15px;
            flex-wrap: wrap;
        }
        button {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 12px 30px;
            font-size: 16px;
            font-weight: 600;
            border-radius: 25px;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
        }
        button:active {
            transform: translateY(0);
        }
        .invert-btn {
            background: white;
            color: #667eea;
            border: 2px solid #667eea;
            padding: 10px 20px;
            font-size: 20px;
        }
        .invert-btn:hover {
            background: #667eea;
            color: white;
        }
        .checkbox-group {
            display: flex;
            align-items: center;
            gap: 8px;
        }
        input[type="checkbox"] {
            width: 20px;
            height: 20px;
            cursor: pointer;
        }
        .checkbox-label {
            font-weight: 500;
            color: #333;
            cursor: pointer;
            margin: 0;
        }
        @media (max-width: 768px) {
            .translator-box {
                grid-template-columns: 1fr;
            }
            h1 {
                font-size: 2em;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Pejelagarto Translator</h1>
        <div class="translator-box">
            <div class="text-group">
                <label id="input-label">Human</label>
                <textarea id="input-text" placeholder="Type here..."></textarea>
            </div>
            <div class="text-group">
                <label id="output-label">Pejelagarto</label>
                <textarea id="output-text" readonly placeholder="Translation will appear here..."></textarea>
            </div>
        </div>
        <div class="controls">
            <button 
                id="translate-btn"
                onclick="translateText(); return false;"
            >
                Translate to Pejelagarto
            </button>
            <button class="invert-btn" id="invert-btn">⇅</button>
            <div class="checkbox-group">
                <input type="checkbox" id="live-translate">
                <label for="live-translate" class="checkbox-label">Live Translation</label>
            </div>
        </div>
    </div>

    <script>
        let isInverted = false;
        let liveTranslateEnabled = false;

        const inputText = document.getElementById('input-text');
        const outputText = document.getElementById('output-text');
        const inputLabel = document.getElementById('input-label');
        const outputLabel = document.getElementById('output-label');
        const translateBtn = document.getElementById('translate-btn');
        const invertBtn = document.getElementById('invert-btn');
        const liveTranslateCheckbox = document.getElementById('live-translate');

        // Invert button handler
        invertBtn.addEventListener('click', () => {
            // Swap text content
            const temp = inputText.value;
            inputText.value = outputText.value;
            outputText.value = temp;

            // Swap labels
            const tempLabel = inputLabel.textContent;
            inputLabel.textContent = outputLabel.textContent;
            outputLabel.textContent = tempLabel;

            // Toggle button state
            isInverted = !isInverted;
            if (isInverted) {
                translateBtn.textContent = 'Translate from Pejelagarto';
                translateBtn.setAttribute('hx-post', '/from');
            } else {
                translateBtn.textContent = 'Translate to Pejelagarto';
                translateBtn.setAttribute('hx-post', '/to');
            }

            // Re-initialize HTMX for the button
            htmx.process(translateBtn);
        });

        // Live translation handler
        liveTranslateCheckbox.addEventListener('change', (e) => {
            liveTranslateEnabled = e.target.checked;
            
            if (liveTranslateEnabled) {
                translateBtn.style.display = 'none';
                inputText.addEventListener('input', handleLiveTranslate);
                // Trigger initial translation
                handleLiveTranslate();
            } else {
                translateBtn.style.display = 'block';
                inputText.removeEventListener('input', handleLiveTranslate);
            }
        });

        // Live translate function
        function handleLiveTranslate() {
            const endpoint = isInverted ? '/from' : '/to';
            
            fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'text/plain',
                },
                body: inputText.value
            })
            .then(response => response.text())
            .then(data => {
                outputText.value = data;
            })
            .catch(error => {
                console.error('Translation error:', error);
            });
        }

        // Manual translate function for button
        function translateText() {
            const endpoint = isInverted ? '/from' : '/to';
            
            fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'text/plain',
                },
                body: inputText.value
            })
            .then(response => response.text())
            .then(data => {
                outputText.value = data;
            })
            .catch(error => {
                console.error('Translation error:', error);
            });
        }
    </script>
</body>
</html>`

// Handler for the root path - serves the HTML UI
func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, htmlTemplate)
}

// Handler for translating TO Pejelagarto
func handleTranslateTo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	input := string(body)
	translated := TranslateToPejelagarto(input)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, translated)
}

// Handler for translating FROM Pejelagarto
func handleTranslateFrom(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}

	input := string(body)
	translated := TranslateFromPejelagarto(input)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, translated)
}

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Printf("Failed to open browser: %v\n", err)
	}
}

func main() {
	// Set up routes
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/to", handleTranslateTo)
	http.HandleFunc("/from", handleTranslateFrom)

	// Start server
	addr := ":8080"
	url := fmt.Sprintf("http://localhost%s", addr)

	fmt.Printf("Starting Pejelagarto Translator server on %s\n", url)
	fmt.Println("Press Ctrl+C to stop the server")

	// Open browser after a short delay to ensure server is ready
	go func() {
		time.Sleep(500 * time.Millisecond)
		fmt.Printf("Opening %s in your browser...\n", url)
		openBrowser(url)
	}()

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("Server error:", err)
	}
}
