// +build ignore

// Simple HTTP server for frontend mode
// This serves the WASM-enabled UI and TTS endpoints only
// Translation happens client-side in the browser via WASM

package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
	"regexp"
)

//go:embed get-requirements.ps1 get-requirements.sh
var embeddedGetRequirements embed.FS

// Import necessary server-side functions for TTS
// (These are copied from server_main.go since we can't import from build-constrained files)

var tempRequirementsDir string
var pronunciationLanguage string
var pronunciationLanguageDropdown bool

var audioCache = struct {
	sync.RWMutex
	cache map[string][]byte
}{
	cache: make(map[string][]byte),
}

const htmlUIFrontend = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pejelagarto Translator (Frontend Mode)</title>
    <script src="/wasm_exec.js"></script>
    <style>
        :root {
            --bg-primary: #f5f5f5;
            --bg-secondary: #ffffff;
            --text-primary: #333333;
            --text-secondary: #666666;
            --border-color: #cccccc;
            --button-bg: #4CAF50;
            --button-hover: #45a049;
            --textarea-bg: #ffffff;
            --shadow: rgba(0, 0, 0, 0.1);
        }

        [data-theme="dark"] {
            --bg-primary: #1a1a1a;
            --bg-secondary: #2d2d2d;
            --text-primary: #e0e0e0;
            --text-secondary: #b0b0b0;
            --border-color: #444444;
            --button-bg: #45a049;
            --button-hover: #4CAF50;
            --textarea-bg: #3a3a3a;
            --shadow: rgba(255, 255, 255, 0.1);
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, var(--bg-primary) 0%, var(--bg-secondary) 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
            color: var(--text-primary);
            transition: background 0.3s ease, color 0.3s ease;
        }

        .container {
            width: 100%;
            max-width: 900px;
            background: var(--bg-secondary);
            border-radius: 20px;
            box-shadow: 0 10px 40px var(--shadow);
            padding: 40px;
            transition: background 0.3s ease;
        }

        h1 {
            text-align: center;
            margin-bottom: 10px;
            font-size: 2.5rem;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }

        .mode-badge {
            text-align: center;
            margin-bottom: 30px;
            color: var(--text-secondary);
            font-size: 0.9rem;
        }

        .mode-badge span {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 4px 12px;
            border-radius: 12px;
            font-weight: bold;
        }

        .status {
            text-align: center;
            padding: 10px;
            margin-bottom: 20px;
            border-radius: 8px;
            font-size: 0.9rem;
        }

        .status.loading {
            background: #fff3cd;
            color: #856404;
            border: 1px solid #ffc107;
        }

        .status.ready {
            background: #d4edda;
            color: #155724;
            border: 1px solid #28a745;
        }

        .status.error {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #dc3545;
        }

        .translator-box {
            display: flex;
            flex-direction: column;
            gap: 20px;
        }

        .input-section, .output-section {
            flex: 1;
        }

        label {
            display: flex;
            align-items: center;
            gap: 10px;
            font-weight: 600;
            margin-bottom: 10px;
            color: var(--text-primary);
        }

        textarea {
            width: 100%;
            height: 200px;
            padding: 15px;
            border: 2px solid var(--border-color);
            border-radius: 12px;
            font-size: 16px;
            font-family: 'Courier New', monospace;
            resize: vertical;
            background: var(--textarea-bg);
            color: var(--text-primary);
            transition: all 0.3s ease;
        }

        textarea:focus {
            outline: none;
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }

        .controls {
            display: flex;
            gap: 15px;
            flex-wrap: wrap;
            justify-content: center;
            margin: 20px 0;
        }

        button, .play-btn {
            padding: 12px 24px;
            font-size: 16px;
            font-weight: 600;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.3s ease;
            box-shadow: 0 4px 6px var(--shadow);
        }

        button:hover, .play-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 12px var(--shadow);
        }

        button:active, .play-btn:active {
            transform: translateY(0);
        }

        .translate-btn {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }

        .invert-btn {
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            color: white;
            padding: 12px 20px;
        }

        .theme-toggle {
            background: var(--button-bg);
            color: white;
        }

        .theme-toggle:hover {
            background: var(--button-hover);
        }

        .play-btn {
            background: #4CAF50;
            color: white;
        }

        .checkbox-container {
            display: flex;
            align-items: center;
            gap: 10px;
            padding: 15px;
            background: var(--bg-primary);
            border-radius: 8px;
        }

        input[type="checkbox"] {
            width: 20px;
            height: 20px;
            cursor: pointer;
        }

        @media (max-width: 768px) {
            .container {
                padding: 20px;
            }

            h1 {
                font-size: 2rem;
            }

            .controls {
                flex-direction: column;
            }

            button, .play-btn {
                width: 100%;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🦎 Pejelagarto Translator</h1>
        <div class="mode-badge">
            <span>⚡ Frontend Mode</span> - Translation runs in your browser
        </div>
        
        <div id="status" class="status loading">
            🔄 Loading WebAssembly module...
        </div>

        <div class="translator-box">
            <div class="input-section">
                <label id="input-label">
                    Human: 
                    <button class="play-btn" id="play-input" onclick="playAudio('input', false)" style="display: none;">🔊</button>
                    {{DROPDOWN_PLACEHOLDER}}
                </label>
                <textarea id="input-text" placeholder="Enter text in Human language..."></textarea>
            </div>

            <div class="controls">
                <button class="translate-btn" onclick="translate()">Translate</button>
                <button class="invert-btn" onclick="invertDirection()">⇅ Invert</button>
                <button class="theme-toggle" onclick="toggleTheme()" id="theme-btn">🌙 Dark Mode</button>
            </div>

            <div class="output-section">
                <label id="output-label">
                    Pejelagarto:
                    <button class="play-btn" id="play-output" onclick="playAudio('output', false)">🔊</button>
                    {{DROPDOWN_PLACEHOLDER}}
                </label>
                <textarea id="output-text" placeholder="Translation will appear here..." readonly></textarea>
            </div>

            <div class="checkbox-container">
                <input type="checkbox" id="live-translate" onchange="toggleLiveTranslation()">
                <label for="live-translate" style="margin: 0;">Enable Live Translation</label>
            </div>
        </div>
    </div>

    <script>
        let isInverted = false;
        let liveTranslateEnabled = false;
        let wasmReady = false;

        // Load WASM module
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("/translator.wasm"), go.importObject).then((result) => {
            go.run(result.instance);
            wasmReady = true;
            document.getElementById('status').textContent = '✓ Ready - Translation happens in your browser!';
            document.getElementById('status').className = 'status ready';
        }).catch((err) => {
            console.error("Failed to load WASM:", err);
            document.getElementById('status').textContent = '✗ Failed to load WebAssembly module';
            document.getElementById('status').className = 'status error';
        });

        function translate() {
            if (!wasmReady) {
                alert('WASM module not ready yet. Please wait...');
                return;
            }

            const inputText = document.getElementById('input-text');
            const outputText = document.getElementById('output-text');

            if (!isInverted) {
                // Human to Pejelagarto (using WASM)
                outputText.value = GoTranslateToPejelagarto(inputText.value);
            } else {
                // Pejelagarto to Human (using WASM)
                outputText.value = GoTranslateFromPejelagarto(inputText.value);
            }
        }

        function invertDirection() {
            isInverted = !isInverted;
            const inputLabel = document.getElementById('input-label');
            const outputLabel = document.getElementById('output-label');

            if (isInverted) {
                inputLabel.childNodes[0].textContent = 'Pejelagarto: ';
                outputLabel.childNodes[0].textContent = 'Human: ';
                document.getElementById('input-text').placeholder = 'Enter text in Pejelagarto language...';
                document.getElementById('output-text').placeholder = 'Translation will appear here...';
            } else {
                inputLabel.childNodes[0].textContent = 'Human: ';
                outputLabel.childNodes[0].textContent = 'Pejelagarto: ';
                document.getElementById('input-text').placeholder = 'Enter text in Human language...';
                document.getElementById('output-text').placeholder = 'Translation will appear here...';
            }

            // Swap textareas
            const inputValue = document.getElementById('input-text').value;
            const outputValue = document.getElementById('output-text').value;
            document.getElementById('input-text').value = outputValue;
            document.getElementById('output-text').value = inputValue;

            if (liveTranslateEnabled) {
                translate();
            }
        }

        function toggleLiveTranslation() {
            liveTranslateEnabled = document.getElementById('live-translate').checked;
            if (liveTranslateEnabled) {
                document.getElementById('input-text').addEventListener('input', translate);
            } else {
                document.getElementById('input-text').removeEventListener('input', translate);
            }
        }

        function toggleTheme() {
            const body = document.body;
            const themeBtn = document.getElementById('theme-btn');
            
            if (body.hasAttribute('data-theme')) {
                body.removeAttribute('data-theme');
                themeBtn.textContent = '🌙 Dark Mode';
            } else {
                body.setAttribute('data-theme', 'dark');
                themeBtn.textContent = '☀️ Light Mode';
            }
        }

        // TTS functionality (server-side)
        function playAudio(source, slow) {
            const inputText = document.getElementById('input-text');
            const outputText = document.getElementById('output-text');
            
            let textToSpeak = '';
            let button = null;
            
            if (source === 'input') {
                textToSpeak = inputText.value;
                button = document.getElementById('play-input');
            } else {
                textToSpeak = outputText.value;
                button = document.getElementById('play-output');
            }
            
            if (!textToSpeak || textToSpeak.trim() === '') {
                alert('No text to convert to speech!');
                return;
            }
            
            button.disabled = true;
            const originalText = button.textContent;
            button.textContent = '⏳';
            
            const languageDropdown = document.getElementById('tts-language');
            const selectedLanguage = languageDropdown ? languageDropdown.value : '';
            
            let url = selectedLanguage ? '/tts?lang=' + selectedLanguage : '/tts';
            if (slow) {
                url += (selectedLanguage ? '&' : '?') + 'slow=true';
            }
            
            fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'text/plain'
                },
                body: textToSpeak
            })
            .then(async response => {
                if (!response.ok) {
                    const errorText = await response.text();
                    throw new Error(errorText || 'TTS request failed: ' + response.statusText);
                }
                return response.blob();
            })
            .then(blob => {
                const audioUrl = URL.createObjectURL(blob);
                const audio = new Audio(audioUrl);
                
                audio.onended = function() {
                    URL.revokeObjectURL(audioUrl);
                    button.disabled = false;
                    button.textContent = originalText;
                };
                
                audio.onerror = function() {
                    URL.revokeObjectURL(audioUrl);
                    button.disabled = false;
                    button.textContent = originalText;
                    alert('Error playing audio');
                };
                
                audio.play();
                button.textContent = '▶️';
            })
            .catch(error => {
                console.error('TTS error:', error);
                button.disabled = false;
                button.textContent = originalText;
                alert('Text-to-speech error:\\n\\n' + error.message);
            });
        }
    </script>
