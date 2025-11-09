// Pejelagarto Translator Android App
// NOTE: This is a simple demonstration. gomobile doesn't support WebView natively.
// The app opens the device browser to http://localhost:8080 where a local server runs.
// For a native UI, you'd need to write a Java/Kotlin Android app with WebView that calls Go via gomobile bind.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"

	"pejelagarto-translator/internal/translator"
)

var serverStarted = false

func main() {
	app.Main(func(a app.App) {
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					if !serverStarted {
						go startServer()
						serverStarted = true
					}
				}
			}
		}
	})
}

func startServer() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/translate", handleTranslate)

	port := 8080
	serverURL := fmt.Sprintf("http://localhost:%d", port)

	// Start server in background
	go func() {
		log.Printf("Starting server on %s", serverURL)
		log.Printf("OS: %s, Arch: %s", runtime.GOOS, runtime.GOARCH)

		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait a moment for server to start, then open browser
	time.Sleep(2 * time.Second)

	// Open default browser (Android will use Chrome/Browser)
	log.Printf("Opening browser to %s", serverURL)
	openBrowser(serverURL)
}

func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "android":
		// Android uses am (activity manager) to start browser
		cmd = exec.Command("am", "start", "-a", "android.intent.action.VIEW", "-d", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		log.Printf("Cannot open browser on %s", runtime.GOOS)
		return
	}

	if err := cmd.Run(); err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <title>Pejelagarto Translator</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
        }
        .container {
            width: 100%;
            max-width: 600px;
            background: white;
            border-radius: 20px;
            padding: 30px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
        }
        h1 {
            color: #667eea;
            text-align: center;
            margin-bottom: 10px;
            font-size: 28px;
        }
        .subtitle {
            text-align: center;
            color: #999;
            font-size: 14px;
            margin-bottom: 30px;
        }
        .input-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #666;
            font-weight: 600;
        }
        textarea {
            width: 100%;
            padding: 15px;
            border: 2px solid #e0e0e0;
            border-radius: 10px;
            font-size: 16px;
            resize: vertical;
            min-height: 120px;
            font-family: inherit;
        }
        textarea:focus {
            outline: none;
            border-color: #667eea;
        }
        .button-group {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
        }
        button {
            flex: 1;
            padding: 15px;
            border: none;
            border-radius: 10px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s;
        }
        .btn-primary {
            background: #667eea;
            color: white;
        }
        .btn-primary:active {
            background: #5568d3;
            transform: scale(0.98);
        }
        .btn-secondary {
            background: #f0f0f0;
            color: #333;
        }
        .btn-secondary:active {
            background: #e0e0e0;
            transform: scale(0.98);
        }
        .output {
            background: #f9f9f9;
            padding: 15px;
            border-radius: 10px;
            min-height: 120px;
            white-space: pre-wrap;
            word-wrap: break-word;
            border: 2px solid #e0e0e0;
            font-size: 16px;
        }
        .swap-btn {
            width: 100%;
            margin: 10px 0;
            background: #764ba2;
        }
        .note {
            text-align: center;
            color: #999;
            font-size: 12px;
            margin-top: 20px;
            padding: 10px;
            background: #f9f9f9;
            border-radius: 8px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🦎 Pejelagarto Translator</h1>
        <div class="subtitle">Android App - Running locally on your device</div>
        
        <div class="input-group">
            <label id="inputLabel">English:</label>
            <textarea id="inputText" placeholder="Type your text here..."></textarea>
        </div>
        
        <div class="button-group">
            <button class="btn-primary" onclick="translate()">Translate</button>
            <button class="btn-secondary" onclick="clearText()">Clear</button>
        </div>
        
        <button class="btn-primary swap-btn" onclick="swapDirection()">⇅ Swap Direction</button>
        
        <div class="input-group">
            <label id="outputLabel">Pejelagarto:</label>
            <div class="output" id="output">Translation will appear here...</div>
        </div>
        
        <div class="note">
            📱 Running on your device • No internet required
        </div>
    </div>

    <script>
        let direction = 'toPejelagarto';

        function translate() {
            const input = document.getElementById('inputText').value;
            if (!input.trim()) {
                document.getElementById('output').textContent = 'Please enter some text to translate.';
                return;
            }

            fetch('/translate', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ text: input, direction: direction })
            })
            .then(response => response.json())
            .then(data => {
                document.getElementById('output').textContent = data.result;
            })
            .catch(error => {
                document.getElementById('output').textContent = 'Error: ' + error.message;
            });
        }

        function clearText() {
            document.getElementById('inputText').value = '';
            document.getElementById('output').textContent = 'Translation will appear here...';
        }

        function swapDirection() {
            if (direction === 'toPejelagarto') {
                direction = 'fromPejelagarto';
                document.getElementById('inputLabel').textContent = 'Pejelagarto:';
                document.getElementById('outputLabel').textContent = 'English:';
            } else {
                direction = 'toPejelagarto';
                document.getElementById('inputLabel').textContent = 'English:';
                document.getElementById('outputLabel').textContent = 'Pejelagarto:';
            }
            clearText();
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(tmpl))
}

type TranslateRequest struct {
	Text      string `json:"text"`
	Direction string `json:"direction"`
}

type TranslateResponse struct {
	Result string `json:"result"`
}

func handleTranslate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TranslateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var result string
	if req.Direction == "toPejelagarto" {
		result = translator.TranslateToPejelagarto(req.Text)
	} else {
		result = translator.TranslateFromPejelagarto(req.Text)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TranslateResponse{Result: result})
}