</body>
</html>`

// Handler for serving the frontend UI
func handleFrontendIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	html := htmlUIFrontend
	if pronunciationLanguageDropdown {
		dropdownHTML := ` <select id="tts-language" style="margin-left: 8px; padding: 4px 8px; border-radius: 4px; border: 1px solid var(--border-color); background: var(--textarea-bg); color: var(--text-primary); font-size: 14px;">
                    <option value="russian">North</option>
                    <option value="kazakh">North-North-East</option>
                    <option value="german">North-East</option>
                    <option value="turkish">North-East-East</option>
                    <option value="portuguese">East</option>
                    <option value="french">South-East-East</option>
                    <option value="hindi">South-East</option>
                    <option value="icelandic">South-South-East</option>
                    <option value="romanian">South</option>
                    <option value="vietnamese">South-South-West</option>
                    <option value="swahili">South-West</option>
                    <option value="swedish">South-West-West</option>
                    <option value="czech">West</option>
                    <option value="chinese">North-West-West</option>
                    <option value="norwegian">North-West</option>
                    <option value="hungarian">North-North-West</option>
                </select>`
		html = strings.Replace(html, "{{DROPDOWN_PLACEHOLDER}}", dropdownHTML, 1)
	} else {
		html = strings.Replace(html, "{{DROPDOWN_PLACEHOLDER}}", "", 1)
	}
	
	fmt.Fprint(w, html)
}

func main() {
	// Parse flags
	pronunciationLangFlag := flag.String("pronunciation_language", "russian", "TTS pronunciation language")
	pronunciationLangDropdownFlag := flag.Bool("pronunciation_language_dropdown", true, "Show language dropdown in UI for TTS")
	flag.Parse()
	
	pronunciationLanguage = *pronunciationLangFlag
	pronunciationLanguageDropdown = *pronunciationLangDropdownFlag
	
	log.Println("Starting Pejelagarto Translator in FRONTEND mode")
	log.Println("Translation: Client-side (WebAssembly)")
	log.Println("TTS Audio: Server-side")
	log.Printf("TTS Language: %s\n", pronunciationLanguage)
	
	// Initialize TTS
	log.Println("Initializing TTS requirements...")
	// TODO: Add extractEmbeddedRequirements() and TTS functions here
	
	// Serve static files
	http.HandleFunc("/", handleFrontendIndex)
	http.HandleFunc("/translator.wasm", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/wasm")
		http.ServeFile(w, r, "bin/translator.wasm")
	})
	http.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "bin/wasm_exec.js")
	})
	
	// TODO: Add TTS endpoints here
	
	addr := ":8080"
	url := "http://localhost:8080"
	
	log.Printf("Server starting on %s\n", url)
	log.Println("Open your browser to test WASM-powered translation!")
	
	// Open browser
	time.Sleep(500 * time.Millisecond)
	openBrowser(url)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
